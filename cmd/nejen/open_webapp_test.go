package main

import (
	"reflect"
	"testing"
)

func TestParseDesktopExec(t *testing.T) {
	cases := []struct {
		in   string
		want []string
	}{
		{"/usr/bin/chromium %U", []string{"/usr/bin/chromium"}},
		{"zen-browser --new-window %u", []string{"zen-browser", "--new-window"}},
		{"browser --url=%u --class=web", []string{"browser", "--class=web"}},
		{
			"env MOZ_ENABLE_WAYLAND=0 /opt/zen-browser-bin/zen-bin %u",
			[]string{"env", "MOZ_ENABLE_WAYLAND=0", "/opt/zen-browser-bin/zen-bin"},
		},
		{`"/opt/zen browser/zen" --flag %u`, []string{"/opt/zen browser/zen", "--flag"}},
		{`browser --title="say \"hi\""`, []string{"browser", `--title=say "hi"`}},
		{"browser --zoom=100%%", []string{"browser", "--zoom=100%"}},
		{"browser --name='my app'", []string{"browser", "--name='my", "app'"}},
		{"  browser\t%U  ", []string{"browser"}},
	}

	for _, c := range cases {
		got := parseDesktopExec(c.in)
		if !reflect.DeepEqual(got, c.want) {
			t.Errorf("parseDesktopExec(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestIsGeckoBrowser(t *testing.T) {
	cases := []struct {
		desktopID string
		tokens    []string
		want      bool
	}{
		{"zen.desktop", []string{"/opt/zen-browser-bin/zen-bin"}, true},
		{"browser.desktop", []string{"/usr/bin/firefox"}, true},
		{"chromium.desktop", []string{"/usr/bin/chromium"}, false},
		{"google-chrome.desktop", []string{"/usr/bin/google-chrome-stable"}, false},
	}

	for _, c := range cases {
		if got := isGeckoBrowser(c.desktopID, c.tokens); got != c.want {
			t.Errorf("isGeckoBrowser(%q, %v) = %v, want %v", c.desktopID, c.tokens, got, c.want)
		}
	}
}
