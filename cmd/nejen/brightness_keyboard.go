package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

func init() {
	registerCommand("brightness keyboard", runBrightnessKeyboard)
}

func runBrightnessKeyboard(args []string) {
	action := "up"
	if len(args) > 0 {
		action = args[0]
	}

	matches, _ := filepath.Glob("/sys/class/leds/*kbd_backlight*")
	var led string
	for _, path := range matches {
		if stat, err := os.Stat(path); err == nil && stat.IsDir() {
			led = filepath.Base(path)
			break
		}
	}

	if led == "" {
		fmt.Fprintln(os.Stderr, "This machine has no keyboard backlight")
		os.Exit(1)
	}

	curBytes, _ := exec.Command("brightnessctl", "--device="+led, "get").Output()
	maxBytes, _ := exec.Command("brightnessctl", "--device="+led, "max").Output()

	cur, _ := strconv.Atoi(strings.TrimSpace(string(curBytes)))
	max, _ := strconv.Atoi(strings.TrimSpace(string(maxBytes)))

	var target int
	switch action {
	case "cycle":
		target = (cur + 1) % (max + 1)
	case "up":
		target = cur + 1
		if target > max {
			target = max
		}
	case "down":
		target = cur - 1
		if target < 0 {
			target = 0
		}
	default:
		fmt.Fprintln(os.Stderr, "Usage: nejen brightness keyboard [up|down|cycle]")
		os.Exit(1)
	}

	exec.Command("brightnessctl", "--device="+led, "set", strconv.Itoa(target)).Run()

	if max > 0 {
		pct := strconv.Itoa(100 * target / max)
		exec.Command(nejenSelf(), "swayosd", "kbd", "brightness", pct).Run()
	}
}
