package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"syscall"
)

func init() {
	registerCommand("open webapp", runOpenWebapp)
}

func runOpenWebapp(args []string) {
	focusPattern := ""
	if len(args) > 0 && args[0] == "--focus" {
		if len(args) < 2 {
			fmt.Fprintln(os.Stderr, "nejen open webapp: --focus requires a window pattern")
			os.Exit(1)
		}
		focusPattern = args[1]
		args = args[2:]
	}

	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "Usage: nejen open webapp [--focus <pattern>] <url> [browser-flags...]")
		os.Exit(1)
	}

	if focusPattern != "" {
		out, err := exec.Command("hyprctl", "clients", "-j").Output()
		if err == nil {
			var clients []struct {
				Address string `json:"address"`
				Class   string `json:"class"`
				Title   string `json:"title"`
			}
			if json.Unmarshal(out, &clients) == nil {
				reStr := `(?i)\b` + regexp.QuoteMeta(focusPattern) + `\b`
				if strings.Contains(focusPattern, `\b`) || strings.Contains(focusPattern, `.*`) {
					reStr = `(?i)\b` + focusPattern + `\b`
				}
				re, err := regexp.Compile(reStr)
				if err == nil {
					var addr string
					for _, client := range clients {
						if re.MatchString(client.Class) || re.MatchString(client.Title) {
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
	}

	desktopID := "chromium.desktop"
	out, err := exec.Command("xdg-settings", "get", "default-web-browser").Output()
	if err == nil {
		id := strings.TrimSpace(string(out))
		lower := strings.ToLower(id)
		if strings.HasPrefix(lower, "chromium") || strings.HasPrefix(lower, "google-chrome") ||
			strings.HasPrefix(lower, "brave") || strings.HasPrefix(lower, "microsoft-edge") ||
			strings.HasPrefix(lower, "vivaldi") || strings.HasPrefix(lower, "opera") ||
			strings.HasPrefix(lower, "helium") || strings.HasPrefix(lower, "thorium") {
			desktopID = id
		}
	}

	home, _ := os.UserHomeDir()
	dirs := []string{
		filepath.Join(home, ".local/share/applications"),
		filepath.Join(home, ".nix-profile/share/applications"),
		"/usr/share/applications",
	}

	browserBin := ""
	for _, dir := range dirs {
		desktopFile := filepath.Join(dir, desktopID)
		if content, err := os.ReadFile(desktopFile); err == nil {
			lines := strings.Split(string(content), "\n")
			for _, line := range lines {
				if strings.HasPrefix(line, "Exec=") {
					execStr := strings.TrimPrefix(line, "Exec=")
					parts := strings.SplitN(execStr, " ", 2)
					if len(parts) > 0 {
						browserBin = parts[0]
					}
					break
				}
			}
		}
		if browserBin != "" {
			break
		}
	}

	if browserBin == "" {
		fmt.Fprintf(os.Stderr, "nejen open webapp: no usable browser found for '%s'\n", desktopID)
		os.Exit(1)
	}

	url := args[0]
	args = args[1:]

	setsidPath, err := exec.LookPath("setsid")
	if err != nil {
		fmt.Fprintln(os.Stderr, "setsid not found")
		os.Exit(1)
	}

	execArgs := []string{"setsid", "uwsm-app", "--", browserBin, fmt.Sprintf("--app=%s", url)}
	execArgs = append(execArgs, args...)

	syscall.Exec(setsidPath, execArgs, os.Environ())
	os.Exit(1)
}
