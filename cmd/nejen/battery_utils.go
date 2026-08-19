package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func findBattery() (string, error) {
	matches, err := filepath.Glob("/sys/class/power_supply/*")
	if err != nil {
		return "", err
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

		return supply, nil
	}

	return "", fmt.Errorf("no battery found")
}
