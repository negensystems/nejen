package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseChord(t *testing.T) {
	cases := []struct {
		chord    string
		wantMods []string
		wantKey  string
	}{
		// Modifiers are normalised and sorted, so the same chord written in
		// any order collapses to one identity.
		{"super+shift+w", []string{"shift", "super"}, "w"},
		{"shift+super+w", []string{"shift", "super"}, "w"},
		{"SUPER+W", []string{"super"}, "W"},
		{"ctrl+alt+delete", []string{"alt", "ctrl"}, "delete"},
		{"w", nil, "w"},
		{"  super+w  ", []string{"super"}, "w"},
		{"super+XF86AudioPlay", []string{"super"}, "XF86AudioPlay"},
	}
	for _, c := range cases {
		mods, key := parseChord(c.chord)
		if key != c.wantKey {
			t.Errorf("parseChord(%q) key = %q, want %q", c.chord, key, c.wantKey)
		}
		if strings.Join(mods, "+") != strings.Join(c.wantMods, "+") {
			t.Errorf("parseChord(%q) mods = %v, want %v", c.chord, mods, c.wantMods)
		}
	}
}

func TestChordIDIsOrderAndCaseInsensitive(t *testing.T) {
	if a, b := chordID("super+shift+w"), chordID("SHIFT+Super+W"); a != b {
		t.Errorf("chordID mismatch: %q vs %q", a, b)
	}
}

func TestHyprKeyAliases(t *testing.T) {
	cases := map[string]string{
		"enter": "RETURN", "return": "RETURN", "esc": "ESCAPE",
		"/": "slash", ".": "period", "-": "minus",
		// Unknown keys pass through with their original case, which is what
		// XF86 media keys depend on.
		"XF86AudioPlay": "XF86AudioPlay", "w": "w",
	}
	for in, want := range cases {
		if got := hyprKey(in); got != want {
			t.Errorf("hyprKey(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestBindLine(t *testing.T) {
	cases := []struct {
		name string
		bind Bind
		want string
	}{
		{"exec", Bind{Keys: "super+return", Desc: "Terminal", Run: "alacritty"},
			`hl.bind("SUPER + RETURN", hl.dsp.exec_cmd("alacritty"), { desc = "Terminal" })`},
		{"comma in desc", Bind{Keys: "super+w", Desc: "Close, now", Run: "x"},
			`hl.bind("SUPER + w", hl.dsp.exec_cmd("x"), { desc = "Close, now" })`},
		{"dispatcher with args", Bind{Keys: "super+h", Desc: "Focus left", Action: "movefocus l"},
			`hl.bind("SUPER + h", hl.dsp.focus({ direction = "left" }), { desc = "Focus left" })`},
		{"move window direction", Bind{Keys: "super+shift+l", Desc: "Move right", Action: "movewindow r"},
			`hl.bind("SHIFT + SUPER + l", hl.dsp.window.move({ direction = "right" }), { desc = "Move right" })`},
		{"resize active", Bind{Keys: "super+ctrl+h", Desc: "Shrink width", Action: "resizeactive -60 0", Flags: "e"},
			`hl.bind("CTRL + SUPER + h", hl.dsp.window.resize({ x = -60, y = 0, relative = true }), { desc = "Shrink width", repeating = true })`},
		{"bare dispatcher", Bind{Keys: "super+q", Desc: "Close", Action: "killactive"},
			`hl.bind("SUPER + q", hl.dsp.window.close(), { desc = "Close" })`},
		{"fullscreen modes", Bind{Keys: "super+shift+f", Desc: "Maximize", Action: "fullscreen 1"},
			`hl.bind("SHIFT + SUPER + f", hl.dsp.window.fullscreen({ mode = "maximized" }), { desc = "Maximize" })`},
		{"workspace numeric", Bind{Keys: "super+1", Desc: "Workspace 1", Action: "workspace 1"},
			`hl.bind("SUPER + 1", hl.dsp.focus({ workspace = 1 }), { desc = "Workspace 1" })`},
		{"workspace named", Bind{Keys: "super+tab", Desc: "Previous workspace", Action: "workspace previous"},
			`hl.bind("SUPER + TAB", hl.dsp.focus({ workspace = "previous" }), { desc = "Previous workspace" })`},
		{"flags", Bind{Keys: "XF86PowerOff", Desc: "Session", Run: "nejen hub session", Flags: "l"},
			`hl.bind("XF86PowerOff", hl.dsp.exec_cmd("nejen hub session"), { desc = "Session", locked = true })`},
		{"mouse bind", Bind{Keys: "super+mouse:272", Desc: "Move", Action: "movewindow", Flags: "m"},
			`hl.bind("SUPER + mouse:272", hl.dsp.window.drag(), { desc = "Move", mouse = true })`},
		{"raw lua dispatcher", Bind{Keys: "super+g", Desc: "Custom", Action: `hl.dsp.layout("togglesplit")`},
			`hl.bind("SUPER + g", hl.dsp.layout("togglesplit"), { desc = "Custom" })`},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := bindLine(c.bind); got != c.want {
				t.Errorf("bindLine = %q, want %q", got, c.want)
			}
		})
	}
}

func TestPrettyChord(t *testing.T) {
	cases := []struct {
		bind Bind
		want string
	}{
		{Bind{Keys: "super+shift+w"}, "Shift + Super + W"},
		{Bind{Keys: "super+return"}, "Super + Return"},
		{Bind{Keys: "XF86AudioPlay"}, "XF86AudioPlay"},
	}
	for _, c := range cases {
		if got := prettyChord(c.bind); got != c.want {
			t.Errorf("prettyChord(%q) = %q, want %q", c.bind.Keys, got, c.want)
		}
	}
}

// writeKeymapLayers stages a default layer and a user layer on disk and points
// the resolver at them.
func writeKeymapLayers(t *testing.T, defaults, user string) {
	t.Helper()
	dir := t.TempDir()
	share := filepath.Join(dir, "share", "config")
	conf := filepath.Join(dir, "config", "nejen")
	for _, d := range []string{share, conf} {
		if err := os.MkdirAll(d, 0755); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(share, "keymap.toml"), []byte(defaults), 0644); err != nil {
		t.Fatal(err)
	}
	if user != "" {
		if err := os.WriteFile(filepath.Join(conf, "keymap.toml"), []byte(user), 0644); err != nil {
			t.Fatal(err)
		}
	}
	t.Setenv("NEJEN_PATH", filepath.Join(dir, "share"))
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(dir, "config"))
}

// Unbinding a default and then binding the same chord is the documented way to
// replace a keybinding. It used to emit the binding twice: the unbind removed
// the chord from the lookup table but not from the ordering list, so the user's
// bind appended a second entry pointing at the same record.
func TestMergedKeymapUnbindThenRebindEmitsOneBind(t *testing.T) {
	writeKeymapLayers(t,
		`[[bind]]
keys = "super+w"
desc = "Default close"
run = "nejen win close"
`,
		`unbind = ["super+w"]

[[bind]]
keys = "super+w"
desc = "My own action"
run = "echo mine"
`)
	binds := mergedKeymap()
	if len(binds) != 1 {
		t.Fatalf("got %d binds, want 1: %+v", len(binds), binds)
	}
	if binds[0].Desc != "My own action" {
		t.Errorf("user bind did not win: %q", binds[0].Desc)
	}
}

func TestMergedKeymapUnbindRemovesDefault(t *testing.T) {
	writeKeymapLayers(t,
		`[[bind]]
keys = "super+w"
desc = "Default close"
run = "x"

[[bind]]
keys = "super+q"
desc = "Quit"
run = "y"
`, `unbind = ["super+w"]`)
	binds := mergedKeymap()
	if len(binds) != 1 || binds[0].Keys != "super+q" {
		t.Fatalf("unbind did not remove the default: %+v", binds)
	}
}

// The user layer overrides a default in place, and an unbind written in any
// modifier order still matches.
func TestMergedKeymapOverrideAndOrderInsensitiveUnbind(t *testing.T) {
	writeKeymapLayers(t,
		`[[bind]]
keys = "super+shift+w"
desc = "Default"
run = "x"
`, `unbind = ["SHIFT+Super+W"]`)
	if binds := mergedKeymap(); len(binds) != 0 {
		t.Fatalf("unbind should match regardless of order/case: %+v", binds)
	}
}

func TestMergedKeymapUserAddsNewBind(t *testing.T) {
	writeKeymapLayers(t,
		`[[bind]]
keys = "super+q"
desc = "Quit"
run = "y"
`,
		`[[bind]]
keys = "super+n"
desc = "New"
run = "z"
`)
	binds := mergedKeymap()
	if len(binds) != 2 {
		t.Fatalf("got %d binds, want 2: %+v", len(binds), binds)
	}
}
