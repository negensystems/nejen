package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
)

func init() {
	registerCommand("display scale", runDisplayScale)
}

func runDisplayScale(args []string) {
	out, err := exec.Command("hyprctl", "monitors", "-j").Output()
	if err != nil {
		os.Exit(1)
	}

	type monitor struct {
		Name        string  `json:"name"`
		Width       int     `json:"width"`
		Height      int     `json:"height"`
		RefreshRate float64 `json:"refreshRate"`
		Scale       float64 `json:"scale"`
		Focused     bool    `json:"focused"`
	}
	var monitors []monitor

	if err := json.Unmarshal(out, &monitors); err != nil {
		os.Exit(1)
	}

	var focused *monitor
	for i := range monitors {
		if monitors[i].Focused {
			focused = &monitors[i]
			break
		}
	}

	if focused == nil || focused.Name == "" {
		os.Exit(1)
	}

	scaleStr := fmt.Sprintf("%.1f", focused.Scale)
	var next float64
	switch scaleStr {
	case "1.0":
		next = 1.6
	case "1.6":
		next = 2.0
	case "2.0":
		next = 3.0
	default:
		next = 1.0
	}

	mode := fmt.Sprintf("%dx%d@%f", focused.Width, focused.Height, focused.RefreshRate)
	hyprEval(fmt.Sprintf(
		`hl.config({ misc = { disable_scale_notification = true } }); hl.monitor({ output = %s, mode = %s, position = "auto", scale = %v }); hl.config({ misc = { disable_scale_notification = false } })`,
		luaQuote(focused.Name), luaQuote(mode), next,
	))

	nextStr := fmt.Sprintf("%v", next)
	if next == 1.0 || next == 2.0 || next == 3.0 {
		nextStr = fmt.Sprintf("%.0f", next) // bash output had 1, 2, 3 instead of 1.0
	}
	exec.Command("notify-send", fmt.Sprintf("%s  Display scale: %sx", iconMonitor, nextStr)).Run()
}
