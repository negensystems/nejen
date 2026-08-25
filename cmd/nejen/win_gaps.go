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

	customVal, _ := option["custom"].(string)
	fields := strings.Fields(customVal)
	if len(fields) > 0 && fields[0] == "0" {
		exec.Command("hyprctl", "-q", "--batch", "keyword general:gaps_out 10; keyword general:gaps_in 5; keyword general:border_size 2").Run()
	} else {
		exec.Command("hyprctl", "-q", "--batch", "keyword general:gaps_out 0; keyword general:gaps_in 0; keyword general:border_size 0").Run()
	}
}
