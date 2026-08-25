package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func init() {
	registerCommand("clip", runClip)
}

func runClip(args []string) {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "Usage: nejen clip [copy|paste]")
		os.Exit(1)
	}

	var key string
	switch args[0] {
	case "copy":
		key = "C"
	case "paste":
		key = "V"
	default:
		fmt.Fprintln(os.Stderr, "Usage: nejen clip [copy|paste]")
		os.Exit(1)
	}

	out, err := exec.Command("hyprctl", "activewindow", "-j").Output()
	if err != nil {
		os.Exit(0)
	}

	var window struct {
		Address string   `json:"address"`
		Tags    []string `json:"tags"`
	}
	if err := json.Unmarshal(out, &window); err != nil || window.Address == "" {
		os.Exit(0)
	}

	mods := "CTRL"
	for _, tag := range window.Tags {
		if strings.TrimRight(tag, "*") == "terminal" {
			mods = "CTRL SHIFT"
			break
		}
	}

	exec.Command("hyprctl", "-q", "dispatch", "sendshortcut", fmt.Sprintf("%s, %s, address:%s", mods, key, window.Address)).Run()
}
