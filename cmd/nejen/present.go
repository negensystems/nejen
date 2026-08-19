package main

import (
	"os"
	"os/exec"
)

func init() {
	registerCommand("present", runPresent)
}

func runPresent(args []string) {
	for _, name := range args {
		_, err := exec.LookPath(name)
		if err != nil {
			os.Exit(1)
		}
	}
	os.Exit(0)
}
