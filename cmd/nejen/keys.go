package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func init() {
	registerCommand("keys", runKeys)
}

func runKeys(args []string) {
	printMode := false
	for _, arg := range args {
		if arg == "--print" {
			printMode = true
			break
		}
	}

	sheet := cheatsheet(mergedKeymap())

	isTerm := false
	if stat, err := os.Stdout.Stat(); err == nil {
		if (stat.Mode() & os.ModeCharDevice) != 0 {
			isTerm = true
		}
	}

	if isTerm || printMode {
		fmt.Println(sheet)
	} else {

		cmdArgs := []string{"open", "launcher", "--dmenu", "-p", "Keys", "--width", "640", "--minheight", "1", "--maxheight", "720"}
		cmd := exec.Command(os.Args[0], cmdArgs...)
		cmd.Stdin = strings.NewReader(sheet)
		cmd.Run() // output goes to /dev/null
	}
}
