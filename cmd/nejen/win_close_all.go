package main

import (
	"encoding/json"
	"fmt"
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
					hyprDispatchLua(fmt.Sprintf("hl.dsp.window.close({ window = %s })", luaQuote("address:"+addr)))
				}
			}
		}
	}
	hyprDispatchLua("hl.dsp.focus({ workspace = 1 })")
}
