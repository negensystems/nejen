package main

import (
	"os"

	"golang.org/x/sys/unix"
)

// A character-device check is not enough: /dev/null is one too.
func isTerminal(f *os.File) bool {
	_, err := unix.IoctlGetTermios(int(f.Fd()), unix.TCGETS)
	return err == nil
}
