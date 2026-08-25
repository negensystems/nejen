package main

import (
	"fmt"
	"os"
	"os/exec"
)

func init() {
	registerCommand("search", runSearch)
}

func runSearch(args []string) {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "usage: nejen search <query>")
		os.Exit(1)
	}
	helper := aurHelper()
	if helper == "" {
		helper = "pacman"
	}
	cmd := exec.Command(helper, append([]string{"-Ss"}, args...)...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			os.Exit(exitError.ExitCode())
		}
		os.Exit(1)
	}
}
