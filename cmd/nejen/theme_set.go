package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
)

func init() {
	registerCommand("theme set", runThemeSet)
}

func runThemeSet(args []string) {
	if len(args) == 0 || args[0] == "" {
		fmt.Fprintln(os.Stderr, "usage: nejen theme set <name>")
		os.Exit(2)
	}
	name := args[0]

	nejenPath := getNejenPath()
	home, _ := os.UserHomeDir()

	themeDir := ""
	candidates := []string{
		filepath.Join(home, ".config", "nejen", "themes", name),
		filepath.Join(nejenPath, "themes", name),
	}
	for _, cand := range candidates {
		if info, err := os.Stat(filepath.Join(cand, "theme.toml")); err == nil && !info.IsDir() {
			themeDir = cand
			break
		}
	}

	if themeDir == "" {
		fmt.Fprintf(os.Stderr, "nejen theme set: unknown theme '%s'. Available themes:\n", name)
		listCmd := exec.Command(os.Args[0], "theme", "list")
		listCmd.Stderr = os.Stderr
		listCmd.Run()
		os.Exit(1)
	}

	renderCmd := exec.Command(os.Args[0], "theme", "render", name)
	renderCmd.Stdout = os.Stdout
	renderCmd.Stderr = os.Stderr
	renderCmd.Run()

	nejenStateDir := nejenState()
	os.MkdirAll(nejenStateDir, 0755)

	os.WriteFile(filepath.Join(nejenStateDir, "theme.name"), []byte(name+"\n"), 0644)

	// Adopt the theme's default wallpaper, but never over one the user
	// chose from their own folder: personal wallpapers are theme-independent
	// and switching themes should not throw the collection away.
	if !bgIsUserOwned(currentBg()) {
		var data map[string]interface{}
		if _, err := toml.DecodeFile(filepath.Join(themeDir, "theme.toml"), &data); err == nil {
			if wpMap, ok := data["wallpaper"].(map[string]interface{}); ok {
				if defaultWp, ok := wpMap["default"].(string); ok && defaultWp != "" {
					wpPath := filepath.Join(themeDir, defaultWp)
					if _, err := os.Stat(wpPath); err == nil {
						linkPath := filepath.Join(nejenStateDir, "background")
						os.Remove(linkPath)
						os.Symlink(wpPath, linkPath)
					}
				}
			}
		}
	}

	if _, err := exec.LookPath("hyprctl"); err == nil && os.Getenv("HYPRLAND_INSTANCE_SIGNATURE") != "" {
		exec.Command("hyprctl", "reload").Run()
	}

	if err := exec.Command("pgrep", "-x", "waybar").Run(); err == nil {
		if _, err := exec.LookPath("nejen"); err == nil {
			exec.Command("nejen", "bar", "restart").Run()
		} else {
			exec.Command("pkill", "-SIGUSR2", "-x", "waybar").Run()
		}
	}

	if _, err := exec.LookPath("makoctl"); err == nil {
		if err := exec.Command("pgrep", "-x", "mako").Run(); err == nil {
			exec.Command("makoctl", "reload").Run()
		}
	}

	if err := exec.Command("pgrep", "-x", "btop").Run(); err == nil {
		exec.Command("pkill", "-SIGUSR2", "-x", "btop").Run()
	}

	if _, err := exec.LookPath("systemctl"); err == nil {
		exec.Command("systemctl", "--user", "try-restart", "swayosd.service").Run()
	}

	if _, err := os.Stat(bgLink()); err == nil {
		reloadSwaybg()
	}

	if _, err := exec.LookPath("gsettings"); err == nil {
		iconsFile := filepath.Join(nejenStateDir, "theme", "icons.theme")
		if b, err := os.ReadFile(iconsFile); err == nil {
			themeName := strings.TrimSpace(string(b))
			exec.Command("gsettings", "set", "org.gnome.desktop.interface", "icon-theme", themeName).Run()
		}
	}

	fmt.Printf("Theme set to '%s'\n", name)
}
