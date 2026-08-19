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
	registerCommand("battery monitor", runBatteryMonitor)
}

func runBatteryMonitor(args []string) {
	const threshold = 10
	markerDir := os.Getenv("XDG_RUNTIME_DIR")
	if markerDir == "" {
		markerDir = fmt.Sprintf("/run/user/%d", os.Getuid())
	}
	marker := filepath.Join(markerDir, "nejen-battery-low.alerted")

	bat, err := findBattery()
	if err != nil {
		os.Exit(0)
	}

	stateBytes, _ := os.ReadFile(filepath.Join(bat, "status"))
	state := strings.TrimSpace(string(stateBytes))

	var level int
	levelFound := false

	capBytes, err := os.ReadFile(filepath.Join(bat, "capacity"))
	if err == nil {
		if l, err := strconv.Atoi(strings.TrimSpace(string(capBytes))); err == nil {
			level = l
			levelFound = true
		}
	}

	if !levelFound {
		for _, prefix := range []string{"energy", "charge"} {
			now, err1 := readIntFile(filepath.Join(bat, prefix+"_now"))
			full, err2 := readIntFile(filepath.Join(bat, prefix+"_full"))
			if err1 == nil && err2 == nil && full > 0 {
				level = 100 * now / full
				levelFound = true
				break
			}
		}
	}

	if !levelFound {
		os.Exit(0)
	}

	if state == "Discharging" && level <= threshold {
		if _, err := os.Stat(marker); os.IsNotExist(err) {
			exec.Command("notify-send", "-u", "critical", "-i", "battery-caution", "-t", "30000",
				"󱐋 Plug in soon", fmt.Sprintf("Battery at %d%%", level)).Run()
			os.WriteFile(marker, []byte(""), 0644)
		}
	} else {
		os.Remove(marker)
	}
}
