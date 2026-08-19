package main

import (
	"os"
	"os/exec"
)

func init() {
	registerCommand("voxtype model", runVoxtypeModel)
}

func runVoxtypeModel(args []string) {
	nejenBin := os.Args[0]

	cmd1 := exec.Command(nejenBin, "present", "terminal", "voxtype setup model")
	cmd1.Stdin = os.Stdin
	cmd1.Stdout = os.Stdout
	cmd1.Stderr = os.Stderr
	cmd1.Run()

	cmd2 := exec.Command(nejenBin, "bar", "restart")
	cmd2.Stdin = os.Stdin
	cmd2.Stdout = os.Stdout
	cmd2.Stderr = os.Stderr
	cmd2.Run()
}
