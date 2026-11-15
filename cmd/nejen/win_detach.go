package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
)

func init() {
	registerCommand("win detach", runWinDetach)
}

func runWinDetach(args []string) {
	width := "1300"
	height := "900"
	var posX, posY string

	if len(args) > 0 {
		width = args[0]
	}
	if len(args) > 1 {
		height = args[1]
	}
	if len(args) > 2 {
		posX = args[2]
	}
	if len(args) > 3 {
		posY = args[3]
	}

	cmd := exec.Command("hyprctl", "activewindow", "-j")
	out, err := cmd.Output()
	if err != nil {
		os.Exit(0)
	}

	var window map[string]interface{}
	if err := json.Unmarshal(out, &window); err != nil {
		os.Exit(0)
	}

	addr, ok := window["address"].(string)
	if !ok || addr == "" {
		os.Exit(0)
	}

	winRef := luaQuote("address:" + addr)
	pinned, _ := window["pinned"].(bool)
	if pinned {
		hyprEval(fmt.Sprintf(
			`hl.dispatch(hl.dsp.window.pin({ action = "toggle", window = %s })); hl.dispatch(hl.dsp.window.float({ action = "toggle", window = %s })); hl.dispatch(hl.dsp.window.tag({ tag = "-pop", window = %s }))`,
			winRef, winRef, winRef,
		))
		os.Exit(0)
	}

	var steps []string
	steps = append(steps,
		fmt.Sprintf(`hl.dispatch(hl.dsp.window.float({ action = "toggle", window = %s }))`, winRef),
		fmt.Sprintf(`hl.dispatch(hl.dsp.window.resize({ x = %s, y = %s, relative = false, window = %s }))`, luaNumberOrQuote(width), luaNumberOrQuote(height), winRef),
	)
	if posX != "" && posY != "" {
		steps = append(steps, fmt.Sprintf(`hl.dispatch(hl.dsp.window.move({ x = %s, y = %s, relative = false, window = %s }))`, luaNumberOrQuote(posX), luaNumberOrQuote(posY), winRef))
	} else {
		steps = append(steps, `hl.dispatch(hl.dsp.window.center())`)
	}
	steps = append(steps,
		fmt.Sprintf(`hl.dispatch(hl.dsp.window.pin({ action = "toggle", window = %s }))`, winRef),
		fmt.Sprintf(`hl.dispatch(hl.dsp.window.alter_zorder({ mode = "top", window = %s }))`, winRef),
		fmt.Sprintf(`hl.dispatch(hl.dsp.window.tag({ tag = "+pop", window = %s }))`, winRef),
	)
	for _, step := range steps {
		hyprEval(step)
	}
}
