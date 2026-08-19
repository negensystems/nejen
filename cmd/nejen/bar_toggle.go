package main

import (
	"os/exec"
)

func init() {
	registerCommand("bar toggle", runBarToggle)
}

func runBarToggle(args []string) {
	err := exec.Command("pgrep", "-x", "waybar").Run()
	if err == nil {
		exec.Command("pkill", "-x", "waybar").Run()
	} else {
		cmd := exec.Command("setsid", "uwsm-app", "--", "waybar")
		cmd.Stdout = nil
		cmd.Stderr = nil
		cmd.Start()
	}
}
