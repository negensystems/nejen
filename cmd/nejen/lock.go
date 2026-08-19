package main

import (
	"os/exec"
)

func init() {
	registerCommand("lock", runLock)
}

func runLock(args []string) {
	if err := exec.Command("pidof", "hyprlock").Run(); err != nil {
		cmd := exec.Command("hyprlock")
		cmd.Start()
	}

	exec.Command("hyprctl", "switchxkblayout", "all", "0").Run()

	if err := exec.Command("pgrep", "-x", "1password").Run(); err == nil {
		cmd := exec.Command("1password", "--lock")
		cmd.Start()
	}

	exec.Command("pkill", "-f", screensaverClass).Run()
}
