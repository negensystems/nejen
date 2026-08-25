package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/BurntSushi/toml"
)

func init() {
	registerCommand("theme render", runThemeRender)
}

var (
	placeholderRegex = regexp.MustCompile(`\$\{([A-Za-z0-9_-]+(?:\.[A-Za-z0-9_-]+)+)(?::([^}]*))?\}`)
	hexColorRegex    = regexp.MustCompile(`^#[0-9a-fA-F]{6}$`)
)

func requireHex(value, expr string) (string, error) {
	if !hexColorRegex.MatchString(value) {
		return "", fmt.Errorf("value %q for ${%s} is not a #rrggbb color", value, expr)
	}
	return value, nil
}

func transformNohash(value string, arg *string, expr string) (string, error) {
	if arg != nil {
		return "", fmt.Errorf("transform 'nohash' in ${%s} takes no argument", expr)
	}
	v, err := requireHex(value, expr)
	if err != nil {
		return "", err
	}
	return v[1:], nil
}

type transformFunc func(value string, arg *string, expr string) (string, error)

var transforms = map[string]transformFunc{
	"nohash": transformNohash,
}

func resolve(data map[string]interface{}, expr string, arg *string) (string, error) {
	parts := strings.Split(expr, ".")
	var node interface{} = data
	consumed := 0

	for _, part := range parts {
		if m, ok := node.(map[string]interface{}); ok {
			if next, ok := m[part]; ok {
				node = next
				consumed++
				continue
			}
		}
		break
	}

	remaining := parts[consumed:]

	if len(remaining) > 1 || (len(remaining) == 1 && transforms[remaining[0]] == nil) {
		return "", fmt.Errorf("unknown key ${%s}", expr)
	}

	switch node.(type) {
	case map[string]interface{}, []interface{}:
		return "", fmt.Errorf("${%s} refers to a table, not a value", expr)
	}

	value := fmt.Sprint(node)
	if len(remaining) == 1 {
		return transforms[remaining[0]](value, arg, expr)
	}
	if arg != nil {
		return "", fmt.Errorf("${%s} takes no argument", expr)
	}
	return value, nil
}

func findThemeDir(name, nejenPath string) (string, error) {
	home, _ := os.UserHomeDir()
	candidates := []string{
		filepath.Join(home, ".config", "nejen", "themes", name),
		filepath.Join(nejenPath, "themes", name),
	}
	for _, cand := range candidates {
		if info, err := os.Stat(filepath.Join(cand, "theme.toml")); err == nil && !info.IsDir() {
			return cand, nil
		}
	}
	return "", fmt.Errorf("theme '%s' not found (searched: %s)", name, strings.Join(candidates, ", "))
}

func runThemeRender(args []string) {
	if len(args) != 1 || args[0] == "-h" || args[0] == "--help" {
		fmt.Fprintln(os.Stderr, "usage: nejen theme render <name>")
		os.Exit(2)
	}

	name := args[0]
	nejenPath := getNejenPath()

	themeDir, err := findThemeDir(name, nejenPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "nejen theme render: %v\n", err)
		os.Exit(1)
	}

	var data map[string]interface{}
	tomlPath := filepath.Join(themeDir, "theme.toml")
	if _, err := toml.DecodeFile(tomlPath, &data); err != nil {
		fmt.Fprintf(os.Stderr, "nejen theme render: %s: %v\n", tomlPath, err)
		os.Exit(1)
	}

	templatesDir := filepath.Join(nejenPath, "templates")
	entries, err := os.ReadDir(templatesDir)
	if err != nil || len(entries) == 0 {
		fmt.Fprintf(os.Stderr, "nejen theme render: no templates found in %s\n", templatesDir)
		os.Exit(1)
	}

	var templates []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".tmpl") {
			templates = append(templates, e.Name())
		}
	}
	if len(templates) == 0 {
		fmt.Fprintf(os.Stderr, "nejen theme render: no templates found in %s\n", templatesDir)
		os.Exit(1)
	}

	stateHome := os.Getenv("XDG_STATE_HOME")
	if stateHome == "" {
		home, _ := os.UserHomeDir()
		stateHome = filepath.Join(home, ".local", "state")
	}
	outDir := filepath.Join(stateHome, "nejen", "theme")

	var errorsList []string
	rendered := make(map[string]string)

	for _, tpl := range templates {
		contentBytes, err := os.ReadFile(filepath.Join(templatesDir, tpl))
		if err != nil {
			// Skipping in silence left the previous render of this file in
			// place, so the theme looked applied while one config kept the
			// old theme's values.
			errorsList = append(errorsList, fmt.Sprintf("%s: %v", tpl, err))
			continue
		}
		text := string(contentBytes)
		outName := strings.TrimSuffix(tpl, ".tmpl")

		renderedText := placeholderRegex.ReplaceAllStringFunc(text, func(match string) string {
			submatch := placeholderRegex.FindStringSubmatch(match)
			if len(submatch) == 0 {
				return match
			}
			expr := submatch[1]
			var arg *string
			if strings.Contains(match, ":") {
				argStr := ""
				if len(submatch) > 2 {
					argStr = submatch[2]
				}
				arg = &argStr
			}

			val, err := resolve(data, expr, arg)
			if err != nil {
				errorsList = append(errorsList, fmt.Sprintf("%s: %v", tpl, err))
				return match
			}
			return val
		})

		rendered[outName] = renderedText
	}

	if len(errorsList) > 0 {
		fmt.Fprintf(os.Stderr, "nejen theme render: failed for theme '%s':\n", name)
		for _, errStr := range errorsList {
			fmt.Fprintf(os.Stderr, "  %s\n", errStr)
		}
		os.Exit(1)
	}

	if err := os.MkdirAll(outDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "nejen theme render: %v\n", err)
		os.Exit(1)
	}

	// Written through a temp file and renamed, the way the keymap renderer
	// does it. These outputs are sourced live by hyprland, waybar and the
	// lock screen: a write interrupted midway leaves a truncated config that
	// those programs fail to parse. The errors were also being discarded, so
	// a failed render still reported success.
	for outName, text := range rendered {
		outPath := filepath.Join(outDir, outName)
		tmpPath := outPath + ".tmp"
		if err := os.WriteFile(tmpPath, []byte(text), 0644); err != nil {
			errorsList = append(errorsList, fmt.Sprintf("%s: %v", outName, err))
			continue
		}
		if err := os.Rename(tmpPath, outPath); err != nil {
			os.Remove(tmpPath)
			errorsList = append(errorsList, fmt.Sprintf("%s: %v", outName, err))
		}
	}

	if len(errorsList) > 0 {
		fmt.Fprintf(os.Stderr, "nejen theme render: could not write theme '%s':\n", name)
		for _, errStr := range errorsList {
			fmt.Fprintf(os.Stderr, "  %s\n", errStr)
		}
		os.Exit(1)
	}

	fmt.Printf("Rendered %d files for theme '%s' into %s\n", len(rendered), name, outDir)
}
