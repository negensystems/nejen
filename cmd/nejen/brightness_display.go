package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

func init() {
	registerCommand("brightness display", runBrightnessDisplay)
}

func runBrightnessDisplay(args []string) {
	step := "+5%"
	if len(args) > 0 {
		step = args[0]
	}

	device, err := pickBacklight()
	if err != nil {
		fmt.Fprintln(os.Stderr, "No backlight device found")
		os.Exit(1)
	}

	// Set brightness
	if err := exec.Command("brightnessctl", "--device="+device, "set", step).Run(); err != nil {
		// Ignore errors for now as per bash script (>/dev/null)
	}

	// Report the new level as a percentage on the OSD
	curStr, err1 := exec.Command("brightnessctl", "--device="+device, "get").Output()
	maxStr, err2 := exec.Command("brightnessctl", "--device="+device, "max").Output()

	if err1 == nil && err2 == nil {
		cur, _ := strconv.Atoi(strings.TrimSpace(string(curStr)))
		max, _ := strconv.Atoi(strings.TrimSpace(string(maxStr)))

		if max > 0 {
			level := strconv.Itoa(100 * cur / max)
			exec.Command(nejenSelf(), "swayosd", "brightness", level).Run()
		}
	}
}

func pickBacklight() (string, error) {
	patterns := []string{"amdgpu_bl", "intel_backlight", "nvidia", "acpi_video", ""}
	for _, pattern := range patterns {
		matches, err := filepath.Glob("/sys/class/backlight/*")
		if err != nil {
			continue
		}
		for _, match := range matches {
			name := filepath.Base(match)
			if strings.HasPrefix(name, pattern) {
				return name, nil
			}
		}
	}
	return "", fmt.Errorf("no backlight device found")
}
