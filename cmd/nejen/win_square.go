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
		hyprEval("hl.config({ layout = { single_window_aspect_ratio = { 0, 0 } } })")
		exec.Command("notify-send", iconSquare+"  Single-window square aspect off").Run()
	} else {
		hyprEval("hl.config({ layout = { single_window_aspect_ratio = { 1, 1 } } })")
		exec.Command("notify-send", iconSquare+"  Single-window square aspect on").Run()
	}
}
