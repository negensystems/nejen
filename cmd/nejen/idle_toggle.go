package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
)

func init() {
	registerCommand("idle toggle", runIdleToggle)
}

func runIdleToggle(args []string) {
	home, _ := os.UserHomeDir()
	toggleDir := filepath.Join(home, ".local", "state", "nejen", "toggles")
	idleOffFlag := filepath.Join(toggleDir, "idle-off")

	os.MkdirAll(toggleDir, 0755)

	if err := exec.Command("pgrep", "-x", "hypridle").Run(); err == nil {
		exec.Command("pkill", "-x", "hypridle").Run()

		f, _ := os.Create(idleOffFlag)
		if f != nil {
			f.Close()
		}

		exec.Command("notify-send", "󱫖    Staying awake: idle locking disabled").Run()
	} else {
		os.Remove(idleOffFlag)

		cmd := exec.Command("uwsm-app", "--", "hypridle")
		cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
		cmd.Stdout = nil
		cmd.Stderr = nil
		cmd.Stdin = nil
		cmd.Start()

		exec.Command("notify-send", "󱫖    Idle locking enabled").Run()
	}

	exec.Command("pkill", "-RTMIN+9", "waybar").Run()
}
