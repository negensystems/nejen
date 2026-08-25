package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

func init() {
	registerCommand("brightness display apple", runBrightnessDisplayApple)
}

func runBrightnessDisplayApple(args []string) {
	maxLevel := 60000

	if len(args) == 0 {
		fmt.Printf("Usage: nejen brightness display apple <+step|-step|level>   (range 0-%d, e.g. +5000)\n", maxLevel)
		os.Exit(0)
	}

	step := args[0]

	// Find the hiddev node belonging to an Apple display.
	matches, _ := filepath.Glob("/dev/usb/hiddev*")
	if len(matches) == 0 {
		fmt.Fprintln(os.Stderr, "No Apple Studio/XDR Display detected")
		os.Exit(1)
	}

	detectArgs := append([]string{"asdcontrol", "--detect"}, matches...)
	out, _ := exec.Command("sudo", detectArgs...).Output()

	var hiddev string
	for _, line := range strings.Split(string(out), "\n") {
		if strings.HasPrefix(line, "/dev/usb/hiddev") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) > 0 {
				hiddev = parts[0]
				break
			}
		}
	}

	if hiddev == "" {
		fmt.Fprintln(os.Stderr, "No Apple Studio/XDR Display detected")
		os.Exit(1)
	}

	exec.Command("sudo", "asdcontrol", hiddev, "--", step).Run()

	out, _ = exec.Command("sudo", "asdcontrol", hiddev).Output()
	var level int
	var found bool
	for _, line := range strings.Split(string(out), "\n") {
		if strings.HasPrefix(line, "BRIGHTNESS=") {
			valStr := strings.TrimPrefix(line, "BRIGHTNESS=")
			if val, err := strconv.Atoi(valStr); err == nil {
				level = val
				found = true
			}
			break
		}
	}

	isDigit := regexp.MustCompile(`^[0-9]+$`).MatchString
	if found && isDigit(strconv.Itoa(level)) {
		pct := strconv.Itoa(100 * level / maxLevel)
		exec.Command(nejenSelf(), "swayosd", "brightness", pct).Run()
	}
}
