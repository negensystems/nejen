package main

import (
	"os/exec"
)

func init() {
	registerCommand("open bluetooth", runOpenBluetooth)
}

// runOpenBluetooth opens the quick-connect menu, which is what you want
// nine times out of ten: reconnect a device you already paired. Pairing,
// scanning, renaming and file transfer live in bluetuith, reachable from the
// menu's own "Manage devices" entry or directly with --manager.
func runOpenBluetooth(args []string) {
	exec.Command("rfkill", "unblock", "bluetooth").Run()

	for _, arg := range args {
		if arg == "--manager" {
			runOpenTui([]string{"--focus", "bluetuith"})
			return
		}
	}

	// --nosearch: the list is every device you own, a handful of rows that fit
	// on screen at once, so a query box would only add a step between opening
	// the menu and arrowing to the device. Rows are two lines (name + detail
	// line), so the height allows for roughly eight of them.
	runOpenLauncher([]string{
		"-m", "menus:nejenbluetooth", "--nosearch",
		"--minheight", "1", "--maxheight", "480",
	})
}
