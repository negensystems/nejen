package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
)

func init() {
	registerCommand("open editor", runOpenEditor)
}

func runOpenEditor(args []string) {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "nvim"
	}
	if _, err := exec.LookPath(editor); err != nil {
		editor = "nvim"
	}

	base := filepath.Base(editor)
	isTui := false
	switch base {
	case "nvim", "vim", "vi", "nano", "micro", "hx", "helix", "fresh":
		isTui = true
	}

	if isTui {
		cmdArgs := append([]string{"--focus", editor}, args...)
		runOpenTui(cmdArgs)
	} else {
		cmdPath, err := exec.LookPath("setsid")
		if err != nil {
			os.Exit(1)
		}
		cmdArgs := append([]string{"setsid", "uwsm-app", "--", editor}, args...)
		env := os.Environ()
		if err := syscall.Exec(cmdPath, cmdArgs, env); err != nil {
			os.Exit(1)
		}
	}
}
