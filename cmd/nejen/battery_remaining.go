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
	registerCommand("battery remaining", runBatteryRemaining)
}

func runBatteryRemaining(args []string) {
	bat, err := findBattery()
	if err != nil {
		os.Exit(1)
	}

	if capVal, err := readIntFile(filepath.Join(bat, "capacity")); err == nil {
		fmt.Println(capVal)
		os.Exit(0)
	}

	for _, prefix := range []string{"energy", "charge"} {
		now, err1 := readIntFile(filepath.Join(bat, prefix+"_now"))
		full, err2 := readIntFile(filepath.Join(bat, prefix+"_full"))
		if err1 == nil && err2 == nil && full > 0 {
			fmt.Println((100*now + full/2) / full)
			os.Exit(0)
		}
	}

	upowerE := exec.Command("upower", "-e")
	outE, err := upowerE.Output()
	if err == nil {
		var batDev string
		for _, line := range strings.Split(string(outE), "\n") {
			if strings.Contains(strings.ToLower(line), "bat") {
				batDev = strings.TrimSpace(line)
				break
			}
		}

		if batDev != "" {
			upowerI := exec.Command("upower", "-i", batDev)
			outI, err := upowerI.Output()
			if err == nil {
				re := regexp.MustCompile(`(?i)percentage:\s+([0-9.]+)`)
				if match := re.FindStringSubmatch(string(outI)); len(match) > 1 {
					if val, err := strconv.ParseFloat(match[1], 64); err == nil {
						fmt.Println(int(val))
						os.Exit(0)
					}
				}
			}
		}
	}

	os.Exit(1)
}
