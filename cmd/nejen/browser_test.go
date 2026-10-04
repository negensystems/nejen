package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestMatchBrowser(t *testing.T) {
	browsers := []*browserApp{
		{id: "zen.desktop", name: "Zen Browser"},
		{id: "firefox.desktop", name: "Firefox"},
		{id: "chromium.desktop", name: "Chromium"},
		{id: "google-chrome.desktop", name: "Google Chrome"},
	}

	cases := []struct {
		query string
		want  string
	}{
		{"zen.desktop", "zen.desktop"},
		{"zen", "zen.desktop"},
		{"Zen", "zen.desktop"},
		{"Zen Browser", "zen.desktop"},
		{"Firefox", "firefox.desktop"},
		{"firefox", "firefox.desktop"},
		{"chrome", "google-chrome.desktop"},
		{"chromium", "chromium.desktop"},
		{"safari", ""},
	}

	for _, c := range cases {
		matched := matchBrowser(browsers, c.query)
		got := ""
		if matched != nil {
			got = matched.id
		}
		if got != c.want {
			t.Errorf("matchBrowser(%q) = %q, want %q", c.query, got, c.want)
		}
	}
}

func TestSaveBrowserEnv(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	b := &browserApp{execTokens: []string{"env", "MOZ_ENABLE_WAYLAND=1", "/usr/bin/firefox"}}
	if err := saveBrowserEnv(b); err != nil {
		t.Fatal(err)
	}

	got, err := os.ReadFile(filepath.Join(home, ".local", "state", "nejen", "browser"))
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "firefox\n" {
		t.Errorf("state file = %q, want %q", got, "firefox\n")
	}
}
