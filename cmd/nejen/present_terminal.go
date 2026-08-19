package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"
)

func init() {
	registerCommand("present terminal", runPresentTerminal)
}

func runPresentTerminal(args []string) {
	nejenPath := getNejenPath()
	os.Setenv("NEJEN_PATH", nejenPath)

	joinedArgs := strings.Join(args, " ")
	session := fmt.Sprintf(`clear; echo -e "\033[1;32m"; cat '%s/logo.txt' 2>/dev/null; echo -e "\033[0m\n"; %s; [[ $? -eq 130 ]] || { echo; if command -v gum >/dev/null; then gum spin --spinner globe --title "Done! Press any key to close..." -- bash -c "read -n 1 -s"; else echo "Done! Press any key to close..."; read -n 1 -s; fi; }`, nejenPath, joinedArgs)

	setsidPath, err := exec.LookPath("setsid")
	if err != nil {
		fmt.Fprintln(os.Stderr, "setsid not found")
		os.Exit(1)
	}

	execArgs := []string{"setsid", "uwsm-app", "--", "xdg-terminal-exec", "--app-id=org.nejen.terminal", "--title=Nejen", "-e", "bash", "-c", session}

	syscall.Exec(setsidPath, execArgs, os.Environ())
	os.Exit(1)
}
