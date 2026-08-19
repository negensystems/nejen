package main

import (
	"os"
	"os/exec"
	"syscall"
	"time"
)

func init() {
	registerCommand("system logout", runSystemLogout)
	registerCommand("__internal delay_logout", runDelayLogout)
}

func runDelayLogout(args []string) {
	time.Sleep(2 * time.Second)
	cmd := exec.Command("uwsm", "stop")
	cmd.Run()
}

func runSystemLogout(args []string) {
	cmd := exec.Command(os.Args[0], "__internal", "delay_logout")
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	cmd.Start()

	closeCmd := exec.Command(nejenSelf(), "win", "close-all")
	closeCmd.Stdout = os.Stdout
	closeCmd.Stderr = os.Stderr
	closeCmd.Run()

	time.Sleep(1 * time.Second)
}
