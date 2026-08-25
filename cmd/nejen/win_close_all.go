package main

import (
	"encoding/json"
	"os/exec"
)

func init() {
	registerCommand("win close-all", runWinCloseAll)
}

func runWinCloseAll(args []string) {
	cmd := exec.Command("hyprctl", "clients", "-j")
	out, err := cmd.Output()
	if err == nil {
		var clients []map[string]interface{}
		if err := json.Unmarshal(out, &clients); err == nil {
			for _, client := range clients {
				if addr, ok := client["address"].(string); ok {
					exec.Command("hyprctl", "-q", "dispatch", "closewindow", "address:"+addr).Run()
				}
			}
		}
	}
	exec.Command("hyprctl", "-q", "dispatch", "workspace", "1").Run()
}
