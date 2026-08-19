package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

func init() {
	registerCommand("theme list", runThemeList)
}

func runThemeList(args []string) {
	nejenPath := getNejenPath()
	home, _ := os.UserHomeDir()

	dirs := []string{
		filepath.Join(nejenPath, "themes"),
		filepath.Join(home, ".config", "nejen", "themes"),
	}

	themeSet := make(map[string]bool)
	for _, dir := range dirs {
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, e := range entries {
			if e.IsDir() {
				themeSet[e.Name()] = true
			}
		}
	}

	var themes []string
	for t := range themeSet {
		themes = append(themes, t)
	}
	sort.Strings(themes)

	for _, t := range themes {
		fmt.Println(t)
	}
}
