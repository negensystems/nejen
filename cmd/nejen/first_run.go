package main

import (
	"os"
	"os/exec"
	"path/filepath"
)

func init() {
	registerCommand("first run", runFirstRun)
}

func runFirstRun(args []string) {
	home, _ := os.UserHomeDir()
	firstRunMode := filepath.Join(home, ".local", "state", "nejen", "first-run.mode")

	if _, err := os.Stat(firstRunMode); err == nil {
		os.Remove(firstRunMode)

		nejenPath := getNejenPath()
		scripts := []string{
			"battery-monitor.sh",
			"cleanup-reboot-sudoers.sh",
			"firewall.sh",
			"dns-resolver.sh",
			"gnome-theme.sh",
			"elephant.sh",
			"pacman-hooks.sh",
			"boot-theme.sh",
		}

		for _, script := range scripts {
			scriptPath := filepath.Join(nejenPath, "install", "first-run", script)
			cmd := exec.Command("bash", scriptPath)
			// The scripts resolve their own paths from NEJEN_PATH; without
			// it inherited they guess, and guess the dev-mode tree.
			cmd.Env = append(os.Environ(), "NEJEN_PATH="+nejenPath)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				os.Exit(1)
			}
		}

		if err := exec.Command("sudo", "rm", "-f", "/etc/sudoers.d/first-run").Run(); err != nil {
			os.Exit(1)
		}

		postScripts := []string{
			"welcome.sh",
			"wifi.sh",
		}

		for _, script := range postScripts {
			scriptPath := filepath.Join(nejenPath, "install", "first-run", script)
			cmd := exec.Command("bash", scriptPath)
			// The scripts resolve their own paths from NEJEN_PATH; without
			// it inherited they guess, and guess the dev-mode tree.
			cmd.Env = append(os.Environ(), "NEJEN_PATH="+nejenPath)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				os.Exit(1)
			}
		}
	}
}
