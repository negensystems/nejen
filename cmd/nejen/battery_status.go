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
	registerCommand("battery status", runBatteryStatus)
}

func runBatteryStatus(args []string) {
	bat, err := findBattery()
	if err != nil {
		fmt.Println("No battery detected")
		os.Exit(1)
	}

	exe, _ := os.Executable()
	levelBytes, _ := exec.Command(exe, "battery", "remaining").Output()
	etaBytes, _ := exec.Command(exe, "battery", "remaining", "time").Output()
	healthBytes, _ := exec.Command(exe, "battery", "capacity").Output()

	level := strings.TrimSpace(string(levelBytes))
	eta := strings.TrimSpace(string(etaBytes))
	health := strings.TrimSpace(string(healthBytes))

	stateBytes, _ := os.ReadFile(filepath.Join(bat, "status"))
	state := strings.TrimSpace(string(stateBytes))

	watts := ""
	if powerNow, err := readIntFile(filepath.Join(bat, "power_now")); err == nil {
		watts = fmt.Sprintf("%.1f", float64(powerNow)/1e6)
	} else {
		currentNow, err1 := readIntFile(filepath.Join(bat, "current_now"))
		voltageNow, err2 := readIntFile(filepath.Join(bat, "voltage_now"))
		if err1 == nil && err2 == nil {
			watts = fmt.Sprintf("%.1f", (float64(currentNow)*float64(voltageNow))/1e12)
		}
	}

	icons := []string{"󰂎", "󰁺", "󰁻", "󰁼", "󰁽", "󰁾", "󰁿", "󰂀", "󰂁", "󰂂", "󰁹"}
	var icon string
	if state == "Charging" {
		icon = "󰂄"
	} else {
		levelInt, _ := strconv.Atoi(level)
		idx := levelInt / 10
		if idx < 0 {
			idx = 0
		}
		if idx > 10 {
			idx = 10
		}
		icon = icons[idx]
	}

	if level == "" {
		level = "?"
	}

	line := fmt.Sprintf("%s  %s%%", icon, level)

	switch state {
	case "Charging":
		line += " charging"
		if eta != "" {
			line += fmt.Sprintf(", %s until full", eta)
		}
	case "Discharging":
		if eta != "" {
			line += fmt.Sprintf(", %s left", eta)
		}
		if watts != "" {
			line += fmt.Sprintf(", drawing %sW", watts)
		}
	case "Full":
		line += " fully charged"
	}

	if health != "" {
		line += fmt.Sprintf("  ·  health %s%%", health)
	}

	fmt.Println(line)
}
