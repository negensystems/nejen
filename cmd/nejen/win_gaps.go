package main

import (
	"encoding/json"
	"os/exec"
	"strings"
)

func init() {
	registerCommand("win gaps", runWinGaps)
}

func runWinGaps(args []string) {
	cmd := exec.Command("hyprctl", "getoption", "general:gaps_out", "-j")
	out, err := cmd.Output()
	if err != nil {
		return
	}

	var option map[string]interface{}
	if err := json.Unmarshal(out, &option); err != nil {
		return
	}

	// Hyprland 0.57+ reports CSS gap values in "css" (earlier versions used "custom").
	gapVal, _ := option["css"].(string)
	if gapVal == "" {
		gapVal, _ = option["custom"].(string)
	}
	fields := strings.Fields(gapVal)
	if len(fields) > 0 && fields[0] == "0" {
		hyprEval("hl.config({ general = { gaps_out = 10, gaps_in = 5, border_size = 2 } })")
	} else {
		hyprEval("hl.config({ general = { gaps_out = 0, gaps_in = 0, border_size = 0 } })")
	}
}
