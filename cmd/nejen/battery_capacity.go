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
	registerCommand("battery capacity", runBatteryCapacity)
}

func readIntFile(path string) (int, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(strings.TrimSpace(string(data)))
}

func runBatteryCapacity(args []string) {
	bat, err := findBattery()
	if err != nil {
		os.Exit(1)
	}

	for _, prefix := range []string{"energy", "charge"} {
		full, err1 := readIntFile(filepath.Join(bat, prefix+"_full"))
		design, err2 := readIntFile(filepath.Join(bat, prefix+"_full_design"))

		if err1 == nil && err2 == nil && design > 0 {
			fmt.Println((100*full + design/2) / design)
			os.Exit(0)
		}
	}

	// Last resort: upower reports the same ratio on its "capacity" line.
	// upower -i "$(upower -e | grep -m1 -i bat)" 2>/dev/null
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
				re := regexp.MustCompile(`(?i)capacity:\s+([0-9.]+)`)
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
