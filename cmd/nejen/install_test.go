package main

import (
	"os"
	"path/filepath"
	"testing"
)

// The first backup holds the user's pre-NEJEN configuration. Overwriting it on
// a later install would destroy the only copy that cannot be regenerated.
func TestBackupNameForNeverReusesAnExistingBackup(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "hyprland.conf")

	if got, want := backupNameFor(target), target+".pre-nejen.bak"; got != want {
		t.Fatalf("first backup = %q, want %q", got, want)
	}

	if err := os.WriteFile(target+".pre-nejen.bak", []byte("original"), 0644); err != nil {
		t.Fatal(err)
	}
	if got, want := backupNameFor(target), target+".pre-nejen.2.bak"; got != want {
		t.Fatalf("second backup = %q, want %q", got, want)
	}

	if err := os.WriteFile(target+".pre-nejen.2.bak", []byte("second"), 0644); err != nil {
		t.Fatal(err)
	}
	if got, want := backupNameFor(target), target+".pre-nejen.3.bak"; got != want {
		t.Fatalf("third backup = %q, want %q", got, want)
	}
}

func TestBackupForeignPreservesUserContent(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "config")
	if err := os.WriteFile(target, []byte("user's own config"), 0644); err != nil {
		t.Fatal(err)
	}

	if err := backupForeign(target); err != nil {
		t.Fatalf("backupForeign: %v", err)
	}
	if _, err := os.Lstat(target); err == nil {
		t.Error("target should have been moved aside")
	}
	b, err := os.ReadFile(target + ".pre-nejen.bak")
	if err != nil {
		t.Fatalf("backup unreadable: %v", err)
	}
	if string(b) != "user's own config" {
		t.Errorf("backup content = %q", b)
	}
}

// A file NEJEN generated is not the user's, so it is replaced in place rather
// than accumulating a backup on every install.
func TestBackupForeignSkipsGeneratedFiles(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "config")
	if err := os.WriteFile(target, []byte("# "+genMark+"\nsource = x"), 0644); err != nil {
		t.Fatal(err)
	}

	if err := backupForeign(target); err != nil {
		t.Fatalf("backupForeign: %v", err)
	}
	if _, err := os.Lstat(target); err != nil {
		t.Error("generated file should have been left in place")
	}
	if _, err := os.Lstat(target + ".pre-nejen.bak"); err == nil {
		t.Error("generated file should not produce a backup")
	}
}

// writeGen must not replace the target when the backup could not be taken --
// that is exactly the case where the user's file would be lost for good.
func TestWriteGenLeavesFileWhenBackupFails(t *testing.T) {
	dir := t.TempDir()
	sub := filepath.Join(dir, "ro")
	if err := os.Mkdir(sub, 0755); err != nil {
		t.Fatal(err)
	}
	target := filepath.Join(sub, "config")
	if err := os.WriteFile(target, []byte("precious"), 0644); err != nil {
		t.Fatal(err)
	}
	// A read-only parent makes the rename (and any write) fail.
	if err := os.Chmod(sub, 0500); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chmod(sub, 0755) })

	writeGen(target, "# "+genMark+"\nreplacement")

	b, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("target vanished: %v", err)
	}
	if string(b) != "precious" {
		t.Errorf("user content was destroyed: %q", b)
	}
}

func writeManifest(t *testing.T, root, name string) string {
	t.Helper()
	dir := filepath.Join(root, "packages")
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, name+".txt")
	if err := os.WriteFile(path, []byte("## Official repos\nbase\n"), 0644); err != nil {
		t.Fatal(err)
	}
	return path
}

// A fresh clone has its manifest in the source tree and nothing in the
// installed tree yet, so the source copy has to win.
func TestResolveManifestPrefersTheSourceTree(t *testing.T) {
	source, nejen := t.TempDir(), t.TempDir()
	want := writeManifest(t, source, "core")
	writeManifest(t, nejen, "core")

	if got := resolveManifest("core", source, nejen); got != want {
		t.Errorf("resolveManifest = %q, want the source-tree copy %q", got, want)
	}
}

func TestResolveManifestFallsBackToTheInstalledTree(t *testing.T) {
	source, nejen := t.TempDir(), t.TempDir()
	want := writeManifest(t, nejen, "core")

	if got := resolveManifest("core", source, nejen); got != want {
		t.Errorf("resolveManifest = %q, want %q", got, want)
	}
}

func TestResolveManifestReportsMissing(t *testing.T) {
	if got := resolveManifest("nope", t.TempDir(), t.TempDir()); got != "" {
		t.Errorf("resolveManifest = %q, want \"\" for an unknown manifest", got)
	}
}

func TestResolveManifestAcceptsAnExplicitPath(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "my-list.txt")
	if err := os.WriteFile(path, []byte("## Official repos\nbase\n"), 0644); err != nil {
		t.Fatal(err)
	}

	if got := resolveManifest(path, "", ""); got != path {
		t.Errorf("resolveManifest = %q, want %q", got, path)
	}
	if got := resolveManifest(filepath.Join(dir, "absent.txt"), "", ""); got != "" {
		t.Errorf("resolveManifest = %q, want \"\" for a path that does not exist", got)
	}
}

// A bare name is looked up under packages/ and nowhere else, so a directory
// called "core" next to wherever install.sh ran cannot shadow the manifest.
func TestResolveManifestIgnoresWorkingDirectoryCollisions(t *testing.T) {
	source, nejen := t.TempDir(), t.TempDir()
	want := writeManifest(t, nejen, "core")

	t.Chdir(t.TempDir())
	if err := os.Mkdir("core", 0755); err != nil {
		t.Fatal(err)
	}

	if got := resolveManifest("core", source, nejen); got != want {
		t.Errorf("resolveManifest = %q, want %q -- a stray \"core\" directory hijacked the name", got, want)
	}
}

// paru wins when both are installed, and "" means none -- never a guess.
func TestAURHelperPrefersParuAndReportsNone(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("PATH", dir)

	if got := aurHelper(); got != "" {
		t.Errorf("aurHelper = %q on a PATH with no helper, want %q", got, "")
	}

	for _, name := range []string{"yay", "paru"} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte("#!/bin/sh\n"), 0755); err != nil {
			t.Fatal(err)
		}
	}
	if got := aurHelper(); got != "paru" {
		t.Errorf("aurHelper = %q with both installed, want %q", got, "paru")
	}

	if err := os.Remove(filepath.Join(dir, "paru")); err != nil {
		t.Fatal(err)
	}
	if got := aurHelper(); got != "yay" {
		t.Errorf("aurHelper = %q with only yay installed, want %q", got, "yay")
	}
}
