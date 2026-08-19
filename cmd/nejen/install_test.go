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
