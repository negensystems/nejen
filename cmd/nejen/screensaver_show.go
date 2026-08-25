package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func init() {
	registerCommand("screensaver show", runScreensaverShow)
}

func runScreensaverShow(args []string) {
	nejenPath := getNejenPath()

	if screensaverRunning() {
		os.Exit(0)
	}

	home, _ := os.UserHomeDir()
	togglesPath := filepath.Join(home, ".local", "state", "nejen", "toggles", "screensaver-off")
	isForce := len(args) > 0 && args[0] == "force"
	if !isForce {
		if _, err := os.Stat(togglesPath); err == nil {
			os.Exit(1)
		}
	}

	exec.Command("walker", "-q").Run()

	outTerm, _ := exec.Command("xdg-terminal-exec", "--print-id").Output()
	terminalID := string(outTerm)

	outMon, _ := exec.Command("hyprctl", "monitors", "-j").Output()
	cameFrom := ""
	var monitors []struct {
		Name    string `json:"name"`
		Focused bool   `json:"focused"`
	}
	if json.Unmarshal(outMon, &monitors) == nil {
		for _, m := range monitors {
			if m.Focused {
				cameFrom = m.Name
				break
			}
		}
	}

	openScreensaverWindow := func() {
		if strings.Contains(terminalID, "Alacritty") {
			opts := []string{"--class=" + screensaverClass}
			cfg := filepath.Join(nejenPath, "config", "screensaver", "alacritty.toml")
			if _, err := os.Stat(cfg); err == nil {
				opts = append(opts, "--config-file", cfg)
			}
			opts = append(opts, "-e", "nejen", "screensaver")
			exec.Command("hyprctl", "dispatch", "exec", "--", "alacritty "+strings.Join(opts, " ")).Run()
		} else if strings.Contains(terminalID, "ghostty") {
			opts := []string{"--class=" + screensaverClass, "--font-size=18"}
			cfg := filepath.Join(nejenPath, "config", "screensaver", "ghostty")
			if _, err := os.Stat(cfg); err == nil {
				opts = append(opts, fmt.Sprintf("--config-file=%s", cfg))
			}
			opts = append(opts, "-e", "nejen", "screensaver")
			exec.Command("hyprctl", "dispatch", "exec", "--", "ghostty "+strings.Join(opts, " ")).Run()
		} else if strings.Contains(terminalID, "kitty") {
			exec.Command("hyprctl", "dispatch", "exec", "--", fmt.Sprintf("kitty --class=%s --override font_size=18 --override window_padding_width=0 -e nejen screensaver", screensaverClass)).Run()
		} else {
			exec.Command("notify-send", "Screensaver needs a supported terminal", "Set the default to Alacritty, Ghostty, or Kitty.").Run()
		}
	}

	for _, m := range monitors {
		exec.Command("hyprctl", "dispatch", "focusmonitor", m.Name).Run()
		openScreensaverWindow()
	}

	if cameFrom != "" {
		exec.Command("hyprctl", "dispatch", "focusmonitor", cameFrom).Run()
	}
}
