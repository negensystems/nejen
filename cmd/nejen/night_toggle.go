package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"time"
)

func init() {
	registerCommand("night toggle", runNightToggle)
}

func runNightToggle(args []string) {
	const warmTemp = "4000"
	const dayTemp = "6000"

	if err := exec.Command("pgrep", "-x", "hyprsunset").Run(); err != nil {
		cmd := exec.Command("setsid", "uwsm-app", "--", "hyprsunset")
		cmd.Start()
		time.Sleep(1 * time.Second)
	}

	out, _ := exec.Command("hyprctl", "hyprsunset", "temperature").Output()

	re := regexp.MustCompile("[^0-9]")
	current := re.ReplaceAllString(string(out), "")

	if current == dayTemp {
		exec.Command("hyprctl", "hyprsunset", "temperature", warmTemp).Run()
		exec.Command("notify-send", "  Night light on").Run()
	} else {
		exec.Command("hyprctl", "hyprsunset", "temperature", dayTemp).Run()
		exec.Command("notify-send", "   Night light off").Run()
	}

	home, _ := os.UserHomeDir()
	configPath := filepath.Join(home, ".config", "waybar", "config.jsonc")
	if content, err := os.ReadFile(configPath); err == nil {
		if bytes.Contains(content, []byte("custom/nightlight")) {
			exec.Command(os.Args[0], "bar", "restart").Run()
		}
	}
}
