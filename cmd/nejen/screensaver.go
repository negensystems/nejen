package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"

	"golang.org/x/sys/unix"
)

func init() {
	registerCommand("screensaver", runScreensaver)
}

// screensaverClass is the app-id the screensaver window is opened with, and
// the handle every check below uses to find it again.
const screensaverClass = "org.nejen.screensaver"

// screensaverFocused reports whether the screensaver is the active window.
//
// The class is compared as a parsed field rather than searched for as a
// substring: hyprctl reports initialClass and the window title alongside
// class, so any window merely mentioning the app-id used to read as focused
// and keep the screensaver up.
func screensaverFocused() bool {
	out, err := exec.Command("hyprctl", "activewindow", "-j").Output()
	if err != nil {
		return false
	}
	var win struct {
		Class string `json:"class"`
	}
	if json.Unmarshal(out, &win) != nil {
		return false
	}
	return win.Class == screensaverClass
}

// screensaverRunning reports whether a screensaver window already exists on
// any workspace.
func screensaverRunning() bool {
	out, err := exec.Command("hyprctl", "clients", "-j").Output()
	if err != nil {
		return false
	}
	var clients []struct {
		Class string `json:"class"`
	}
	if json.Unmarshal(out, &clients) != nil {
		return false
	}
	for _, c := range clients {
		if c.Class == screensaverClass {
			return true
		}
	}
	return false
}

func runScreensaver(args []string) {
	leave := func() {
		exec.Command("hyprctl", "keyword", "cursor:invisible", "false").Run()
		exec.Command("pkill", "-f", screensaverClass).Run()
		os.Exit(0)
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGQUIT)
	go func() {
		<-sigCh
		leave()
	}()

	fmt.Print("\033]11;rgb:00/00/00\007")
	exec.Command("hyprctl", "keyword", "cursor:invisible", "true").Run()

	renderCmd := exec.Command(nejenSelf(), "screensaver", "render")
	renderCmd.Stdout = os.Stdout
	renderCmd.Stderr = os.Stderr
	if err := renderCmd.Start(); err != nil {
		leave()
	}

	// We don't want to leave zombie, but we are going to exit soon anyway.
	// Actually we should wait on it in background to reap.
	go func() {
		renderCmd.Wait()
	}()

	oldState, err := makeRaw(int(os.Stdin.Fd()))
	if err == nil {
		defer restoreTerm(int(os.Stdin.Fd()), oldState)
	}

	inputCh := make(chan struct{})
	go func() {
		b := make([]byte, 1)
		os.Stdin.Read(b)
		inputCh <- struct{}{}
	}()

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		if renderCmd.Process == nil || renderCmd.Process.Signal(syscall.Signal(0)) != nil {
			break
		}

		select {
		case <-inputCh:
			leave()
		case <-ticker.C:
			if !screensaverFocused() {
				leave()
			}
		}
	}
	leave()
}

func makeRaw(fd int) (*unix.Termios, error) {
	termios, err := unix.IoctlGetTermios(fd, unix.TCGETS)
	if err != nil {
		return nil, err
	}
	oldState := *termios
	termios.Lflag &^= unix.ICANON | unix.ECHO | unix.ISIG
	termios.Iflag &^= unix.IXON | unix.ICRNL
	termios.Cc[unix.VMIN] = 1
	termios.Cc[unix.VTIME] = 0
	if err := unix.IoctlSetTermios(fd, unix.TCSETS, termios); err != nil {
		return nil, err
	}
	return &oldState, nil
}

func restoreTerm(fd int, oldState *unix.Termios) {
	unix.IoctlSetTermios(fd, unix.TCSETS, oldState)
}
