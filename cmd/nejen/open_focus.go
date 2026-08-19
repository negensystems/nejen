package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"syscall"
)

type HyprClient struct {
	Address string `json:"address"`
	Class   string `json:"class"`
	Title   string `json:"title"`
}

func init() {
	registerCommand("open focus", runOpenFocus)
}

func runOpenFocus(args []string) {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "Usage: nejen open focus <pattern> [command...]")
		os.Exit(1)
	}

	pattern := args[0]
	cmdArgs := args[1:]

	out, err := exec.Command("hyprctl", "clients", "-j").Output()
	if err == nil {
		var clients []HyprClient
		if err := json.Unmarshal(out, &clients); err == nil {
			reStr := fmt.Sprintf(`(?i)\b%s\b`, regexp.QuoteMeta(pattern))
			if re, err := regexp.Compile(reStr); err == nil {
				var targetAddr string
				for _, client := range clients {
					if re.MatchString(client.Class) || re.MatchString(client.Title) {
						targetAddr = client.Address
						break
					}
				}

				if targetAddr != "" {
					cmdPath, _ := exec.LookPath("hyprctl")
					execArgs := []string{"hyprctl", "dispatch", "focuswindow", "address:" + targetAddr}
					syscall.Exec(cmdPath, execArgs, os.Environ())
					os.Exit(0)
				}
			}
		}
	}

	cmdPath, err := exec.LookPath("setsid")
	if err != nil {
		os.Exit(1)
	}

	if len(cmdArgs) > 0 {
		// Executing directly avoids the bash -c "$*" shell injection
		execArgs := append([]string{"setsid"}, cmdArgs...)
		syscall.Exec(cmdPath, execArgs, os.Environ())
	} else {
		execArgs := []string{"setsid", "uwsm-app", "--", pattern}
		syscall.Exec(cmdPath, execArgs, os.Environ())
	}
}
