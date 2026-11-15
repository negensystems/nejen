package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/BurntSushi/toml"
)

type Bind struct {
	Keys   string `toml:"keys"`
	Desc   string `toml:"desc"`
	Run    string `toml:"run"`
	Action string `toml:"action"`
	Group  string `toml:"group"`
	Flags  string `toml:"flags"`
}

type KeymapData struct {
	Bind   []Bind   `toml:"bind"`
	Unbind []string `toml:"unbind"`
}

var mods = map[string]bool{
	"super": true,
	"shift": true,
	"ctrl":  true,
	"alt":   true,
}

var keyAliases = map[string]string{
	"enter":     "RETURN",
	"return":    "RETURN",
	"space":     "SPACE",
	"tab":       "TAB",
	"escape":    "ESCAPE",
	"esc":       "ESCAPE",
	"backspace": "BACKSPACE",
	"delete":    "DELETE",
	"print":     "PRINT",
	"up":        "up",
	"down":      "down",
	"left":      "left",
	"right":     "right",
	"/":         "slash",
	".":         "period",
	",":         "comma",
	";":         "semicolon",
	"'":         "apostrophe",
	"-":         "minus",
	"=":         "equal",
	"[":         "bracketleft",
	"]":         "bracketright",
}

func init() {
	registerCommand("keymap render", runKeymapRender)
}

func dieRender(msg string) {
	fmt.Fprintf(os.Stderr, "nejen keymap render: %s\n", msg)
	os.Exit(1)
}

func parseChord(chord string) ([]string, string) {
	partsRaw := strings.Split(strings.TrimSpace(chord), "+")
	var parts []string
	for _, p := range partsRaw {
		if p != "" {
			parts = append(parts, p)
		}
	}
	if len(parts) == 0 {
		dieRender(fmt.Sprintf("empty chord in keys = %q", chord))
	}
	key := parts[len(parts)-1]
	modList := parts[:len(parts)-1]

	var lowerMods []string
	for _, m := range modList {
		lm := strings.ToLower(m)
		if !mods[lm] {
			dieRender(fmt.Sprintf("unknown modifier %q in keys = %q", m, chord))
		}
		lowerMods = append(lowerMods, lm)
	}
	sort.Strings(lowerMods)
	return lowerMods, key
}

func chordID(chord string) string {
	modsList, key := parseChord(chord)
	modsList = append(modsList, strings.ToLower(key))
	return strings.Join(modsList, "+")
}

func hyprKey(key string) string {
	lk := strings.ToLower(key)
	if val, ok := keyAliases[lk]; ok {
		return val
	}
	return key
}

func loadLayer(path string) ([]Bind, []string) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, nil
	}
	var data KeymapData
	if _, err := toml.DecodeFile(path, &data); err != nil {
		dieRender(fmt.Sprintf("%s: failed to parse toml: %v", path, err))
	}

	for _, b := range data.Bind {
		if b.Keys == "" || b.Desc == "" {
			dieRender(fmt.Sprintf("%s: a [[bind]] entry is missing 'keys' or 'desc'", path))
		}
		hasRun := b.Run != ""
		hasAction := b.Action != ""
		if hasRun == hasAction {
			dieRender(fmt.Sprintf("%s: bind %q needs exactly one of run/action", path, b.Keys))
		}
	}
	return data.Bind, data.Unbind
}

func mergedKeymap() []Bind {
	nejenPath := getNejenPath()
	configHome := os.Getenv("XDG_CONFIG_HOME")
	if configHome == "" {
		home, _ := os.UserHomeDir()
		configHome = filepath.Join(home, ".config")
	}

	defaultBinds, _ := loadLayer(filepath.Join(nejenPath, "config", "keymap.toml"))
	userBinds, unbinds := loadLayer(filepath.Join(configHome, "nejen", "keymap.toml"))

	table := make(map[string]Bind)
	// queued tracks what is already in `order`, which is not the same question
	// as what is currently in `table`: an unbind removes the chord from the
	// table but cannot remove it from `order`. Testing the table instead meant
	// the documented "unbind a default, then bind the chord yourself" flow
	// appended the chord a second time, and the merged keymap emitted the
	// binding twice, once per entry in `order`, both resolving to the user's
	// bind. Hyprland took the duplicate, and it showed up twice in the
	// cheat sheet.
	queued := make(map[string]bool)
	var order []string

	for _, b := range defaultBinds {
		cid := chordID(b.Keys)
		if queued[cid] {
			dieRender(fmt.Sprintf("duplicate default chord %q", b.Keys))
		}
		table[cid] = b
		queued[cid] = true
		order = append(order, cid)
	}

	for _, chord := range unbinds {
		delete(table, chordID(chord))
	}

	for _, b := range userBinds {
		cid := chordID(b.Keys)
		if !queued[cid] {
			order = append(order, cid)
			queued[cid] = true
		}
		table[cid] = b
	}

	var final []Bind
	for _, cid := range order {
		if b, exists := table[cid]; exists {
			final = append(final, b)
		}
	}
	return final
}

var bindFlagMap = []struct {
	flag rune
	opt  string
}{
	{'l', "locked"},
	{'e', "repeating"},
	{'r', "release"},
	{'m', "mouse"},
	{'n', "non_consuming"},
	{'t', "transparent"},
	{'i', "ignore_mods"},
	{'p', "dont_inhibit"},
	{'o', "long_press"},
	{'u', "submap_universal"},
	{'c', "click"},
	{'g', "drag"},
}

var directionAliases = map[string]string{
	"l": "left", "left": "left",
	"r": "right", "right": "right",
	"u": "up", "up": "up",
	"d": "down", "down": "down",
}

func luaWorkspaceArg(arg string) string {
	if n, err := strconv.Atoi(arg); err == nil {
		return strconv.Itoa(n)
	}
	return luaQuote(arg)
}

func luaDispatcher(b Bind) (string, error) {
	if b.Run != "" {
		return fmt.Sprintf("hl.dsp.exec_cmd(%s)", luaQuote(b.Run)), nil
	}

	action := strings.TrimSpace(b.Action)
	if strings.HasPrefix(action, "hl.") {
		return action, nil
	}

	if strings.Contains(b.Flags, "m") {
		switch action {
		case "movewindow":
			return "hl.dsp.window.drag()", nil
		case "resizewindow":
			return "hl.dsp.window.resize()", nil
		}
	}

	fields := strings.Fields(action)
	if len(fields) == 0 {
		return "", fmt.Errorf("empty action in bind %q", b.Keys)
	}

	switch fields[0] {
	case "killactive":
		return "hl.dsp.window.close()", nil
	case "togglefloating":
		return `hl.dsp.window.float({ action = "toggle" })`, nil
	case "fullscreen":
		if len(fields) > 1 && fields[1] == "1" {
			return `hl.dsp.window.fullscreen({ mode = "maximized" })`, nil
		}
		return "hl.dsp.window.fullscreen()", nil
	case "movefocus":
		if len(fields) != 2 {
			return "", fmt.Errorf("movefocus expects a direction, got %q", action)
		}
		dir, ok := directionAliases[strings.ToLower(fields[1])]
		if !ok {
			return "", fmt.Errorf("unknown direction %q in %q", fields[1], action)
		}
		return fmt.Sprintf("hl.dsp.focus({ direction = %s })", luaQuote(dir)), nil
	case "movewindow":
		if len(fields) != 2 {
			return "", fmt.Errorf("movewindow expects a direction, got %q", action)
		}
		dir, ok := directionAliases[strings.ToLower(fields[1])]
		if !ok {
			return "", fmt.Errorf("unknown direction %q in %q", fields[1], action)
		}
		return fmt.Sprintf("hl.dsp.window.move({ direction = %s })", luaQuote(dir)), nil
	case "resizeactive":
		if len(fields) != 3 {
			return "", fmt.Errorf("resizeactive expects dx dy, got %q", action)
		}
		return fmt.Sprintf("hl.dsp.window.resize({ x = %s, y = %s, relative = true })", luaNumberOrQuote(fields[1]), luaNumberOrQuote(fields[2])), nil
	case "workspace":
		if len(fields) != 2 {
			return "", fmt.Errorf("workspace expects a target, got %q", action)
		}
		return fmt.Sprintf("hl.dsp.focus({ workspace = %s })", luaWorkspaceArg(fields[1])), nil
	case "movetoworkspace":
		if len(fields) != 2 {
			return "", fmt.Errorf("movetoworkspace expects a target, got %q", action)
		}
		return fmt.Sprintf("hl.dsp.window.move({ workspace = %s })", luaWorkspaceArg(fields[1])), nil
	case "centerwindow":
		return "hl.dsp.window.center()", nil
	case "pin":
		return `hl.dsp.window.pin({ action = "toggle" })`, nil
	case "pseudo":
		return "hl.dsp.window.pseudo()", nil
	case "togglesplit", "swapsplit":
		return fmt.Sprintf("hl.dsp.layout(%s)", luaQuote(fields[0])), nil
	case "togglespecialworkspace":
		if len(fields) > 1 {
			return fmt.Sprintf("hl.dsp.workspace.toggle_special(%s)", luaQuote(fields[1])), nil
		}
		return "hl.dsp.workspace.toggle_special()", nil
	case "exit":
		return "hl.dsp.exit()", nil
	default:
		return "", fmt.Errorf("unsupported keymap action %q in %q", action, b.Keys)
	}
}

func bindLine(b Bind) string {
	modsList, key := parseChord(b.Keys)
	var upperMods []string
	for _, m := range modsList {
		upperMods = append(upperMods, strings.ToUpper(m))
	}
	chord := hyprKey(key)
	if len(upperMods) > 0 {
		chord = strings.Join(upperMods, " + ") + " + " + chord
	}

	dsp, err := luaDispatcher(b)
	if err != nil {
		dieRender(err.Error())
	}

	var opts []string
	if b.Desc != "" {
		opts = append(opts, "desc = "+luaQuote(b.Desc))
	}
	for _, entry := range bindFlagMap {
		if strings.ContainsRune(b.Flags, entry.flag) {
			opts = append(opts, entry.opt+" = true")
		}
	}

	if len(opts) == 0 {
		return fmt.Sprintf("hl.bind(%s, %s)", luaQuote(chord), dsp)
	}
	return fmt.Sprintf("hl.bind(%s, %s, { %s })", luaQuote(chord), dsp, strings.Join(opts, ", "))
}

func prettyChord(b Bind) string {
	modsList, key := parseChord(b.Keys)
	shown := map[string]string{
		"super": "Super",
		"shift": "Shift",
		"ctrl":  "Ctrl",
		"alt":   "Alt",
	}
	var out []string
	for _, m := range modsList {
		out = append(out, shown[m])
	}
	if strings.HasPrefix(key, "XF86") {
		out = append(out, key)
	} else if len(key) > 0 {
		out = append(out, strings.ToUpper(key[:1])+key[1:])
	}
	return strings.Join(out, " + ")
}

func renderKeymap(binds []Bind) string {
	lines := []string{
		"-- Generated by `nejen keymap render` from keymap.toml. Do not edit.",
		"-- Change bindings in ~/.config/nejen/keymap.toml and re-render.",
		"",
	}
	for _, b := range binds {
		lines = append(lines, bindLine(b))
	}
	return strings.Join(lines, "\n") + "\n"
}

func cheatsheet(binds []Bind) string {
	groups := make(map[string][]Bind)
	for _, b := range binds {
		g := b.Group
		if g == "" {
			g = "other"
		}
		groups[g] = append(groups[g], b)
	}

	var width int
	for _, b := range binds {
		l := len(prettyChord(b))
		if l > width {
			width = l
		}
	}

	var groupNames []string
	for group := range groups {
		groupNames = append(groupNames, group)
	}
	sort.Strings(groupNames)

	var out []string
	for _, group := range groupNames {
		entries := groups[group]
		out = append(out, strings.ToUpper(group))
		for _, b := range entries {
			pc := prettyChord(b)
			padding := strings.Repeat(" ", width-len(pc))
			out = append(out, fmt.Sprintf("  %s%s  %s", pc, padding, b.Desc))
		}
		out = append(out, "")
	}

	res := strings.Join(out, "\n")
	return strings.TrimRight(res, "\n") + "\n"
}

func runKeymapRender(args []string) {
	cheatsheetMode := false
	for _, arg := range args {
		if arg == "--cheatsheet" {
			cheatsheetMode = true
			break
		}
	}

	binds := mergedKeymap()
	if cheatsheetMode {
		fmt.Print(cheatsheet(binds))
		return
	}

	stateHome := os.Getenv("XDG_STATE_HOME")
	if stateHome == "" {
		home, _ := os.UserHomeDir()
		stateHome = filepath.Join(home, ".local", "state")
	}
	outDir := filepath.Join(stateHome, "nejen", "keymap")
	if err := os.MkdirAll(outDir, 0755); err != nil {
		dieRender(fmt.Sprintf("failed to create directory: %v", err))
	}

	outPath := filepath.Join(outDir, "hyprland.lua")
	tmpPath := outPath + ".tmp"
	if err := os.WriteFile(tmpPath, []byte(renderKeymap(binds)), 0644); err != nil {
		dieRender(fmt.Sprintf("failed to write tmp file: %v", err))
	}
	if err := os.Rename(tmpPath, outPath); err != nil {
		dieRender(fmt.Sprintf("failed to move tmp file: %v", err))
	}
	fmt.Printf("rendered %d binds -> %s\n", len(binds), outPath)
}
