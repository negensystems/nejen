package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func init() {
	registerCommand("theme clone", runThemeClone)
}

func runThemeClone(args []string) {
	name := ""
	if len(args) > 0 {
		name = strings.Join(args, " ")
	}

	if name == "" {
		if _, err := exec.LookPath("zenity"); err == nil {
			out, err := exec.Command("zenity", "--entry", "--title", "Clone Theme", "--text", "Enter a name for your new theme:").Output()
			if err != nil {
				os.Exit(1)
			}
			name = strings.TrimSpace(string(out))
		} else if _, err := exec.LookPath("gum"); err == nil {
			out, err := exec.Command("gum", "input", "--header", "Enter a name for your new theme:").Output()
			if err != nil {
				os.Exit(1)
			}
			name = strings.TrimSpace(string(out))
		} else {
			fmt.Println("Error: No theme name provided and zenity/gum not found.")
			os.Exit(1)
		}
	}

	if name == "" {
		os.Exit(1)
	}

	home, _ := os.UserHomeDir()
	destDir := filepath.Join(home, ".config", "nejen", "themes", name)
	if _, err := os.Stat(destDir); err == nil {
		exec.Command("notify-send", "-u", "critical", "NEJEN Theme", "Theme '"+name+"' already exists!").Run()
		os.Exit(1)
	}

	// Copy from active theme or default to nejen
	activeTheme := "nejen"
	currentBytes, err := os.ReadFile(filepath.Join(home, ".local", "state", "nejen", "theme.name"))
	if err == nil && len(currentBytes) > 0 {
		activeTheme = strings.TrimSpace(string(currentBytes))
	}

	nejenPath := getNejenPath()
	var srcTheme string
	// Check user dir first, then system dir
	userThemePath := filepath.Join(home, ".config", "nejen", "themes", activeTheme, "theme.toml")
	if _, err := os.Stat(userThemePath); err == nil {
		srcTheme = userThemePath
	} else {
		srcTheme = filepath.Join(nejenPath, "themes", activeTheme, "theme.toml")
	}

	if err := os.MkdirAll(destDir, 0755); err != nil {
		fmt.Printf("Failed to create theme directory: %v\n", err)
		os.Exit(1)
	}

	srcFile, err := os.Open(srcTheme)
	if err != nil {
		fmt.Printf("Failed to open source theme: %v\n", err)
		os.Exit(1)
	}
	defer srcFile.Close()

	destPath := filepath.Join(destDir, "theme.toml")
	destFile, err := os.Create(destPath)
	if err != nil {
		fmt.Printf("Failed to create new theme file: %v\n", err)
		os.Exit(1)
	}
	defer destFile.Close()

	if _, err := io.Copy(destFile, srcFile); err != nil {
		fmt.Printf("Failed to copy theme file: %v\n", err)
		os.Exit(1)
	}

	// Create empty backgrounds dir
	os.MkdirAll(filepath.Join(destDir, "backgrounds"), 0755)

	// Replace the name field in the copied TOML
	content, _ := os.ReadFile(destPath)
	newContent := bytes.Replace(content, []byte("name = \""+activeTheme+"\""), []byte("name = \""+name+"\""), 1)
	if !bytes.Contains(content, []byte("name = \""+activeTheme+"\"")) {
		// fallback to replacing any name=... line if the active theme name doesn't exactly match
		lines := strings.Split(string(content), "\n")
		for i, line := range lines {
			if strings.HasPrefix(strings.TrimSpace(line), "name =") {
				lines[i] = "name = \"" + name + "\""
				break
			}
		}
		newContent = []byte(strings.Join(lines, "\n"))
	}
	os.WriteFile(destPath, newContent, 0644)

	// Open in editor
	exec.Command(filepath.Join(nejenPath, "bin", "nejen"), "open", "editor", destPath).Run()

	exec.Command("notify-send", "-u", "normal", "NEJEN Theme", "Cloned '"+activeTheme+"' to '"+name+"'. Edit the colors and save!").Run()
}
