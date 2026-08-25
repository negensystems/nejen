package main

import (
	"fmt"
	"os"
	"os/exec"
)

func init() {
	registerCommand("app restart", runAppRestart)
}

func runAppRestart(args []string) {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "Usage: nejen app restart <name> [args...]")
		os.Exit(1)
	}

	appName := args[0]
	// Ignore errors, replicates `|| true`
	exec.Command("pkill", "-x", "--", appName).Run()

	cmdArgs := append([]string{"uwsm-app", "--"}, args...)
	cmd := exec.Command("setsid", cmdArgs...)
	cmd.Stdout = nil
	cmd.Stderr = nil
	cmd.Start()
}
