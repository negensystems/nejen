package main

import (
	"os/exec"
	"strings"
)

func init() {
	registerCommand("dnd toggle", runDndToggle)
}

func runDndToggle(args []string) {
	exec.Command("makoctl", "mode", "-t", "do-not-disturb").Run()

	out, _ := exec.Command("makoctl", "mode").Output()
	if strings.Contains(string(out), "do-not-disturb") {
		exec.Command("notify-send", "󰂛    Do not disturb on").Run()
	} else {
		exec.Command("notify-send", "󰂚    Do not disturb off").Run()
	}

	exec.Command("pkill", "-RTMIN+10", "waybar").Run()
}
