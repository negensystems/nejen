package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

func init() {
	registerCommand("terminal cwd", runTerminalCwd)
}

func giveUp() {
	home, _ := os.UserHomeDir()
	fmt.Println(home)
	os.Exit(0)
}

func looksLikeShell(exe string) bool {
	base := filepath.Base(exe)
	switch base {
	case "bash", "zsh", "fish", "sh", "dash", "ksh", "tcsh", "csh", "nu", "nushell", "elvish", "xonsh":
		return true
	}

	f, err := os.Open("/etc/shells")
	if err == nil {
		defer f.Close()
		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == exe {
				return true
			}
		}
	}
	return false
}

func runTerminalCwd(args []string) {
	cmd := exec.Command("hyprctl", "activewindow", "-j")
	out, err := cmd.Output()
	if err != nil {
		giveUp()
	}

	var active struct {
		Pid int `json:"pid"`
	}
	if err := json.Unmarshal(out, &active); err != nil {
		giveUp()
	}
	if active.Pid == 0 {
		giveUp()
	}

	pgrep := exec.Command("pgrep", "-P", strconv.Itoa(active.Pid))
	pout, _ := pgrep.Output()

	shellPid := ""
	lines := strings.Split(strings.TrimSpace(string(pout)), "\n")
	for _, line := range lines {
		child := strings.TrimSpace(line)
		if child == "" {
			continue
		}
		exe, err := os.Readlink(fmt.Sprintf("/proc/%s/exe", child))
		if err != nil {
			continue
		}
		if looksLikeShell(exe) {
			shellPid = child
		}
	}

	if shellPid == "" {
		giveUp()
	}

	cwd, err := os.Readlink(fmt.Sprintf("/proc/%s/cwd", shellPid))
	if err != nil {
		giveUp()
	}
	info, err := os.Stat(cwd)
	if err != nil || !info.IsDir() {
		giveUp()
	}

	fmt.Println(cwd)
}
