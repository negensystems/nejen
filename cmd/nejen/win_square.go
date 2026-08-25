package main

import (
	"os/exec"
	"strings"
)

func init() {
	registerCommand("win square", runWinSquare)
}

func runWinSquare(args []string) {
	cmd := exec.Command("hyprctl", "getoption", "layout:single_window_aspect_ratio")
	out, err := cmd.Output()
	if err == nil && strings.Contains(string(out), "[1, 1]") {
		exec.Command("hyprctl", "-q", "keyword", "layout:single_window_aspect_ratio", "0 0").Run()
		exec.Command("notify-send", "󰝤    Single-window square aspect off").Run()
	} else {
		exec.Command("hyprctl", "-q", "keyword", "layout:single_window_aspect_ratio", "1 1").Run()
		exec.Command("notify-send", "󰝤    Single-window square aspect on").Run()
	}
}
