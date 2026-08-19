package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
)

func getNejenPath() string {
	if path := os.Getenv("NEJEN_PATH"); path != "" {
		return path
	}
	home, _ := os.UserHomeDir()
	localShare := filepath.Join(home, ".local", "share", "nejen")
	if stat, err := os.Stat(localShare); err == nil && stat.IsDir() {
		return localShare
	}
	return "/usr/share/nejen"
}

// nejenSelf is the path to this binary, for subcommands that need to
// re-enter the dispatcher as a separate process. Falls back to argv[0],
// which resolves through PATH.
func nejenSelf() string {
	if exe, err := os.Executable(); err == nil {
		return exe
	}
	return os.Args[0]
}

func runVersion(nejenPath string, args []string) {
	if len(args) > 0 {
		switch args[0] {
		case "branch", "channel":
			cmd := exec.Command("git", "-C", nejenPath, "rev-parse", "--abbrev-ref", "HEAD")
			if out, err := cmd.Output(); err == nil {
				fmt.Println(strings.TrimSpace(string(out)))
			} else {
				fmt.Println("unknown")
			}
			return
		case "pkgs":
			out, err := exec.Command("bash", "-c", "grep upgraded /var/log/pacman.log | tail -1 | sed -E 's/\\[([^]]+)\\].*/\\1/'").Output()
			if err != nil {
				fmt.Println("unknown")
				return
			}
			dateStr := strings.TrimSpace(string(out))
			if dateStr == "" {
				fmt.Println("unknown")
				return
			}
			dateOut, err := exec.Command("date", "-d", dateStr, "+%A, %B %d %Y at %H:%M").Output()
			if err != nil {
				fmt.Println("unknown")
				return
			}
			fmt.Println(strings.TrimSpace(string(dateOut)))
			return
		}
	}

	versionFile := filepath.Join(nejenPath, "version")
	if data, err := os.ReadFile(versionFile); err == nil {
		fmt.Print(string(data))
		return
	}
	cmd := exec.Command("git", "-C", nejenPath, "describe", "--tags", "--always")
	if out, err := cmd.Output(); err == nil {
		fmt.Print(string(out))
		return
	}
	fmt.Println("unknown")
}

func runDoctor(nejenPath string) {
	passCount, warnCount, failCount := 0, 0, 0

	pass := func(msg string) {
		fmt.Printf("PASS  %s\n", msg)
		passCount++
	}
	warn := func(msg string) {
		fmt.Printf("WARN  %s\n", msg)
		warnCount++
	}
	fail := func(msg string) {
		fmt.Printf("FAIL  %s\n", msg)
		failCount++
	}

	if stat, err := os.Stat(nejenPath); err == nil && stat.IsDir() {
		pass("NEJEN_PATH resolves: " + nejenPath)
	} else {
		fail("NEJEN_PATH does not exist: " + nejenPath)
	}

	for _, dir := range []string{"config", "templates", "themes"} {
		dirPath := filepath.Join(nejenPath, dir)
		if stat, err := os.Stat(dirPath); err == nil && stat.IsDir() {
			pass("NEJEN_PATH contains " + dir + "/")
		} else {
			fail("NEJEN_PATH is missing " + dir + "/ (" + dirPath + ")")
		}
	}

	// Only a dev-mode tree keeps the dispatcher under NEJEN_PATH/bin, and only
	// there can it be stale or unbuilt. A packaged install gets it from pacman
	// at /usr/bin/nejen, so requiring bin/ failed every packaged machine.
	devDispatcher := filepath.Join(nejenPath, "bin", "nejen")
	if stat, err := os.Stat(devDispatcher); err == nil && !stat.IsDir() && stat.Mode()&0111 != 0 {
		pass("dispatcher built: " + devDispatcher)
	} else if _, err := os.Stat(filepath.Join(nejenPath, "go.mod")); err == nil {
		fail("dispatcher missing or not executable: " + devDispatcher + " (run install.sh to rebuild)")
	} else if path, err := exec.LookPath("nejen"); err == nil {
		pass("dispatcher on PATH: " + path)
	} else {
		fail("no dispatcher found: neither " + devDispatcher + " nor nejen on PATH")
	}

	requiredCmds := []string{"bash", "git", "hyprctl", "notify-send"}
	var missingRequired []string
	for _, cmd := range requiredCmds {
		if _, err := exec.LookPath(cmd); err != nil {
			missingRequired = append(missingRequired, cmd)
		}
	}
	if len(missingRequired) == 0 {
		pass(fmt.Sprintf("required commands present: %s", strings.Join(requiredCmds, " ")))
	} else {
		fail(fmt.Sprintf("required commands missing: %s", strings.Join(missingRequired, " ")))
	}

	optionalCmds := []string{
		"gum", "walker", "uwsm", "makoctl", "brightnessctl", "swayosd-client", "slurp", "grim", "satty",
		"hyprpicker", "gpu-screen-recorder", "wl-copy", "elephant", "wiremix", "bluetuith",
		"hyprlock", "hypridle", "hyprsunset", "btop", "swaybg", "zenity",
	}
	var missingOptional []string
	for _, cmd := range optionalCmds {
		if _, err := exec.LookPath(cmd); err != nil {
			missingOptional = append(missingOptional, cmd)
		}
	}
	if len(missingOptional) == 0 {
		pass("optional commands all present")
	} else {
		warn(fmt.Sprintf("optional commands missing (some subcommands degrade): %s", strings.Join(missingOptional, " ")))
	}

	fmt.Printf("\nnejen doctor: %d passed, %d warnings, %d failed\n", passCount, warnCount, failCount)
	if failCount > 0 {
		os.Exit(1)
	}
}

func showHelp() {
	fmt.Println("Usage: nejen <command> [args...]")
	fmt.Println()
	fmt.Println("NEJEN, an opinionated Arch + Hyprland desktop. Commands:")
	fmt.Println()
	fmt.Println("  nejen doctor                                 Check system configuration and dependencies")
	fmt.Println("  nejen version                                Show NEJEN version")
	fmt.Println("  nejen help                                   Show this help")
}

func runOpenBrowser(args []string) {
	cmdOut, err := exec.Command("xdg-settings", "get", "default-web-browser").Output()
	if err != nil {
		fmt.Fprintf(os.Stderr, "nejen open browser: failed to get default browser\n")
		os.Exit(1)
	}
	desktopId := strings.TrimSpace(string(cmdOut))

	home, _ := os.UserHomeDir()
	dirs := []string{
		filepath.Join(home, ".local", "share", "applications"),
		filepath.Join(home, ".nix-profile", "share", "applications"),
		"/usr/share/applications",
	}

	var browserBin string
	for _, dir := range dirs {
		desktopFile := filepath.Join(dir, desktopId)
		if content, err := os.ReadFile(desktopFile); err == nil {
			lines := strings.Split(string(content), "\n")
			for _, line := range lines {
				if strings.HasPrefix(line, "Exec=") {
					execCmd := strings.TrimPrefix(line, "Exec=")
					parts := strings.Split(execCmd, " ")
					if len(parts) > 0 {
						browserBin = parts[0]
						break
					}
				}
			}
			if browserBin != "" {
				break
			}
		}
	}

	if browserBin == "" {
		fmt.Fprintf(os.Stderr, "nejen open browser: no default browser found (xdg-settings said '%s')\n", desktopId)
		os.Exit(1)
	}

	privateFlag := "--incognito"
	switch {
	case strings.Contains(desktopId, "firefox"), strings.Contains(desktopId, "librewolf"), strings.Contains(desktopId, "waterfox"), strings.Contains(desktopId, "zen"), strings.Contains(desktopId, "mozilla"):
		privateFlag = "--private-window"
	case strings.Contains(desktopId, "microsoft-edge"):
		privateFlag = "--inprivate"
	}

	var finalArgs []string
	finalArgs = append(finalArgs, browserBin)
	for _, arg := range args {
		if arg == "--private" {
			finalArgs = append(finalArgs, privateFlag)
		} else {
			finalArgs = append(finalArgs, arg)
		}
	}

	cmd := exec.Command("uwsm-app", append([]string{"--"}, finalArgs...)...)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	if err := cmd.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "nejen open browser: failed to start browser: %v\n", err)
		os.Exit(1)
	}
}

func runThemeCurrent() {
	home, _ := os.UserHomeDir()
	stateHome := os.Getenv("XDG_STATE_HOME")
	if stateHome == "" {
		stateHome = filepath.Join(home, ".local", "state")
	}
	nameFile := filepath.Join(stateHome, "nejen", "theme.name")

	if content, err := os.ReadFile(nameFile); err == nil {
		themeName := strings.TrimSpace(string(content))
		parts := strings.Split(themeName, "-")
		for i, p := range parts {
			if len(p) > 0 {
				parts[i] = strings.ToUpper(p[:1]) + p[1:]
			}
		}
		fmt.Println(strings.Join(parts, " "))
	} else {
		fmt.Println("Unknown")
	}
}

func main() {
	if len(os.Args) < 2 {
		showHelp()
		os.Exit(0)
	}

	nejenPath := getNejenPath()
	if err := os.Setenv("NEJEN_PATH", nejenPath); err != nil {
		fmt.Fprintf(os.Stderr, "failed to set NEJEN_PATH: %v\n", err)
	}

	command := os.Args[1]

	if command == "--internal-complete" {
		for _, cmd := range listCommands() {
			fmt.Println(strings.ReplaceAll(cmd, " ", "-"))
		}
		fmt.Println("doctor")
		fmt.Println("version")
		fmt.Println("theme-current")
		fmt.Println("open-browser")
		return
	}

	// Find the longest match in the registry
	// For `nejen a b c`, try "a b c", then "a b", then "a"
	for n := len(os.Args) - 1; n >= 1; n-- {
		cmdName := strings.Join(os.Args[1:n+1], " ")
		if runCommand(cmdName, os.Args[n+1:]) {
			return
		}
	}

	// Hardcoded fallbacks for the initial commands
	switch command {
	case "help", "--help", "-h":
		showHelp()
	case "doctor":
		runDoctor(nejenPath)
	case "version":
		runVersion(nejenPath, os.Args[2:])
	case "theme":
		if len(os.Args) > 2 && os.Args[2] == "current" {
			runThemeCurrent()
			return
		}
		fallthrough
	case "open":
		if len(os.Args) > 2 && os.Args[2] == "browser" {
			runOpenBrowser(os.Args[3:])
			return
		}
		fallthrough
	default:
		fmt.Fprintf(os.Stderr, "nejen: '%s' is not a nejen command. See 'nejen help'.\n", command)
		os.Exit(1)
	}
}
