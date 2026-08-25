package main

import (
	"os"
	"path/filepath"
	"strings"
)

func init() {
	registerCommand("battery present", runBatteryPresent)
}

func runBatteryPresent(args []string) {
	matches, err := filepath.Glob("/sys/class/power_supply/*")
	if err != nil {
		os.Exit(1)
	}

	for _, supply := range matches {
		typeBytes, err := os.ReadFile(filepath.Join(supply, "type"))
		if err != nil || strings.TrimSpace(string(typeBytes)) != "Battery" {
			continue
		}

		scopeBytes, err := os.ReadFile(filepath.Join(supply, "scope"))
		if err == nil && strings.TrimSpace(string(scopeBytes)) == "Device" {
			continue
		}

		presentBytes, err := os.ReadFile(filepath.Join(supply, "present"))
		if err == nil && strings.TrimSpace(string(presentBytes)) != "1" {
			continue
		}

		os.Exit(0)
	}

	os.Exit(1)
}
