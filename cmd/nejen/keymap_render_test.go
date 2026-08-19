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
			"bindd = SUPER, RETURN, Terminal, exec, alacritty"},
		// Commas in a description would be read as extra fields by hyprland.
		{"comma in desc", Bind{Keys: "super+w", Desc: "Close, now", Run: "x"},
			"bindd = SUPER, w, Close; now, exec, x"},
		{"dispatcher with args", Bind{Keys: "super+h", Desc: "Focus left", Action: "movefocus l"},
			"bindd = SUPER, h, Focus left, movefocus, l"},
		{"bare dispatcher", Bind{Keys: "super+q", Desc: "Close", Action: "killactive"},
			"bindd = SUPER, q, Close, killactive"},
		{"flags", Bind{Keys: "XF86PowerOff", Desc: "Session", Run: "nejen hub session", Flags: "l"},
			"bindld = , XF86PowerOff, Session, exec, nejen hub session"},
		{"mouse bind", Bind{Keys: "super+mouse:272", Desc: "Move", Action: "movewindow", Flags: "m"},
			"bindm = SUPER, mouse:272, movewindow"},
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
