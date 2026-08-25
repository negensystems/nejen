package main

import (
	"os"
	"os/exec"
	"syscall"
	"time"
)

func init() {
	registerCommand("system reboot", runSystemReboot)
	registerCommand("__internal delay_reboot", runDelayReboot)
}

func runDelayReboot(args []string) {
	time.Sleep(2 * time.Second)
	cmd := exec.Command("systemctl", "reboot", "--no-wall")
	cmd.Run()
}

func runSystemReboot(args []string) {
	cmd := exec.Command(os.Args[0], "__internal", "delay_reboot")
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	cmd.Start()

	closeCmd := exec.Command(nejenSelf(), "win", "close-all")
	closeCmd.Stdout = os.Stdout
	closeCmd.Stderr = os.Stderr
	closeCmd.Run()

	time.Sleep(1 * time.Second)
}
