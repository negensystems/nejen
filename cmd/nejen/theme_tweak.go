package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

func init() {
	registerCommand("theme tweak", runThemeTweak)
}

// Helper to convert rgb(r,g,b) from zenity to #rrggbb
func zenityToHex(zenityColor string) string {
	zenityColor = strings.TrimSpace(zenityColor)
	if strings.HasPrefix(zenityColor, "#") {
		return zenityColor
	}
	var r, g, b int
	if _, err := fmt.Sscanf(zenityColor, "rgb(%d,%d,%d)", &r, &g, &b); err == nil {
		return fmt.Sprintf("#%02x%02x%02x", r, g, b)
	}
	if _, err := fmt.Sscanf(zenityColor, "rgba(%d,%d,%d,%*f)", &r, &g, &b); err == nil {
		return fmt.Sprintf("#%02x%02x%02x", r, g, b)
	}
	return zenityColor
}

func runThemeTweak(args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: nejen theme tweak <property> [color]")
		os.Exit(1)
	}
	property := args[0]
	color := ""
	if len(args) > 1 {
		color = args[1]
	}

	if color == "" {
		if _, err := exec.LookPath("zenity"); err == nil {
			out, err := exec.Command("zenity", "--color-selection", "--show-palette", "--title", "Pick "+property+" color").Output()
			if err != nil {
				os.Exit(1)
			}
			color = zenityToHex(string(out))
		} else {
			fmt.Println("Error: zenity not found and no color provided.")
			os.Exit(1)
		}
	}

	home, _ := os.UserHomeDir()
	nejenPath := getNejenPath()

	// Get active theme
	activeTheme := "nejen"
	currentBytes, err := os.ReadFile(filepath.Join(home, ".local", "state", "nejen", "theme.name"))
	if err == nil && len(currentBytes) > 0 {
		activeTheme = strings.TrimSpace(string(currentBytes))
	}

	userThemeDir := filepath.Join(home, ".config", "nejen", "themes", activeTheme)
	userThemePath := filepath.Join(userThemeDir, "theme.toml")

	// If the active theme is a system theme, we must clone it first so we don't modify system files
	if _, err := os.Stat(userThemePath); err != nil {
		newTheme := activeTheme + "-custom"
		exec.Command(filepath.Join(nejenPath, "bin", "nejen"), "theme", "clone", newTheme).Run()
		activeTheme = newTheme
		userThemeDir = filepath.Join(home, ".config", "nejen", "themes", activeTheme)
		userThemePath = filepath.Join(userThemeDir, "theme.toml")
		exec.Command(filepath.Join(nejenPath, "bin", "nejen"), "theme", "set", activeTheme).Run()
	}

	properties := []string{property}
	if property == "main" {
		properties = []string{"accent", "border", "cursor"}
	}

	content, err := os.ReadFile(userThemePath)
	if err != nil {
		fmt.Printf("Error reading theme.toml: %v\n", err)
		os.Exit(1)
	}

	newContent := content
	for _, prop := range properties {
		re := regexp.MustCompile(`(?m)^(\s*` + regexp.QuoteMeta(prop) + `\s*=\s*")[^"]+(".*)$`)
		if re.Match(newContent) {
			newContent = re.ReplaceAll(newContent, []byte("${1}"+color+"${2}"))
		}
	}
	
	os.WriteFile(userThemePath, newContent, 0644)

	// Re-render and apply
	exec.Command(filepath.Join(nejenPath, "bin", "nejen"), "theme", "set", activeTheme).Run()
	exec.Command(filepath.Join(nejenPath, "bin", "nejen"), "bar", "restart").Run()
	exec.Command("notify-send", "-u", "normal", "Theme Updated", "Changed "+property+" to "+color).Run()
}
