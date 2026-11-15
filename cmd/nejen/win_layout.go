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

	hyprEval(fmt.Sprintf("hl.workspace_rule({ workspace = %s, layout = %s })", luaQuote(wsID), luaQuote(target)))
	exec.Command("notify-send", fmt.Sprintf("%s  Workspace layout: %s", iconLayout, target)).Run()
}
