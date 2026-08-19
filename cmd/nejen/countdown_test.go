package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func mustDate(t *testing.T, s string) time.Time {
	t.Helper()
	d, err := time.ParseInLocation("2006-01-02", s, time.Local)
	if err != nil {
		t.Fatal(err)
	}
	return d
}

func TestCalendarDaysBetween(t *testing.T) {
	cases := []struct {
		from, to string
		want     int
	}{
		{"2026-08-18", "2026-08-22", 4},
		{"2026-08-18", "2026-08-18", 0},
		{"2026-08-18", "2026-08-17", -1},
		// Across a year boundary.
		{"2026-12-30", "2027-01-02", 3},
		// Across the March DST spring-forward (a 23-hour local day in
		// zones that observe it): must still count calendar days.
		{"2026-03-07", "2026-03-09", 2},
		// Across the November fall-back (a 25-hour local day).
		{"2026-10-31", "2026-11-02", 2},
	}
	for _, c := range cases {
		got := calendarDaysBetween(mustDate(t, c.from), mustDate(t, c.to))
		if got != c.want {
			t.Errorf("calendarDaysBetween(%s, %s) = %d, want %d", c.from, c.to, got, c.want)
		}
	}
}

func TestDaysInfo(t *testing.T) {
	today := mustDate(t, "2026-08-18")
	cases := []struct {
		name              string
		event             countdownEvent
		marker, css, stat string
	}{
		{"future", countdownEvent{"date": "2026-08-22"}, "D-4", "future", "4 day(s) left"},
		{"today", countdownEvent{"date": "2026-08-18"}, "D-Day", "today", "Today"},
		{"yesterday", countdownEvent{"date": "2026-08-17"}, "D+1", "past", "1 day(s) passed"},
		{"missing", countdownEvent{}, "NO-DATE", "error", "Missing date"},
		{"malformed", countdownEvent{"date": "22/08/2026"}, "BAD-DATE", "error", "Invalid date format"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			marker, css, stat := daysInfo(c.event, today)
			if marker != c.marker || css != c.css || stat != c.stat {
				t.Errorf("daysInfo = (%q, %q, %q), want (%q, %q, %q)",
					marker, css, stat, c.marker, c.css, c.stat)
			}
		})
	}
}

func TestSortedCountdownEventsOrder(t *testing.T) {
	today := mustDate(t, "2026-08-18")
	events := []countdownEvent{
		{"name": "long past", "date": "2026-01-01"},
		{"name": "undated"},
		{"name": "far future", "date": "2026-12-25"},
		{"name": "soon", "date": "2026-08-20"},
		{"name": "just past", "date": "2026-08-16"},
	}
	got := sortedCountdownEvents(events, today)
	// Undated events sort as infinitely far in the future (end of the
	// upcoming group, before anything past) -- checked against the
	// python implementation this ports.
	want := []string{"soon", "far future", "undated", "just past", "long past"}
	for i, w := range want {
		if got[i].name() != w {
			t.Errorf("position %d = %q, want %q", i, got[i].name(), w)
		}
	}
}

func TestCountdownConfigRoundTripPreservesUnknownKeys(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "countdown.json")
	src := `{"events": [{"name": "x", "date": "2026-08-22", "color": "#ff0000"}], "theme": "dark"}`
	if err := os.WriteFile(path, []byte(src), 0644); err != nil {
		t.Fatal(err)
	}

	var raw map[string]interface{}
	if err := json.Unmarshal([]byte(src), &raw); err != nil {
		t.Fatal(err)
	}
	cfg := &countdownConfig{path: path, raw: raw}
	for _, item := range raw["events"].([]interface{}) {
		cfg.events = append(cfg.events, countdownEvent(item.(map[string]interface{})))
	}

	cfg.events[0]["name"] = "renamed"
	if err := cfg.save(); err != nil {
		t.Fatal(err)
	}

	b, _ := os.ReadFile(path)
	var back map[string]interface{}
	if err := json.Unmarshal(b, &back); err != nil {
		t.Fatal(err)
	}
	if back["theme"] != "dark" {
		t.Error("top-level unknown key dropped on save")
	}
	ev := back["events"].([]interface{})[0].(map[string]interface{})
	if ev["color"] != "#ff0000" {
		t.Error("event-level unknown key dropped on save")
	}
	if ev["name"] != "renamed" {
		t.Error("edit not persisted")
	}
}

func TestWaybarEventSelection(t *testing.T) {
	a := countdownEvent{"name": "a"}
	b := countdownEvent{"name": "b", "show_on_waybar": true}
	if got := waybarEvent([]countdownEvent{a, b}); got.name() != "b" {
		t.Errorf("flagged event not selected, got %q", got.name())
	}
	if got := waybarEvent([]countdownEvent{a}); got.name() != "a" {
		t.Errorf("fallback to first event failed, got %q", got.name())
	}
	if waybarEvent(nil) != nil {
		t.Error("empty list should return nil")
	}
}
