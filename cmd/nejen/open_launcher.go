package main

import (
	"os"
	"os/exec"
	"syscall"
)

func init() {
	registerCommand("open launcher", runOpenLauncher)
}

func runOpenLauncher(args []string) {
	if err := exec.Command("pgrep", "-x", "elephant").Run(); err != nil {
		cmd := exec.Command("setsid", "uwsm-app", "--", "elephant")
		cmd.Start()
	}

	if err := exec.Command("pgrep", "-f", "walker --gapplication-service").Run(); err != nil {
		cmd := exec.Command("setsid", "uwsm-app", "--", "walker", "--gapplication-service")
		cmd.Start()
	}

	walkerPath, err := exec.LookPath("walker")
	if err == nil {
		// The 300px default suits a text-only result list. Callers whose rows
		// are taller -- anything with icons -- pass their own sizing, and
		// repeating the flag would leave walker with two of each.
		sized := false
		for _, arg := range args {
			if arg == "--maxheight" || arg == "--minheight" || arg == "-h" {
				sized = true
				break
			}
		}

		cmdArgs := []string{"walker", "--width", "644"}
		if !sized {
			cmdArgs = append(cmdArgs, "--maxheight", "300", "--minheight", "300")
		}
		cmdArgs = append(cmdArgs, args...)

		syscall.Exec(walkerPath, cmdArgs, os.Environ())
	}
	os.Exit(1)
}
