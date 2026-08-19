package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"

	"github.com/BurntSushi/toml"
)

func init() {
	registerCommand("hub", runHub)
}

const (
	// branchPrefix marks a row that opens another level, the same mark the
	// device menus use for an entry that leads somewhere else. leafPrefix is
	// the same width so labels stay in one column.
	branchPrefix = "▸ "
	leafPrefix   = "  "
)

type HubMenu struct {
	Title string    `toml:"title"`
	Items []HubItem `toml:"items"`
}

type HubItem struct {
	Label   string `toml:"label"`
	Menu    string `toml:"menu"`
	Run     string `toml:"run"`
	Each    string `toml:"each"`
	Confirm bool   `toml:"confirm"`
}

func die(msg string) {
	fmt.Fprintf(os.Stderr, "nejen hub: %s\n", msg)
	os.Exit(1)
}

func loadMenus() map[string]HubMenu {
	nejenPath := getNejenPath()

	configHome := os.Getenv("XDG_CONFIG_HOME")
	if configHome == "" {
		home, _ := os.UserHomeDir()
		configHome = filepath.Join(home, ".config")
	}

	layers := []string{
		filepath.Join(nejenPath, "config", "hub.toml"),
		filepath.Join(configHome, "nejen", "hub.toml"),
	}

	menus := make(map[string]HubMenu)

	for _, layer := range layers {
		if _, err := os.Stat(layer); err == nil {
			var config struct {
				Menu map[string]HubMenu `toml:"menu"`
			}
			if _, err := toml.DecodeFile(layer, &config); err == nil {
				for k, v := range config.Menu {
					menus[k] = v
				}
			}
		}
	}

	if _, ok := menus["root"]; !ok {
		die("no [menu.root] found in hub.toml")
	}
	return menus
}

func pick(title string, labels []string) int {
	cmd := exec.Command(nejenSelf(), "open", "launcher", "--dmenu", "--index", "-p", title,
		"--width", "440", "--minheight", "1", "--maxheight", "640")
	cmd.Stdin = strings.NewReader(strings.Join(labels, "\n"))
	out, err := cmd.Output()

	outStr := strings.TrimSpace(string(out))
	if err != nil || outStr == "" {
		return -1
	}

	if idx, err := strconv.Atoi(outStr); err == nil && idx >= 0 && idx < len(labels) {
		return idx
	}

	for i, label := range labels {
		if strings.TrimSpace(label) == outStr {
			return i
		}
	}
	return -1
}

func confirmed(label string) bool {
	return pick(label+"?", []string{"no", "yes"}) == 1
}

func expand(item HubItem) []HubItem {
	out, _ := exec.Command("bash", "-c", item.Each).Output()
	var expanded []HubItem
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			expanded = append(expanded, HubItem{
				Label: line,
				Run:   strings.ReplaceAll(item.Run, "{}", line),
			})
		}
	}
	return expanded
}

func runDetached(command string) {
	cmd := exec.Command("bash", "-c", command)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	cmd.Stdout = nil
	cmd.Stderr = nil
	cmd.Stdin = nil
	cmd.Start()
}

func childrenOf(menus map[string]HubMenu, item HubItem) ([]HubItem, string) {
	if item.Menu != "" {
		if sub, ok := menus[item.Menu]; ok {
			return sub.Items, item.Menu
		}
		die("no such menu: " + item.Menu)
	}
	return expand(item), ""
}

// level is one screen of the hub: the items on show, plus the label that led
// here, which the breadcrumb reads back.
type level struct {
	title    string
	menuName string
	items    []HubItem
}

// breadcrumb renders the path down from the root as the search prompt --
// "NEJEN › THEME › WALLPAPER".
//
// The hub used to expand submenus in place, which meant several branches could
// be open at once and there was no single path to name. It also left every row
// of every open branch on screen, separated only by two spaces of indent per
// level; against the uppercase, tracked-out type the menu is set in, that
// indent is close to invisible, so a deep menu read as one long undifferentiated
// list. One level per screen, named at the top, is what makes position legible.
func breadcrumb(stack []level) string {
	parts := make([]string, 0, len(stack))
	for _, l := range stack {
		parts = append(parts, strings.ToUpper(l.title))
	}
	return strings.Join(parts, " › ")
}

// levelLabels renders one screen's rows: everything that opens another level
// is marked, and leaves are padded to the same width so labels stay in one
// column rather than stepping in and out as the mark comes and goes.
func levelLabels(items []HubItem) []string {
	labels := make([]string, 0, len(items))
	for _, item := range items {
		prefix := leafPrefix
		if item.Menu != "" || item.Each != "" {
			prefix = branchPrefix
		}
		labels = append(labels, prefix+item.Label)
	}
	return labels
}

func show(menus map[string]HubMenu, name string) {
	menu, ok := menus[name]
	if !ok {
		die("no such menu: " + name)
	}

	title := menu.Title
	if title == "" {
		title = name
	}

	stack := []level{{title: title, menuName: name, items: menu.Items}}

	for len(stack) > 0 {
		current := stack[len(stack)-1]

		index := pick(breadcrumb(stack), levelLabels(current.items))
		if index == -1 {
			// Dismissed: step back up, and out of the hub entirely from the
			// root. Escape reading as "back" is the half that makes drilling
			// in cheap -- without it every wrong turn costs a reopen.
			stack = stack[:len(stack)-1]
			continue
		}

		item := current.items[index]

		if item.Menu != "" || item.Each != "" {
			children, childMenu := childrenOf(menus, item)
			if len(children) == 0 {
				// An `each` whose command returned nothing. Pushing an empty
				// level would just flash a blank menu and bounce straight back.
				continue
			}
			if childMenu == "" {
				childMenu = current.menuName
			}
			stack = append(stack, level{
				title:    item.Label,
				menuName: childMenu,
				items:    children,
			})
			continue
		}

		if item.Confirm && !confirmed(item.Label) {
			continue
		}

		runDetached(item.Run)
		return
	}
}

func runHub(args []string) {
	menus := loadMenus()
	start := "root"
	if len(args) > 0 {
		start = args[0]
	}
	show(menus, start)
}
