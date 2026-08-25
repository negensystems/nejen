package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func init() {
	registerCommand("tz select", runTzSelect)
}

func runTzSelect(args []string) {
	// timedatectl list-timezones
	tzCmd := exec.Command("timedatectl", "list-timezones")
	tzOutput, err := tzCmd.Output()
	if err != nil {
		os.Exit(1)
	}

	// gum filter
	gumCmd := exec.Command("gum", "filter", "--height", "20", "--header", "Set system timezone", "--placeholder", "Type to search...")
	gumCmd.Stdin = bytes.NewReader(tzOutput)
	var gumOut bytes.Buffer
	gumCmd.Stdout = &gumOut
	gumCmd.Stderr = os.Stderr

	if err := gumCmd.Run(); err != nil {
		// e.g. cancelled (exit status 130)
		if exitError, ok := err.(*exec.ExitError); ok {
			os.Exit(exitError.ExitCode())
		}
		os.Exit(1)
	}

	zone := strings.TrimSpace(gumOut.String())
	if zone == "" {
		os.Exit(1)
	}

	// sudo timedatectl set-timezone "$zone"
	sudoCmd := exec.Command("sudo", "timedatectl", "set-timezone", zone)
	sudoCmd.Stdin = os.Stdin
	sudoCmd.Stdout = os.Stdout
	sudoCmd.Stderr = os.Stderr
	if err := sudoCmd.Run(); err != nil {
		os.Exit(1)
	}

	fmt.Printf("System timezone is now %s\n", zone)

	// nejen bar restart
	// Since we are in nejen, we could just execute the binary
	barCmd := exec.Command(os.Args[0], "bar", "restart")
	barCmd.Stdout = os.Stdout
	barCmd.Stderr = os.Stderr
	barCmd.Run()
}
