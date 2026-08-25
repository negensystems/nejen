package main

import (
	"os"
	"os/exec"
	"regexp"
)

func init() {
	registerCommand("walker restart", runWalkerRestart)
}

func restartWalkerServices() {
	cmd1 := exec.Command("systemctl", "--user", "is-enabled", "elephant.service")
	if err := cmd1.Run(); err == nil {
		exec.Command("systemctl", "--user", "restart", "elephant.service").Run()
	}

	cmd2 := exec.Command("systemctl", "--user", "is-enabled", "app-walker@autostart.service")
	if err := cmd2.Run(); err == nil {
		exec.Command("systemctl", "--user", "restart", "app-walker@autostart.service").Run()
	}
}

func runWalkerRestart(args []string) {
	if os.Geteuid() == 0 {
		entries, err := os.ReadDir("/run/user/")
		if err != nil {
			return
		}

		uidRegex := regexp.MustCompile(`^[0-9]+$`)
		for _, entry := range entries {
			uid := entry.Name()
			if uidRegex.MatchString(uid) {
				xdgRuntimeDir := "/run/user/" + uid

				cmd := exec.Command("systemd-run", "--quiet", "--uid="+uid, "--setenv=XDG_RUNTIME_DIR="+xdgRuntimeDir,
					os.Args[0], "walker", "restart")
				cmd.Run()
			}
		}
	} else {
		restartWalkerServices()
	}
}
