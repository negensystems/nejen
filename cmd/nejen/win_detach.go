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

	pinned, _ := window["pinned"].(bool)
	if pinned {
		batchCmd := fmt.Sprintf("dispatch pin address:%s; dispatch togglefloating address:%s; dispatch tagwindow -pop address:%s", addr, addr, addr)
		exec.Command("hyprctl", "-q", "--batch", batchCmd).Run()
		os.Exit(0)
	}

	exec.Command("hyprctl", "-q", "dispatch", "togglefloating", "address:"+addr).Run()
	exec.Command("hyprctl", "-q", "dispatch", "resizewindowpixel", fmt.Sprintf("exact %s %s,address:%s", width, height, addr)).Run()

	if posX != "" && posY != "" {
		exec.Command("hyprctl", "-q", "dispatch", "movewindowpixel", fmt.Sprintf("exact %s %s,address:%s", posX, posY, addr)).Run()
	} else {
		exec.Command("hyprctl", "-q", "dispatch", "centerwindow").Run()
	}

	batchCmd2 := fmt.Sprintf("dispatch pin address:%s; dispatch alterzorder top address:%s; dispatch tagwindow +pop address:%s", addr, addr, addr)
	exec.Command("hyprctl", "-q", "--batch", batchCmd2).Run()
}
