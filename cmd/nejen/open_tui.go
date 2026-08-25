package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
)

func init() {
	registerCommand("open tui", runOpenTui)
}

func runOpenTui(args []string) {
	focus := false
	appID := ""

	// Flags come first and in any order: --focus reuses an existing window
	// of the same app-id, --app-id names that window explicitly (needed
	// whenever the program is a generic host like `bash`, since every such
	// panel would otherwise collide on org.nejen.bash).
flags:
	for len(args) > 0 {
		switch {
		case args[0] == "--focus":
			focus = true
			args = args[1:]
		case strings.HasPrefix(args[0], "--app-id="):
			appID = strings.TrimPrefix(args[0], "--app-id=")
			args = args[1:]
		default:
			break flags
		}
	}

	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "Usage: nejen open tui [--focus] [--app-id=<id>] <program> [args...]")
		os.Exit(1)
	}

	if appID == "" {
		prog := strings.Split(args[0], " ")[0]
		appID = "org.nejen." + filepath.Base(prog)
	}

	if focus {
		out, err := exec.Command("hyprctl", "clients", "-j").Output()
		if err == nil {
			var clients []struct {
				Address string `json:"address"`
				Class   string `json:"class"`
			}
			if json.Unmarshal(out, &clients) == nil {
				var addr string
				for _, client := range clients {
					if client.Class == appID {
						addr = client.Address
						break
					}
				}
				if addr != "" {
					hyprctlPath, err := exec.LookPath("hyprctl")
					if err == nil {
						syscall.Exec(hyprctlPath, []string{"hyprctl", "dispatch", "focuswindow", "address:" + addr}, os.Environ())
					}
					os.Exit(1)
				}
			}
		}
	}

	setsidPath, err := exec.LookPath("setsid")
	if err != nil {
		fmt.Fprintln(os.Stderr, "setsid not found")
		os.Exit(1)
	}

	execArgs := []string{"setsid", "uwsm-app", "--", "xdg-terminal-exec", fmt.Sprintf("--app-id=%s", appID), "-e"}
	execArgs = append(execArgs, args...)

	syscall.Exec(setsidPath, execArgs, os.Environ())
	os.Exit(1)
}
