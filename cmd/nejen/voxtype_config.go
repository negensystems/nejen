package main

import (
	"os"
	"path/filepath"
)

func init() {
	registerCommand("voxtype config", runVoxtypeConfig)
}

func runVoxtypeConfig(args []string) {
	home, _ := os.UserHomeDir()
	configPath := filepath.Join(home, ".config", "voxtype", "config.toml")
	runOpenEditor([]string{configPath})
}
