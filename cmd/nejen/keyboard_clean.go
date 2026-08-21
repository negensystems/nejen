package main

// Cleaning mode: the keyboard goes off so it can be wiped, and comes back
// on by itself.
//
// Turning a keyboard off on a machine you drive with a keyboard is only
// safe if every path leads back. There are four, and they do not depend on
// each other:
//
//	1. the deadline the supervisor sleeps on (always set, always capped),
//	2. `nejen keyboard clean --stop`, from the hub or the bar or a shell,
//	3. a middle click, bound inside the submap itself,
//	4. `nejen keyboard status`, which reaps a session whose supervisor died.
//
// Hyprland reloading its config is a fifth, accidental one: it drops back
// to the default submap. Every failure mode we could think of ends with a
// working keyboard.

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

func init() {
	registerCommand("keyboard clean", runKeyboardClean)
	registerCommand("keyboard status", runKeyboardStatus)
}

const (
	// cleanSubmap must match the submap name in config/hypr/core/cleaning.conf.
	cleanSubmap = "nejen-clean"
	// cleanClass is the app-id of the full-screen overlay, and the handle
	// used to close it again.
	cleanClass = "org.nejen.cleaning"
	// cleanSignal refreshes the waybar indicator, alongside RTMIN+8/9/10
	// already spoken for by screenrecord, idle and dnd.
	cleanSignal = "-RTMIN+11"

	cleanDefault = 30 * time.Second
	cleanMin     = 5 * time.Second
	// A wipe is a minute of work. The cap is what keeps a fat-fingered
	// `--for 10h` from being a problem in the first place.
	cleanMax = 5 * time.Minute
)

func dieClean(msg string) {
	fmt.Fprintf(os.Stderr, "nejen keyboard clean: %s\n", msg)
	os.Exit(1)
}

// parseCleanDuration reads the argument to --for. A bare number is seconds,
// so `--for 45` and `--for 45s` and `--for 1m30s` all work, and the result
// is clamped into [cleanMin, cleanMax] rather than rejected: an out-of-range
// duration is a typo, and refusing to clean is a worse answer than cleaning
// for as long as we are willing to.
func parseCleanDuration(s string) (time.Duration, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, fmt.Errorf("empty duration")
	}
	if n, err := strconv.Atoi(s); err == nil {
		return clampCleanDuration(time.Duration(n) * time.Second), nil
	}
	d, err := time.ParseDuration(s)
	if err != nil {
		return 0, fmt.Errorf("bad duration %q (try 30, 45s, 2m)", s)
	}
	return clampCleanDuration(d), nil
}

func clampCleanDuration(d time.Duration) time.Duration {
	if d < cleanMin {
		return cleanMin
	}
	if d > cleanMax {
		return cleanMax
	}
	return d
}

// cleanStatePath is where a running session records itself: one line of
// "<pid> <deadline unix seconds>". Both halves matter -- the pid says who
// to stop, the deadline says when the session is stale even if that pid is
// somehow still alive.
func cleanStatePath() string {
	stateHome := os.Getenv("XDG_STATE_HOME")
	if stateHome == "" {
		home, _ := os.UserHomeDir()
		stateHome = filepath.Join(home, ".local", "state")
	}
	return filepath.Join(stateHome, "nejen", "keyboard-clean")
}

func writeCleanState(pid int, deadline time.Time) error {
	path := cleanStatePath()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	line := fmt.Sprintf("%d %d\n", pid, deadline.Unix())
	return os.WriteFile(path, []byte(line), 0644)
}

// readCleanState returns the recorded session. ok is false when there is no
// state file or it cannot be read as one.
func readCleanState() (pid int, deadline time.Time, ok bool) {
	data, err := os.ReadFile(cleanStatePath())
	if err != nil {
		return 0, time.Time{}, false
	}
	fields := strings.Fields(string(data))
	if len(fields) != 2 {
		return 0, time.Time{}, false
	}
	pid, err = strconv.Atoi(fields[0])
	if err != nil {
		return 0, time.Time{}, false
	}
	sec, err := strconv.ParseInt(fields[1], 10, 64)
	if err != nil {
		return 0, time.Time{}, false
	}
	return pid, time.Unix(sec, 0), true
}

func processAlive(pid int) bool {
	if pid <= 0 {
		return false
	}
	proc, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	return proc.Signal(syscall.Signal(0)) == nil
}

// cleanLive reports a session that is genuinely still running: a live
// supervisor and a deadline that has not passed. Anything else is wreckage
// for the caller to clear.
func cleanLive() (pid int, deadline time.Time, live bool) {
	pid, deadline, ok := readCleanState()
	if !ok {
		return 0, time.Time{}, false
	}
	if !processAlive(pid) || !time.Now().Before(deadline) {
		return pid, deadline, false
	}
	return pid, deadline, true
}

func enterCleanSubmap() {
	exec.Command("hyprctl", "dispatch", "submap", cleanSubmap).Run()
}

// leaveCleanSubmap is written to be safe to call when cleaning mode is not
// running: resetting the submap that is already default is a no-op, and so
// is closing an overlay that is not open.
func leaveCleanSubmap() {
	exec.Command("hyprctl", "dispatch", "submap", "reset").Run()
	closeCleanOverlay()
	os.Remove(cleanStatePath())
	exec.Command("pkill", cleanSignal, "waybar").Run()
}

// cleanOverlayWindows finds the overlay by app-id, and returns the handles
// Hyprland gives for it rather than the app-id itself.
func cleanOverlayWindows() (addrs []string, pids []int) {
	out, err := exec.Command("hyprctl", "clients", "-j").Output()
	if err != nil {
		return nil, nil
	}
	var clients []struct {
		Address string `json:"address"`
		Class   string `json:"class"`
		PID     int    `json:"pid"`
	}
	if json.Unmarshal(out, &clients) != nil {
		return nil, nil
	}
	for _, c := range clients {
		if c.Class != cleanClass {
			continue
		}
		if c.Address != "" {
			addrs = append(addrs, c.Address)
		}
		if c.PID > 0 {
			pids = append(pids, c.PID)
		}
	}
	return addrs, pids
}

// closeCleanOverlay shuts the overlay down by address, then by pid if the
// terminal ignored the close request.
//
// Explicitly not `pkill -f org.nejen.cleaning`: that matches every command
// line merely mentioning the app-id -- the shell that started cleaning mode,
// an editor with this file open, a test harness watching for the window --
// and kills those too.
func closeCleanOverlay() {
	addrs, pids := cleanOverlayWindows()
	for _, addr := range addrs {
		exec.Command("hyprctl", "dispatch", "closewindow", "address:"+addr).Run()
	}
	for _, pid := range pids {
		for i := 0; i < 10 && processAlive(pid); i++ {
			time.Sleep(50 * time.Millisecond)
		}
		if processAlive(pid) {
			syscall.Kill(pid, syscall.SIGTERM)
		}
	}
}

func runKeyboardClean(args []string) {
	duration := cleanDefault
	stop := false
	supervise := false
	overlayUntil := int64(0)

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--stop", "-s":
			stop = true
		case "--for", "-f":
			if i+1 >= len(args) {
				dieClean("--for needs a duration (try 30, 45s, 2m)")
			}
			i++
			d, err := parseCleanDuration(args[i])
			if err != nil {
				dieClean(err.Error())
			}
			duration = d
		// --supervise and --overlay are how this command re-enters itself;
		// they are not part of the interface anyone is meant to type.
		case "--supervise":
			supervise = true
		case "--overlay":
			if i+1 >= len(args) {
				dieClean("--overlay needs a deadline")
			}
			i++
			sec, err := strconv.ParseInt(args[i], 10, 64)
			if err != nil {
				dieClean("bad overlay deadline")
			}
			overlayUntil = sec
		case "--help", "-h":
			fmt.Println("Usage: nejen keyboard clean [--for <duration>] [--stop]")
			fmt.Println()
			fmt.Println("Turn the keyboard off so it can be wiped, then turn it back on.")
			fmt.Println()
			fmt.Printf("  --for <duration>  how long to stay off (default %s, max %s)\n",
				humanCleanDuration(cleanDefault), humanCleanDuration(cleanMax))
			fmt.Println("  --stop            end cleaning mode now")
			fmt.Println()
			fmt.Println("While cleaning, a middle click or the bar indicator also ends it.")
			return
		default:
			if strings.HasPrefix(args[i], "-") {
				dieClean(fmt.Sprintf("unknown option %q", args[i]))
			}
			// A bare duration reads well enough to accept: `nejen keyboard
			// clean 2m`.
			d, err := parseCleanDuration(args[i])
			if err != nil {
				dieClean(err.Error())
			}
			duration = d
		}
	}

	if overlayUntil != 0 {
		runCleanOverlay(time.Unix(overlayUntil, 0))
		return
	}
	if stop {
		stopClean()
		return
	}
	if supervise {
		superviseClean(duration)
		return
	}

	// Asking for cleaning mode while it is already running means "I am
	// done" -- the bar indicator and the hub entry both land here, and
	// they are the two things you can still reach with a pointer.
	if _, _, live := cleanLive(); live {
		stopClean()
		return
	}

	// Clear any wreckage before laying down our own state, so a session
	// whose supervisor was killed cannot leave a stale pid behind that a
	// later --stop would try to signal.
	os.Remove(cleanStatePath())

	// The work happens in a detached copy of ourselves: whatever launched
	// this (a keybind, the hub, a shell) gets its prompt back immediately,
	// and the supervisor outlives it.
	cmd := exec.Command(nejenSelf(), "keyboard", "clean", "--supervise", "--for", strconv.Itoa(int(duration.Seconds())))
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	if err := cmd.Start(); err != nil {
		dieClean(fmt.Sprintf("failed to start: %v", err))
	}
	// Release the child rather than leaving it as a zombie on our exit.
	go cmd.Wait()
}

// stopClean ends a running session. It signals the supervisor so the
// supervisor's own restore path runs, then restores unconditionally: if the
// supervisor was already gone, this call is the one that gives the keyboard
// back.
func stopClean() {
	pid, _, ok := readCleanState()
	if ok && processAlive(pid) {
		syscall.Kill(pid, syscall.SIGTERM)
		// Give the supervisor a moment to tear down its own overlay before
		// we do it underneath it.
		for i := 0; i < 20 && processAlive(pid); i++ {
			time.Sleep(50 * time.Millisecond)
		}
	}
	leaveCleanSubmap()
}

// superviseClean is the session itself: it holds the submap open for the
// duration and is responsible for putting everything back.
func superviseClean(duration time.Duration) {
	deadline := time.Now().Add(duration)

	if err := writeCleanState(os.Getpid(), deadline); err != nil {
		dieClean(fmt.Sprintf("failed to record state: %v", err))
	}

	// A SIGTERM landing exactly on the deadline would otherwise run the
	// teardown twice, from the handler and from the main path at once.
	var once sync.Once
	restore := func() {
		once.Do(func() {
			leaveCleanSubmap()
			exec.Command("notify-send", "-t", "2500", "󰌌    Keyboard back on").Run()
		})
	}

	// A signal is the normal way this ends early, so restoring from the
	// handler is the main path out, not a safety net.
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGQUIT)
	go func() {
		<-sigCh
		restore()
		os.Exit(0)
	}()

	// Say it before the keys die, so the notification is on screen while
	// the user still has a chance to read it.
	exec.Command("notify-send", "-t", "2000", "󰌐    Cleaning mode",
		fmt.Sprintf("Keyboard off for %s. Middle click to finish early.", humanCleanDuration(duration))).Run()

	enterCleanSubmap()
	exec.Command("pkill", cleanSignal, "waybar").Run()
	openCleanOverlay(deadline)

	time.Sleep(time.Until(deadline))
	restore()
}

func humanCleanDuration(d time.Duration) string {
	d = d.Round(time.Second)
	if d < time.Minute {
		return fmt.Sprintf("%d seconds", int(d.Seconds()))
	}
	mins := int(d / time.Minute)
	secs := int((d % time.Minute) / time.Second)
	if secs == 0 {
		if mins == 1 {
			return "1 minute"
		}
		return fmt.Sprintf("%d minutes", mins)
	}
	return fmt.Sprintf("%dm%02ds", mins, secs)
}

// openCleanOverlay puts a full-screen window over the desktop for the
// duration. It is what makes cleaning mode legible -- a countdown you can
// read from arm's length -- and it also catches the stray pointer taps a
// cloth produces, which would otherwise land in whatever was on screen.
//
// Best effort by design: the overlay needs a terminal we know how to give a
// black canvas and an app-id, and cleaning mode still works without one.
func openCleanOverlay(deadline time.Time) {
	nejenPath := getNejenPath()
	self := fmt.Sprintf("nejen keyboard clean --overlay %d", deadline.Unix())

	out, _ := exec.Command("xdg-terminal-exec", "--print-id").Output()
	terminalID := string(out)

	var launch string
	switch {
	case strings.Contains(terminalID, "Alacritty"):
		opts := []string{"--class=" + cleanClass}
		cfg := filepath.Join(nejenPath, "config", "cleaning", "alacritty.toml")
		if _, err := os.Stat(cfg); err == nil {
			opts = append(opts, "--config-file", cfg)
		}
		opts = append(opts, "-e", self)
		launch = "alacritty " + strings.Join(opts, " ")
	case strings.Contains(terminalID, "ghostty"):
		opts := []string{"--class=" + cleanClass, "--font-size=22"}
		cfg := filepath.Join(nejenPath, "config", "cleaning", "ghostty")
		if _, err := os.Stat(cfg); err == nil {
			opts = append(opts, fmt.Sprintf("--config-file=%s", cfg))
		}
		opts = append(opts, "-e", self)
		launch = "ghostty " + strings.Join(opts, " ")
	case strings.Contains(terminalID, "kitty"):
		launch = fmt.Sprintf("kitty --class=%s --override font_size=22 --override window_padding_width=0 -e %s", cleanClass, self)
	default:
		return
	}

	exec.Command("hyprctl", "dispatch", "exec", "--", launch).Run()
}

// runCleanOverlay draws the countdown inside the overlay window. It ends on
// its own deadline as well as on the supervisor closing it, so an overlay
// that somehow outlives its session still goes away.
func runCleanOverlay(deadline time.Time) {
	fmt.Print("\033[?25l")       // hide the cursor
	defer fmt.Print("\033[?25h") // and give it back

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGQUIT)
	go func() {
		<-sigCh
		fmt.Print("\033[?25h\033[2J\033[H")
		os.Exit(0)
	}()

	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()

	for {
		left := time.Until(deadline)
		if left <= 0 {
			break
		}
		drawCleanOverlay(left)
		<-ticker.C
	}
	fmt.Print("\033[2J\033[H")
}

// cleanOverlayLines is the overlay's text, split out from the drawing so the
// layout can be checked without a terminal.
func cleanOverlayLines(left time.Duration) []string {
	secs := int(left.Round(time.Second) / time.Second)
	if secs < 0 {
		secs = 0
	}
	return []string{
		"󰌐",
		"",
		"C L E A N I N G   M O D E",
		"",
		"The keyboard is off. Wipe away.",
		"",
		fmt.Sprintf("%d:%02d", secs/60, secs%60),
		"",
		"Middle click to finish early",
	}
}

func drawCleanOverlay(left time.Duration) {
	cols, rows := getTerminalSize()

	lines := cleanOverlayLines(left)
	top := (rows - len(lines)) / 2
	if top < 0 {
		top = 0
	}

	var b strings.Builder
	b.WriteString("\033[2J\033[H")
	for i := 0; i < top; i++ {
		b.WriteString("\n")
	}
	for i, line := range lines {
		// The glyphs here are all single-width apart from the leading
		// icon, so rune count is a good enough centering measure.
		pad := (cols - len([]rune(line))) / 2
		if pad < 0 {
			pad = 0
		}
		// Dim everything except the icon and the count, so the two things
		// worth reading across the room are the two that stand out.
		emphasis := i == 0 || i == 6
		if emphasis {
			b.WriteString("\033[1m")
		} else {
			b.WriteString("\033[2m")
		}
		b.WriteString(strings.Repeat(" ", pad))
		b.WriteString(line)
		b.WriteString("\033[0m\n")
	}
	fmt.Print(b.String())
}

// runKeyboardStatus feeds the waybar indicator, and doubles as the watchdog
// that clears a session whose supervisor died: waybar polls it, so the bar
// is what notices a keyboard that should have come back and did not.
func runKeyboardStatus(args []string) {
	pid, deadline, ok := readCleanState()
	if !ok {
		fmt.Println(`{"text": ""}`)
		return
	}

	if !processAlive(pid) || !time.Now().Before(deadline) {
		leaveCleanSubmap()
		fmt.Println(`{"text": ""}`)
		return
	}

	left := int(time.Until(deadline).Round(time.Second) / time.Second)
	status := map[string]string{
		"text":    "󰌐",
		"tooltip": fmt.Sprintf("Keyboard off for cleaning - %ds left, click to finish", left),
		"class":   "active",
	}
	out, err := json.Marshal(status)
	if err != nil {
		fmt.Println(`{"text": ""}`)
		return
	}
	fmt.Println(string(out))
}
