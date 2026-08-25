package main

// Waybar indicator feeds, one JSON line per invocation. These were three
// shell scripts under config/waybar/indicators/; they live in the dispatcher
// now so the bar depends on nothing but the nejen binary. Each pairs with the
// toggle that refreshes it: dnd toggle pokes waybar with RTMIN+10,
// idle toggle with RTMIN+9, screenrecord with RTMIN+8.

import (
	"fmt"
	"os/exec"
	"strings"
)

func init() {
	registerCommand("dnd status", runDndStatus)
	registerCommand("idle status", runIdleStatus)
	registerCommand("screenrecord status", runScreenrecordStatus)
}

// runDndStatus is visible only while mako is in do-not-disturb mode.
func runDndStatus(args []string) {
	out, err := exec.Command("makoctl", "mode").Output()
	if err == nil && strings.Contains(string(out), "do-not-disturb") {
		fmt.Println(`{"text": "󰂛", "tooltip": "Notifications silenced", "class": "active"}`)
		return
	}
	fmt.Println(`{"text": ""}`)
}

// runIdleStatus is visible only while idle locking (hypridle) is disabled.
func runIdleStatus(args []string) {
	if exec.Command("pgrep", "-x", "hypridle").Run() == nil {
		fmt.Println(`{"text": ""}`)
		return
	}
	fmt.Println(`{"text": "󱫖", "tooltip": "Idle locking disabled", "class": "active"}`)
}

// runScreenrecordStatus is visible only while gpu-screen-recorder is capturing.
func runScreenrecordStatus(args []string) {
	if exec.Command("pgrep", "-f", recorderPattern).Run() == nil {
		fmt.Println(`{"text": "󰻂", "tooltip": "Recording - click to stop", "class": "active"}`)
		return
	}
	fmt.Println(`{"text": ""}`)
}
