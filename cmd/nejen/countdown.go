package main

// Waybar countdown: day counters toward (or since) dated events, with a
// walker-driven management menu on click. This is the Go port of the vendored
// waybar-countdown python script (github.com/takumiymd/waybar-countdown), the
// last interpreter-dependent module in the shipped bar.
//
// Events live in ~/.config/nejen/countdown.json; an existing
// ~/.config/waybar-countdown/events.json from the standalone widget is read
// (and written back) when the nejen path does not exist yet, so nothing is
// lost on migration. Events are kept as raw maps rather than structs so keys
// this port does not know about survive a rewrite.

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

func init() {
	registerCommand("countdown", runCountdown)
	registerCommand("countdown menu", runCountdownMenu)
}

type countdownEvent map[string]interface{}

type countdownConfig struct {
	path   string
	raw    map[string]interface{}
	events []countdownEvent
}

func (e countdownEvent) str(key string) string {
	if v, ok := e[key].(string); ok {
		return v
	}
	return ""
}

func (e countdownEvent) name() string {
	if n := e.str("name"); n != "" {
		return n
	}
	return "Unnamed"
}

func countdownConfigPath() (string, bool) {
	configHome := os.Getenv("XDG_CONFIG_HOME")
	if configHome == "" {
		home, _ := os.UserHomeDir()
		configHome = filepath.Join(home, ".config")
	}
	nejenPath := filepath.Join(configHome, "nejen", "countdown.json")
	if _, err := os.Stat(nejenPath); err == nil {
		return nejenPath, true
	}
	legacy := filepath.Join(configHome, "waybar-countdown", "events.json")
	if _, err := os.Stat(legacy); err == nil {
		return legacy, true
	}
	return nejenPath, false
}

func countdownError(message string) {
	out, _ := json.Marshal(map[string]string{
		"text": "countdown error", "tooltip": message, "class": "error",
	})
	fmt.Println(string(out))
	os.Exit(0)
}

func loadCountdownConfig() *countdownConfig {
	path, exists := countdownConfigPath()
	if !exists {
		countdownError("Config file not found:\n" + path)
	}
	b, err := os.ReadFile(path)
	if err != nil {
		countdownError("Config file not readable:\n" + err.Error())
	}
	var raw map[string]interface{}
	if err := json.Unmarshal(b, &raw); err != nil {
		countdownError("Invalid JSON:\n" + err.Error())
	}
	cfg := &countdownConfig{path: path, raw: raw}
	if list, ok := raw["events"].([]interface{}); ok {
		for _, item := range list {
			if m, ok := item.(map[string]interface{}); ok {
				cfg.events = append(cfg.events, countdownEvent(m))
			}
		}
	}
	return cfg
}

func (c *countdownConfig) save() error {
	list := make([]interface{}, len(c.events))
	for i, e := range c.events {
		list[i] = map[string]interface{}(e)
	}
	c.raw["events"] = list
	b, err := json.MarshalIndent(c.raw, "", "  ")
	if err != nil {
		return err
	}
	os.MkdirAll(filepath.Dir(c.path), 0755)
	return os.WriteFile(c.path, append(b, '\n'), 0644)
}

// daysInfo mirrors the python get_days_info: marker, CSS class, and a status
// line for one event, relative to today.
func daysInfo(e countdownEvent, today time.Time) (marker, cssClass, status string) {
	dateText := e.str("date")
	if dateText == "" {
		return "NO-DATE", "error", "Missing date"
	}
	target, err := time.ParseInLocation("2006-01-02", dateText, time.Local)
	if err != nil {
		return "BAD-DATE", "error", "Invalid date format"
	}
	days := calendarDaysBetween(today, target)
	switch {
	case days == 0:
		return "D-Day", "today", "Today"
	case days > 0:
		return fmt.Sprintf("D-%d", days), "future", fmt.Sprintf("%d day(s) left", days)
	default:
		return fmt.Sprintf("D+%d", -days), "past", fmt.Sprintf("%d day(s) passed", -days)
	}
}

// calendarDaysBetween counts whole calendar days from a to b, the way
// python's date subtraction does. Both dates are re-anchored to UTC midnight
// before dividing, so a DST transition (a 23- or 25-hour local day) cannot
// skew the count.
func calendarDaysBetween(a, b time.Time) int {
	au := time.Date(a.Year(), a.Month(), a.Day(), 0, 0, 0, 0, time.UTC)
	bu := time.Date(b.Year(), b.Month(), b.Day(), 0, 0, 0, 0, time.UTC)
	return int(bu.Sub(au).Hours() / 24)
}

// localToday returns midnight today in local time, the reference point every
// day count is computed against.
func localToday() time.Time {
	now := time.Now()
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)
}

// sortedCountdownEvents orders upcoming events soonest-first, then past
// events most-recent-first, with undated events last -- the same order the
// python version presented.
func sortedCountdownEvents(events []countdownEvent, today time.Time) []countdownEvent {
	ordered := make([]countdownEvent, len(events))
	copy(ordered, events)
	sort.SliceStable(ordered, func(i, j int) bool {
		gi, oi := countdownSortKey(ordered[i], today)
		gj, oj := countdownSortKey(ordered[j], today)
		if gi != gj {
			return gi < gj
		}
		if oi != oj {
			return oi < oj
		}
		return ordered[i].name() < ordered[j].name()
	})
	return ordered
}

func countdownSortKey(e countdownEvent, today time.Time) (group, order int) {
	target, err := time.ParseInLocation("2006-01-02", e.str("date"), time.Local)
	if err != nil {
		// An undated event sorts as infinitely far in the future -- the end
		// of the upcoming group, ahead of everything already past.
		return 0, int(^uint(0) >> 1)
	}
	days := calendarDaysBetween(today, target)
	if days >= 0 {
		return 0, days
	}
	return 1, -days
}

func waybarEvent(events []countdownEvent) countdownEvent {
	for _, e := range events {
		if shown, ok := e["show_on_waybar"].(bool); ok && shown {
			return e
		}
	}
	if len(events) > 0 {
		return events[0]
	}
	return nil
}

func runCountdown(args []string) {
	cfg := loadCountdownConfig()
	if len(cfg.events) == 0 {
		countdownError("No events found in " + filepath.Base(cfg.path))
	}
	e := waybarEvent(cfg.events)
	today := localToday()
	marker, cssClass, status := daysInfo(e, today)

	tooltip := strings.Join([]string{
		e.name(),
		"Date: " + e.str("date"),
		"Status: " + status,
		"Today: " + today.Format("2006-01-02"),
	}, "\n")

	out, _ := json.Marshal(map[string]string{
		"text":    fmt.Sprintf("%s: %s", e.name(), marker),
		"tooltip": tooltip,
		"class":   cssClass,
	})
	fmt.Println(string(out))
}

// countdownMenu runs a selection through walker's dmenu mode (the NEJEN
// launcher), with wofi kept as a fallback for setups that still have it.
func countdownMenu(menuText, prompt, width, height string) string {
	var cmd *exec.Cmd
	if _, err := exec.LookPath("walker"); err == nil {
		cmd = exec.Command("walker", "--dmenu", "--width", width, "--maxheight", height, "-p", prompt)
	} else if _, err := exec.LookPath("wofi"); err == nil {
		cmd = exec.Command("wofi", "--dmenu", "--prompt", prompt, "--width", width, "--height", height)
	} else {
		fmt.Println(menuText)
		return ""
	}
	cmd.Stdin = strings.NewReader(menuText)
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

// askCountdownValue collects free text. walker's dmenu is selection-only, so
// text entry needs zenity (preferred) or wofi.
func askCountdownValue(prompt, current string) string {
	if _, err := exec.LookPath("zenity"); err == nil {
		out, err := exec.Command("zenity", "--entry", "--title", "nejen countdown",
			"--text", prompt, "--entry-text", current).Output()
		if err != nil {
			return ""
		}
		return strings.TrimSpace(string(out))
	}
	if _, err := exec.LookPath("wofi"); err == nil {
		cmd := exec.Command("wofi", "--dmenu", "--prompt", prompt, "--width", "500", "--height", "160")
		cmd.Stdin = strings.NewReader(current)
		out, err := cmd.Output()
		if err != nil {
			return ""
		}
		return strings.TrimSpace(string(out))
	}
	exec.Command("notify-send", "nejen countdown", "Text entry needs zenity or wofi installed.").Run()
	return ""
}

func validCountdownDate(s string) bool {
	_, err := time.ParseInLocation("2006-01-02", s, time.Local)
	return err == nil
}

func runCountdownMenu(args []string) {
	cfg := loadCountdownConfig()
	today := localToday()
	ordered := sortedCountdownEvents(cfg.events, today)

	lines := []string{
		"+ Add countdown",
		"",
		"NO  DAYS      TARGET DATE   START DATE    WAYBAR   NAME",
		"────────────────────────────────────────────────────────────",
	}
	for i, e := range ordered {
		marker, _, _ := daysInfo(e, today)
		dateText := e.str("date")
		if dateText == "" {
			dateText = "no-date"
		}
		startText := e.str("start")
		if startText == "" {
			startText = "-"
		}
		shown := "-"
		if s, ok := e["show_on_waybar"].(bool); ok && s {
			shown = "shown"
		}
		lines = append(lines, fmt.Sprintf("%02d  %-9s %-12s %-12s %-8s %s",
			i+1, marker, dateText, startText, shown, e.name()))
	}

	selected := countdownMenu(strings.Join(lines, "\n"), "Countdowns", "560", "360")
	if selected == "" {
		return
	}
	if selected == "+ Add countdown" {
		addCountdown(cfg)
		return
	}

	first := strings.Fields(selected)
	if len(first) == 0 {
		return
	}
	index, err := strconv.Atoi(first[0])
	if err != nil || index < 1 || index > len(ordered) {
		return
	}
	manageCountdown(cfg, ordered[index-1])
}

func addCountdown(cfg *countdownConfig) {
	name := askCountdownValue("Countdown name", "")
	if name == "" {
		return
	}
	target := askCountdownValue("Target date: YYYY-MM-DD", "")
	if target == "" {
		return
	}
	if !validCountdownDate(target) {
		exec.Command("notify-send", "Countdown error", "Invalid target date. Use YYYY-MM-DD.").Run()
		return
	}
	start := askCountdownValue("Start date: YYYY-MM-DD or leave empty for today", "")
	if start == "" {
		start = localToday().Format("2006-01-02")
	}
	if !validCountdownDate(start) {
		exec.Command("notify-send", "Countdown error", "Invalid start date. Use YYYY-MM-DD.").Run()
		return
	}
	cfg.events = append(cfg.events, countdownEvent{"name": name, "date": target, "start": start})
	cfg.save()
	exec.Command("notify-send", "Countdown added", name+" added").Run()
}

func manageCountdown(cfg *countdownConfig, e countdownEvent) {
	today := localToday()
	marker, _, status := daysInfo(e, today)
	target := e.str("date")
	if target == "" {
		target = "no-date"
	}
	start := e.str("start")
	if start == "" {
		start = "-"
	}
	shownLine := "not shown on Waybar"
	if s, ok := e["show_on_waybar"].(bool); ok && s {
		shownLine = "shown on Waybar"
	}

	actions := strings.Join([]string{
		"Countdown: " + e.name(),
		"Days:      " + marker,
		"Status:    " + status,
		"Target:    " + target,
		"Start:     " + start,
		"Waybar:    " + shownLine,
		"",
		"Show on Waybar",
		"Edit name",
		"Edit target date",
		"Edit start date",
		"Delete",
	}, "\n")

	selected := countdownMenu(actions, "Manage: "+e.name(), "520", "420")
	switch {
	case selected == "":
		return
	case strings.HasPrefix(selected, "Countdown:"), strings.HasPrefix(selected, "Days:"),
		strings.HasPrefix(selected, "Status:"), strings.HasPrefix(selected, "Target:"),
		strings.HasPrefix(selected, "Start:"), strings.HasPrefix(selected, "Waybar:"):
		return
	case selected == "Show on Waybar":
		for _, other := range cfg.events {
			delete(other, "show_on_waybar")
		}
		e["show_on_waybar"] = true
		cfg.save()
		exec.Command("notify-send", "Countdown", "Showing on Waybar: "+e.name()).Run()
	case selected == "Edit name":
		if v := askCountdownValue("Edit name\nCurrent: "+e.name(), e.name()); v != "" {
			e["name"] = v
			cfg.save()
			exec.Command("notify-send", "Countdown updated", "Name updated").Run()
		}
	case selected == "Edit target date":
		v := askCountdownValue("Edit target date\nCurrent: "+target, target)
		if v == "" {
			return
		}
		if !validCountdownDate(v) {
			exec.Command("notify-send", "Countdown error", "Invalid date. Use YYYY-MM-DD.").Run()
			return
		}
		e["date"] = v
		cfg.save()
		exec.Command("notify-send", "Countdown updated", "Target date updated").Run()
	case selected == "Edit start date":
		cur := start
		if cur == "-" {
			cur = ""
		}
		v := askCountdownValue("Edit start date\nCurrent: "+start, cur)
		if v == "" {
			delete(e, "start")
			cfg.save()
			exec.Command("notify-send", "Countdown updated", "Start date removed").Run()
			return
		}
		if !validCountdownDate(v) {
			exec.Command("notify-send", "Countdown error", "Invalid start date. Use YYYY-MM-DD.").Run()
			return
		}
		e["start"] = v
		cfg.save()
		exec.Command("notify-send", "Countdown updated", "Start date updated").Run()
	case selected == "Delete":
		if countdownMenu("Cancel\nConfirm delete", "Confirm delete: "+e.name(), "420", "160") != "Confirm delete" {
			return
		}
		for i, other := range cfg.events {
			if sameCountdownEvent(other, e) {
				cfg.events = append(cfg.events[:i], cfg.events[i+1:]...)
				break
			}
		}
		cfg.save()
		exec.Command("notify-send", "Countdown deleted", e.name()).Run()
	}
}

// sameCountdownEvent identifies an event by map identity: events are shared
// references between cfg.events and the sorted view, so pointer-like equality
// via a marker key is unnecessary -- comparing the underlying maps' addresses
// is not possible directly, but two references to the same map stay equal
// under fmt %p formatting.
func sameCountdownEvent(a, b countdownEvent) bool {
	return fmt.Sprintf("%p", map[string]interface{}(a)) == fmt.Sprintf("%p", map[string]interface{}(b))
}
