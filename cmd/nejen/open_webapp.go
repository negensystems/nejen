package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"syscall"
)

func init() {
	registerCommand("open webapp", runOpenWebapp)
}

// parseDesktopExec splits an Exec= value into the command nejen runs. Any
// argument carrying a field code (%u, --url=%U) is dropped, since the caller
// appends its own target; %% is a literal percent.
func parseDesktopExec(execStr string) []string {
	var tokens []string
	rest := strings.TrimLeft(execStr, " \t")
	for rest != "" {
		var arg string
		arg, rest = cutExecArg(rest)
		rest = strings.TrimLeft(rest, " \t")

		if strings.Contains(strings.ReplaceAll(arg, "%%", ""), "%") {
			continue
		}
		tokens = append(tokens, strings.ReplaceAll(arg, "%%", "%"))
	}
	return tokens
}

// cutExecArg takes the first argument off s. The Desktop Entry spec quotes
// with double quotes only, and a backslash escapes only inside them.
func cutExecArg(s string) (arg, rest string) {
	var b strings.Builder
	quoted := false
	for i := 0; i < len(s); i++ {
		switch c := s[i]; {
		case c == '"':
			quoted = !quoted
		case quoted && c == '\\' && i+1 < len(s):
			i++
			b.WriteByte(s[i])
		case !quoted && (c == ' ' || c == '\t'):
			return b.String(), s[i:]
		default:
			b.WriteByte(c)
		}
	}
	return b.String(), ""
}

func runOpenWebapp(args []string) {
	focusPattern := ""
	if len(args) > 0 && args[0] == "--focus" {
		if len(args) < 2 {
			fmt.Fprintln(os.Stderr, "nejen open webapp: --focus requires a window pattern")
			os.Exit(1)
		}
		focusPattern = args[1]
		args = args[2:]
	}

	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "Usage: nejen open webapp [--focus <pattern>] <url> [browser-flags...]")
		os.Exit(1)
	}

	if focusPattern != "" {
		out, err := exec.Command("hyprctl", "clients", "-j").Output()
		if err == nil {
			var clients []struct {
				Address string `json:"address"`
				Class   string `json:"class"`
				Title   string `json:"title"`
			}
			if json.Unmarshal(out, &clients) == nil {
				reStr := `(?i)\b` + regexp.QuoteMeta(focusPattern) + `\b`
				if strings.Contains(focusPattern, `\b`) || strings.Contains(focusPattern, `.*`) {
					reStr = `(?i)\b` + focusPattern + `\b`
				}
				re, err := regexp.Compile(reStr)
				if err == nil {
					var addr string
					for _, client := range clients {
						if re.MatchString(client.Class) || re.MatchString(client.Title) {
							addr = client.Address
							break
						}
					}
					if addr != "" {
						hyprctlPath, err := exec.LookPath("hyprctl")
						if err == nil {
							syscall.Exec(hyprctlPath, []string{"hyprctl", "dispatch", "focuswindow", "address:" + addr}, os.Environ())
						}
						os.Exit(1)
					}
				}
			}
		}
	}

	var candidates []string
	if out, err := exec.Command("xdg-settings", "get", "default-web-browser").Output(); err == nil {
		id := strings.TrimSpace(string(out))
		if id != "" {
			candidates = append(candidates, id)
		}
	}
	candidates = append(candidates, "zen.desktop", "zen-browser.desktop", "chromium.desktop", "google-chrome.desktop", "firefox.desktop")

	home, _ := os.UserHomeDir()
	dirs := []string{
		filepath.Join(home, ".local/share/applications"),
		filepath.Join(home, ".nix-profile/share/applications"),
		"/usr/share/applications",
	}

	var cmdTokens []string
	var selectedDesktopID string
	for _, cand := range candidates {
		for _, dir := range dirs {
			desktopFile := filepath.Join(dir, cand)
			if content, err := os.ReadFile(desktopFile); err == nil {
				lines := strings.Split(string(content), "\n")
				for _, line := range lines {
					if strings.HasPrefix(line, "Exec=") {
						execStr := strings.TrimPrefix(line, "Exec=")
						tokens := parseDesktopExec(execStr)
						if len(tokens) > 0 {
							cmdTokens = tokens
							selectedDesktopID = cand
							break
						}
					}
				}
			}
			if len(cmdTokens) > 0 {
				break
			}
		}
		if len(cmdTokens) > 0 {
			break
		}
	}

	if len(cmdTokens) == 0 {
		fmt.Fprintf(os.Stderr, "nejen open webapp: no usable browser found\n")
		os.Exit(1)
	}

	url := args[0]
	args = args[1:]

	setsidPath, err := exec.LookPath("setsid")
	if err != nil {
		fmt.Fprintln(os.Stderr, "setsid not found")
		os.Exit(1)
	}

	// Webapps in Chromium-based browsers use --app=URL for a dedicated window;
	// Firefox and its forks (like Zen) open an app window via --new-window.
	isGecko := isGeckoBrowser(selectedDesktopID, cmdTokens)

	execArgs := append([]string{"setsid", "uwsm-app", "--"}, cmdTokens...)
	if isGecko {
		execArgs = append(execArgs, "--new-window", url)
	} else {
		execArgs = append(execArgs, fmt.Sprintf("--app=%s", url))
	}
	execArgs = append(execArgs, args...)

	syscall.Exec(setsidPath, execArgs, os.Environ())
	os.Exit(1)
}

func isGeckoBrowser(desktopID string, tokens []string) bool {
	line := strings.ToLower(desktopID)
	for _, tok := range tokens {
		line += " " + strings.ToLower(filepath.Base(tok))
	}
	for _, name := range []string{"firefox", "zen", "librewolf"} {
		if strings.Contains(line, name) {
			return true
		}
	}
	return false
}
