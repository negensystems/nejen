package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
)

func init() {
	registerCommand("swayosd kbd brightness", runSwayosdKbdBrightness)
}

func runSwayosdKbdBrightness(args []string) {
	percentStr := "0"
	if len(args) > 0 {
		percentStr = args[0]
	}

	percent, err := strconv.ParseFloat(percentStr, 64)
	if err != nil {
		log.Fatalf("Invalid percent: %v", err)
	}

	fraction := percent / 100.0
	if fraction < 0.01 {
		fraction = 0.01
	}
	if fraction > 1.0 {
		fraction = 1.0
	}
	fractionStr := fmt.Sprintf("%.2f", fraction)

	cmd := exec.Command("hyprctl", "monitors", "-j")
	out, err := cmd.Output()
	if err != nil {
		log.Fatalf("Failed to run hyprctl: %v", err)
	}

	type Monitor struct {
		Name    string `json:"name"`
		Focused bool   `json:"focused"`
	}

	var monitors []Monitor
	if err := json.Unmarshal(out, &monitors); err != nil {
		log.Fatalf("Failed to parse hyprctl output: %v", err)
	}

	focusedName := ""
	for _, m := range monitors {
		if m.Focused {
			focusedName = m.Name
			break
		}
	}

	swayCmd := exec.Command("swayosd-client",
		"--monitor", focusedName,
		"--custom-icon", "keyboard-brightness",
		"--custom-progress", fractionStr,
		"--custom-progress-text", fmt.Sprintf("%.0f%%", percent),
	)
	swayCmd.Stdout = os.Stdout
	swayCmd.Stderr = os.Stderr
	if err := swayCmd.Run(); err != nil {
		log.Fatalf("Failed to run swayosd-client: %v", err)
	}
}
