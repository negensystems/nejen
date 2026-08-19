package main

import (
	"os"
	"os/exec"
	"syscall"
	"time"
)

func init() {
	registerCommand("system shutdown", runSystemShutdown)
	registerCommand("__internal delay_shutdown", runDelayShutdown)
}

func runDelayShutdown(args []string) {
	time.Sleep(2 * time.Second)
	cmd := exec.Command("systemctl", "poweroff", "--no-wall")
	cmd.Run()
}

func runSystemShutdown(args []string) {
	cmd := exec.Command(os.Args[0], "__internal", "delay_shutdown")
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	cmd.Start()

	closeCmd := exec.Command(nejenSelf(), "win", "close-all")
	closeCmd.Stdout = os.Stdout
	closeCmd.Stderr = os.Stderr
	closeCmd.Run()

	time.Sleep(1 * time.Second)
}
