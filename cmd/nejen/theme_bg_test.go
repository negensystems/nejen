package main

import (
	"os"
	"path/filepath"
	"testing"
)

// stageWallpapers points HOME and XDG_STATE_HOME at a scratch tree and
// creates the named files under the user wallpaper folder. WAYLAND_DISPLAY
// is cleared so applyBg never tries to talk to a compositor.
func stageWallpapers(t *testing.T, names ...string) string {
	t.Helper()

	root := t.TempDir()
	t.Setenv("HOME", root)
	t.Setenv("XDG_STATE_HOME", filepath.Join(root, "state"))
	t.Setenv("WAYLAND_DISPLAY", "")

	dir := bgUserDir()
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	for _, n := range names {
		if err := os.WriteFile(filepath.Join(dir, n), []byte("x"), 0644); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

func TestBgListSkipsNonImages(t *testing.T) {
	stageWallpapers(t, "a.png", "b.JPG", "notes.txt", "archive.tar.gz")

	got := bgList()
	if len(got) != 2 {
		t.Fatalf("expected 2 wallpapers, got %d: %v", len(got), got)
	}
	if filepath.Base(got[0]) != "a.png" || filepath.Base(got[1]) != "b.JPG" {
		t.Errorf("unexpected list: %v", got)
	}
}

func TestBgCycleWrapsBothWays(t *testing.T) {
	stageWallpapers(t, "1.png", "2.png", "3.png")

	// Forward from the last entry wraps to the first.
	if err := applyBg(filepath.Join(bgUserDir(), "3.png")); err != nil {
		t.Fatal(err)
	}
	cycleBg(+1)
	if got := filepath.Base(currentBg()); got != "1.png" {
		t.Errorf("next from last = %s, want 1.png", got)
	}

	// Backward from the first entry wraps to the last.
	cycleBg(-1)
	if got := filepath.Base(currentBg()); got != "3.png" {
		t.Errorf("prev from first = %s, want 3.png", got)
	}
}

func TestBgCycleStartsAtTopWhenCurrentIsUnknown(t *testing.T) {
	stageWallpapers(t, "1.png", "2.png")

	cycleBg(+1) // nothing set yet
	if got := filepath.Base(currentBg()); got != "1.png" {
		t.Errorf("next with no current = %s, want 1.png", got)
	}
}

func TestBgIsUserOwned(t *testing.T) {
	stageWallpapers(t, "mine.png")

	if !bgIsUserOwned(filepath.Join(bgUserDir(), "mine.png")) {
		t.Error("wallpaper in the user folder should be user-owned")
	}
	if bgIsUserOwned(filepath.Join(t.TempDir(), "theme.png")) {
		t.Error("wallpaper outside the user folder should not be user-owned")
	}
	if bgIsUserOwned("") {
		t.Error("empty path should not be user-owned")
	}
}

func TestResolveBgAcceptsNameAndLabel(t *testing.T) {
	dir := stageWallpapers(t, "10-misty-ridge.jpg")
	want := filepath.Join(dir, "10-misty-ridge.jpg")

	for _, arg := range []string{want, "10-misty-ridge.jpg", "10-misty-ridge", "Misty Ridge", "misty ridge"} {
		if got := resolveBg(arg); got != want {
			t.Errorf("resolveBg(%q) = %q, want %q", arg, got, want)
		}
	}
	if got := resolveBg("nothing-like-this"); got != "" {
		t.Errorf("resolveBg on an unknown name = %q, want empty", got)
	}
}

func TestBgLabel(t *testing.T) {
	cases := map[string]string{
		"/x/10-misty-ridge.jpg": "Misty Ridge",
		"/x/deep_forest.png":    "Deep Forest",
		"/x/SUNSET.webp":        "Sunset",
		"/x/2024.jpg":           "2024.jpg", // all digits: fall back to the filename
	}
	for in, want := range cases {
		if got := bgLabel(in); got != want {
			t.Errorf("bgLabel(%q) = %q, want %q", in, got, want)
		}
	}
}
