package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

func init() {
	registerCommand("browser", runBrowserHelp)
	registerCommand("browser list", runBrowserList)
	registerCommand("browser current", runBrowserCurrent)
	registerCommand("browser set", runBrowserSet)
	registerCommand("browser select", runBrowserSelect)
}

type browserApp struct {
	id         string
	name       string
	path       string
	execCmd    string
	execTokens []string
	isDefault  bool
}

// discoverBrowsers scans application directories for desktop entries
// declaring the WebBrowser category or handling http/https schemes.
func discoverBrowsers() []*browserApp {
	home, _ := os.UserHomeDir()
	dirs := []string{
		filepath.Join(home, ".local", "share", "applications"),
		filepath.Join(home, ".nix-profile", "share", "applications"),
		"/usr/local/share/applications",
		"/usr/share/applications",
	}

	currentDefault := ""
	if out, err := exec.Command("xdg-settings", "get", "default-web-browser").Output(); err == nil {
		currentDefault = strings.TrimSpace(string(out))
	}

	seen := make(map[string]bool)
	var browsers []*browserApp

	for _, dir := range dirs {
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}

		for _, entry := range entries {
			id := entry.Name()
			if entry.IsDir() || !strings.HasSuffix(id, ".desktop") || seen[id] {
				continue
			}

			filePath := filepath.Join(dir, id)
			content, err := os.ReadFile(filePath)
			if err != nil {
				continue
			}

			lines := strings.Split(string(content), "\n")
			isApp := false
			noDisplay := false
			hasHTTP := false
			hasWebCat := false
			name := ""
			execLine := ""

			for _, line := range lines {
				line = strings.TrimSpace(line)
				switch {
				case line == "Type=Application":
					isApp = true
				case strings.EqualFold(line, "NoDisplay=true"):
					noDisplay = true
				case strings.HasPrefix(line, "Name=") && name == "":
					name = strings.TrimPrefix(line, "Name=")
				case strings.HasPrefix(line, "Exec=") && execLine == "":
					execLine = strings.TrimPrefix(line, "Exec=")
				case strings.HasPrefix(line, "MimeType="):
					lower := strings.ToLower(line)
					if strings.Contains(lower, "x-scheme-handler/http") || strings.Contains(lower, "x-scheme-handler/https") {
						hasHTTP = true
					}
				case strings.HasPrefix(line, "Categories="):
					if strings.Contains(strings.ToLower(line), "webbrowser") {
						hasWebCat = true
					}
				}
			}

			if !isApp || noDisplay {
				continue
			}

			if !hasHTTP && !(hasWebCat && (strings.Contains(execLine, "%u") || strings.Contains(execLine, "%U"))) {
				continue
			}

			tokens := parseDesktopExec(execLine)
			if len(tokens) == 0 {
				continue
			}

			if name == "" {
				name = strings.TrimSuffix(id, ".desktop")
			}

			isDef := id == currentDefault || strings.TrimSuffix(id, ".desktop") == strings.TrimSuffix(currentDefault, ".desktop")

			seen[id] = true
			browsers = append(browsers, &browserApp{
				id:         id,
				name:       name,
				path:       filePath,
				execCmd:    execLine,
				execTokens: tokens,
				isDefault:  isDef,
			})
		}
	}

	sort.Slice(browsers, func(i, j int) bool {
		return strings.ToLower(browsers[i].name) < strings.ToLower(browsers[j].name)
	})

	return browsers
}

func matchBrowser(browsers []*browserApp, query string) *browserApp {
	q := strings.ToLower(strings.TrimSpace(query))
	if q == "" {
		return nil
	}

	for _, b := range browsers {
		id := strings.ToLower(strings.TrimSuffix(b.id, ".desktop"))
		if id == q || strings.ToLower(b.name) == q || strings.ToLower(b.id) == q {
			return b
		}
	}

	for _, b := range browsers {
		name := strings.ToLower(b.name)
		id := strings.ToLower(b.id)
		if strings.HasPrefix(name, q) || strings.HasPrefix(id, q) || strings.Contains(name, q) {
			return b
		}
	}

	return nil
}

func browserBinary(tokens []string) string {
	for _, tok := range tokens {
		if tok == "env" || strings.Contains(tok, "=") {
			continue
		}
		base := filepath.Base(tok)
		if trimmed := strings.TrimSuffix(base, "-bin"); trimmed != base {
			if _, err := exec.LookPath(trimmed); err == nil {
				return trimmed
			}
		}
		return base
	}
	return ""
}

// saveBrowserEnv records the browser binary that config/uwsm/default exports
// as $BROWSER. It goes in state because ~/.config/uwsm links to the shipped
// tree: a git checkout in dev mode, root-owned under /usr/share otherwise.
func saveBrowserEnv(b *browserApp) error {
	bin := browserBinary(b.execTokens)
	if bin == "" {
		return nil
	}

	home, _ := os.UserHomeDir()
	dir := filepath.Join(home, ".local", "state", "nejen")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, "browser"), []byte(bin+"\n"), 0644)
}

func setBrowser(b *browserApp) {
	exec.Command("xdg-settings", "set", "default-web-browser", b.id).Run()

	mimes := []string{
		"text/html",
		"application/xhtml+xml",
		"x-scheme-handler/http",
		"x-scheme-handler/https",
	}
	for _, m := range mimes {
		exec.Command("xdg-mime", "default", b.id, m).Run()
	}

	if err := saveBrowserEnv(b); err != nil {
		fmt.Fprintf(os.Stderr, "nejen browser set: $BROWSER not saved: %v\n", err)
	}

	exec.Command("notify-send", fmt.Sprintf("Default browser: %s", b.name)).Run()
	fmt.Printf("Default browser set to %s (%s)\n", b.name, b.id)
}

func runBrowserList(args []string) {
	browsers := discoverBrowsers()
	if len(browsers) == 0 {
		fmt.Println("No web browsers detected.")
		return
	}

	if len(args) > 0 && (args[0] == "--names" || args[0] == "--raw") {
		for _, b := range browsers {
			fmt.Println(b.name)
		}
		return
	}

	fmt.Println("Installed web browsers:")
	for _, b := range browsers {
		prefix := "  "
		suffix := ""
		if b.isDefault {
			prefix = "* "
			suffix = " [current default]"
		}
		fmt.Printf("%s%-24s (%s)%s\n", prefix, b.name, b.id, suffix)
	}
}

func runBrowserCurrent(args []string) {
	browsers := discoverBrowsers()
	for _, b := range browsers {
		if b.isDefault {
			fmt.Printf("%s (%s)\n", b.name, b.id)
			return
		}
	}

	if out, err := exec.Command("xdg-settings", "get", "default-web-browser").Output(); err == nil {
		id := strings.TrimSpace(string(out))
		if id != "" {
			fmt.Println(id)
			return
		}
	}
	fmt.Println("No default browser set")
}

func runBrowserSet(args []string) {
	if len(args) == 0 {
		runBrowserSelect(args)
		return
	}

	browsers := discoverBrowsers()
	if len(browsers) == 0 {
		fmt.Fprintln(os.Stderr, "nejen browser set: no installed web browsers found")
		os.Exit(1)
	}

	query := strings.Join(args, " ")
	target := matchBrowser(browsers, query)
	if target == nil {
		fmt.Fprintf(os.Stderr, "nejen browser set: unknown browser '%s'\n\n", query)
		runBrowserList(nil)
		os.Exit(1)
	}

	setBrowser(target)
}

func runBrowserSelect(args []string) {
	browsers := discoverBrowsers()
	if len(browsers) == 0 {
		fmt.Fprintln(os.Stderr, "nejen browser select: no web browsers found")
		os.Exit(1)
	}

	var names []string
	for _, b := range browsers {
		names = append(names, b.name)
	}

	selectedName := ""

	isTerm := false
	if stat, err := os.Stdout.Stat(); err == nil {
		if (stat.Mode() & os.ModeCharDevice) != 0 {
			isTerm = true
		}
	}

	if isTerm {
		gumCmd := exec.Command("gum", "filter", "--height", "10", "--header", "Select Default Browser", "--placeholder", "Type to filter...")
		gumCmd.Stdin = strings.NewReader(strings.Join(names, "\n"))
		var gumOut bytes.Buffer
		gumCmd.Stdout = &gumOut
		gumCmd.Stderr = os.Stderr
		if err := gumCmd.Run(); err == nil {
			selectedName = strings.TrimSpace(gumOut.String())
		}
	} else {
		cmd := exec.Command(nejenSelf(), "open", "launcher", "--dmenu", "-p", "Default Browser",
			"--width", "400", "--minheight", "1", "--maxheight", "400")
		cmd.Stdin = strings.NewReader(strings.Join(names, "\n"))
		out, err := cmd.Output()
		if err == nil {
			selectedName = strings.TrimSpace(string(out))
		}
	}

	if selectedName == "" {
		return
	}

	target := matchBrowser(browsers, selectedName)
	if target != nil {
		setBrowser(target)
	}
}

func runBrowserHelp(args []string) {
	out, status := os.Stdout, 0
	if len(args) > 0 && args[0] != "help" && args[0] != "--help" && args[0] != "-h" {
		fmt.Fprintf(os.Stderr, "nejen browser: unknown subcommand '%s'\n\n", args[0])
		out, status = os.Stderr, 2
	}

	fmt.Fprintln(out, "Usage: nejen browser <command>")
	fmt.Fprintln(out)
	fmt.Fprintln(out, "  list [--names]       List installed browsers (* marks current default)")
	fmt.Fprintln(out, "  current              Show the active default browser")
	fmt.Fprintln(out, "  set <name|id>        Set default browser by name or desktop file")
	fmt.Fprintln(out, "  select               Pick default browser interactively")
	fmt.Fprintln(out)

	os.Exit(status)
}
