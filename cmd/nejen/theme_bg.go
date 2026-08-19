package main

// Shared plumbing for the `nejen theme bg` family: where wallpapers live,
// which one is active, and how to swap it without flashing the desktop.

import (
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"time"
)

// Extensions swaybg can render. Anything else in a wallpaper directory is
// ignored rather than offered and then failing at display time.
var bgImageExt = map[string]bool{
	".png":  true,
	".jpg":  true,
	".jpeg": true,
	".webp": true,
	".gif":  true,
	".bmp":  true,
	".tga":  true,
}

// nejenState is ~/.local/state/nejen, where the active theme name and the
// background symlink live.
func nejenState() string {
	stateHome := os.Getenv("XDG_STATE_HOME")
	if stateHome == "" {
		home, _ := os.UserHomeDir()
		stateHome = filepath.Join(home, ".local", "state")
	}
	return filepath.Join(stateHome, "nejen")
}

// bgLink is the stable path every consumer points at (swaybg, hyprlock).
// Setting a wallpaper repoints this symlink; nothing else needs to know
// where the image actually came from.
func bgLink() string {
	return filepath.Join(nejenState(), "background")
}

// bgUserDir is the theme-independent drop folder for the user's own
// wallpapers. It outranks both theme layers and survives `nejen theme set`,
// so a personal collection is not lost when themes change.
func bgUserDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "nejen", "backgrounds")
}

// activeThemeName reads the theme recorded by `nejen theme set`, or "".
func activeThemeName() string {
	b, err := os.ReadFile(filepath.Join(nejenState(), "theme.name"))
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(b))
}

// bgDirs lists wallpaper directories in precedence order: the user's own
// folder, then the active theme's user override, then the shipped theme.
func bgDirs() []string {
	home, _ := os.UserHomeDir()
	dirs := []string{bgUserDir()}
	if theme := activeThemeName(); theme != "" {
		dirs = append(dirs,
			filepath.Join(home, ".config", "nejen", "themes", theme, "backgrounds"),
			filepath.Join(getNejenPath(), "themes", theme, "backgrounds"),
		)
	}
	return dirs
}

// bgList collects every wallpaper on offer, sorted by filename. A basename
// seen in an earlier layer shadows the same name later on, so a user file
// can replace a shipped one without being renamed.
func bgList() []string {
	claimed := map[string]bool{}
	var paths []string

	for _, dir := range bgDirs() {
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, e := range entries {
			name := e.Name()
			if claimed[name] || !bgImageExt[strings.ToLower(filepath.Ext(name))] {
				continue
			}
			// ReadDir does not follow symlinks; stat so dangling links and
			// symlinked directories do not reach the picker.
			full := filepath.Join(dir, name)
			if info, err := os.Stat(full); err != nil || info.IsDir() {
				continue
			}
			claimed[name] = true
			paths = append(paths, full)
		}
	}

	sort.Slice(paths, func(i, j int) bool {
		bi, bj := filepath.Base(paths[i]), filepath.Base(paths[j])
		if bi != bj {
			return bi < bj
		}
		return paths[i] < paths[j]
	})
	return paths
}

// currentBg resolves the state symlink to the image actually in use, or "".
func currentBg() string {
	target, err := filepath.EvalSymlinks(bgLink())
	if err != nil {
		return ""
	}
	return target
}

// bgIndexOf locates current within list, or -1. Comparison is by inode
// rather than by string so a wallpaper still matches when it is reached
// through a symlink or a differently spelled path.
func bgIndexOf(list []string, current string) int {
	if current == "" {
		return -1
	}
	curInfo, err := os.Stat(current)
	if err != nil {
		return -1
	}
	for i, p := range list {
		if info, err := os.Stat(p); err == nil && os.SameFile(curInfo, info) {
			return i
		}
	}
	return -1
}

// bgIsUserOwned reports whether path came from the user's own wallpaper
// folder rather than from a theme. Callers use it to leave a deliberate
// choice alone.
func bgIsUserOwned(path string) bool {
	if path == "" {
		return false
	}
	dir, err := filepath.EvalSymlinks(bgUserDir())
	if err != nil {
		return false
	}
	rel, err := filepath.Rel(dir, path)
	if err != nil {
		return false
	}
	return rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator))
}

// bgLabel turns "10-misty-ridge.jpg" into "Misty Ridge" for display.
func bgLabel(path string) string {
	stem := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	stem = strings.TrimLeft(stem, "0123456789")
	stem = strings.TrimLeft(stem, "-_ ")
	stem = strings.NewReplacer("-", " ", "_", " ").Replace(stem)

	words := strings.Fields(stem)
	for i, w := range words {
		words[i] = strings.ToUpper(w[:1]) + strings.ToLower(w[1:])
	}
	if len(words) == 0 {
		return filepath.Base(path)
	}
	return strings.Join(words, " ")
}

// applyBg repoints the state symlink at path and reloads everything that
// renders the wallpaper.
func applyBg(path string) error {
	abs, err := filepath.Abs(path)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(nejenState(), 0755); err != nil {
		return err
	}

	link := bgLink()
	os.Remove(link) // mimic ln -sfn
	if err := os.Symlink(abs, link); err != nil {
		return err
	}

	reloadSwaybg()
	return nil
}

// reloadSwaybg brings up a swaybg on the state symlink and retires whatever
// was already running. The new instance starts *before* the old one is
// killed: doing it the other way round leaves the bare desktop visible for
// a frame, which is glaring when cycling wallpapers.
func reloadSwaybg() {
	if os.Getenv("WAYLAND_DISPLAY") == "" {
		return
	}

	stale := swaybgPids()

	var cmd *exec.Cmd
	if _, err := exec.LookPath("uwsm-app"); err == nil {
		cmd = exec.Command("uwsm-app", "--", "swaybg", "-i", bgLink(), "-m", "fill")
	} else {
		cmd = exec.Command("swaybg", "-i", bgLink(), "-m", "fill")
	}
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	if err := cmd.Start(); err != nil {
		return
	}

	// Long enough for the new instance to attach its layer surface.
	time.Sleep(400 * time.Millisecond)

	// Retire the old instances only once a new one is actually up. If the
	// replacement failed to start, the previous wallpaper stays on screen
	// rather than the desktop going bare.
	replaced := false
	for _, pid := range swaybgPids() {
		if !slices.Contains(stale, pid) {
			replaced = true
			break
		}
	}
	if !replaced {
		return
	}

	for _, pid := range stale {
		syscall.Kill(pid, syscall.SIGTERM)
	}
}

func swaybgPids() []int {
	out, err := exec.Command("pgrep", "-x", "swaybg").Output()
	if err != nil {
		return nil
	}
	var pids []int
	for _, f := range strings.Fields(string(out)) {
		if pid, err := strconv.Atoi(f); err == nil {
			pids = append(pids, pid)
		}
	}
	return pids
}
