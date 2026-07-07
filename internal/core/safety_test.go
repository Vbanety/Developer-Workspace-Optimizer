package core

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGuardRejectsNeverDeletePaths(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatal(err)
	}
	target := filepath.Join(home, "Documents", "something")
	if err := Guard(target); err == nil {
		t.Fatalf("expected Guard to reject %s, got nil error", target)
	}
}

func TestGuardRejectsProjectDirectory(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module x\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := Guard(dir); err == nil {
		t.Fatalf("expected Guard to reject project dir %s, got nil error", dir)
	}
}

func TestGuardAllowsPlainCacheDir(t *testing.T) {
	dir := t.TempDir()
	if err := Guard(dir); err != nil {
		t.Fatalf("expected Guard to allow plain cache dir %s, got %v", dir, err)
	}
}

func TestShouldSkipSmall(t *testing.T) {
	if !ShouldSkipSmall(1024) {
		t.Fatal("expected 1KB to be skipped")
	}
	if ShouldSkipSmall(300 * 1024 * 1024) {
		t.Fatal("expected 300MB to not be skipped")
	}
}
