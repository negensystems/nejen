package main

import (
	"encoding/json"
	"fmt"
	"os/exec"
)

func init() {
	registerCommand("win layout", runWinLayout)
}

func runWinLayout(args []string) {
	cmd := exec.Command("hyprctl", "activeworkspace", "-j")
	out, err := cmd.Output()
	if err != nil {
		return
	}

	var ws map[string]interface{}
	if err := json.Unmarshal(out, &ws); err != nil {
		return
	}

	wsIDFloat, ok1 := ws["id"].(float64)
	if !ok1 {
		return
	}
	wsID := fmt.Sprintf("%v", int(wsIDFloat))

	tiledLayout, _ := ws["tiledLayout"].(string)

	target := "dwindle"
	if tiledLayout == "dwindle" {
		target = "scrolling"
	}

	exec.Command("hyprctl", "-q", "keyword", "workspace", fmt.Sprintf("%s, layout:%s", wsID, target)).Run()
	exec.Command("notify-send", fmt.Sprintf("󱂬    Workspace layout: %s", target)).Run()
}
