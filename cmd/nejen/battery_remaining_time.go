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
	registerCommand("battery remaining time", runBatteryRemainingTime)
}

func formatMinutes(total int) string {
	if total%60 == 0 {
		return fmt.Sprintf("%dh", total/60)
	}
	return fmt.Sprintf("%dh %dm", total/60, total%60)
}

func runBatteryRemainingTime(args []string) {
	bat, err := findBattery()
	if err != nil {
		os.Exit(1)
	}

	stateBytes, _ := os.ReadFile(filepath.Join(bat, "status"))
	state := strings.TrimSpace(string(stateBytes))

	pairs := []struct {
		prefix   string
		rateFile string
	}{
		{"energy", "power_now"},
		{"charge", "current_now"},
	}

	var minutes int
	var found = false

	for _, p := range pairs {
		nowFile := filepath.Join(bat, p.prefix+"_now")
		fullFile := filepath.Join(bat, p.prefix+"_full")
		ratePath := filepath.Join(bat, p.rateFile)

		now, err1 := readIntFile(nowFile)
		rate, err2 := readIntFile(ratePath)
		if err1 != nil || err2 != nil || rate <= 0 {
			continue
		}

		if state == "Discharging" {
			minutes = 60 * now / rate
			found = true
			break
		} else if state == "Charging" {
			full, err3 := readIntFile(fullFile)
			if err3 != nil {
				continue
			}
			minutes = 60 * (full - now) / rate
			found = true
			break
		}
	}

	if found && minutes >= 0 {
		fmt.Println(formatMinutes(minutes))
		os.Exit(0)
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
				re := regexp.MustCompile(`(?i)time to (?:empty|full):\s+([0-9.]+)\s+(hours|minutes)`)
				if match := re.FindStringSubmatch(string(outI)); len(match) > 2 {
					if val, err := strconv.ParseFloat(match[1], 64); err == nil {
						mins := 0
						if match[2] == "minutes" {
							mins = int(val)
						} else {
							mins = int(val * 60)
						}
						fmt.Println(formatMinutes(mins))
						os.Exit(0)
					}
				}
			}
		}
	}

	os.Exit(1)
}
