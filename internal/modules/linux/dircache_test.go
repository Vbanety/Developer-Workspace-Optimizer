package linux

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Vbanety/Developer-Workspace-Optimizer/internal/core"
)

func writeFixtureFile(t *testing.T, dir, name string, size int) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), make([]byte, size), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestDirCacheDetect(t *testing.T) {
	dir := t.TempDir()
	c := NewDirCache("yarn", filepath.Join(dir, "missing"), true)
	found, err := c.Detect()
	if err != nil {
		t.Fatal(err)
	}
	if found {
		t.Fatal("expected Detect to report false for a missing path")
	}

	c2 := NewDirCache("yarn", dir, true)
	found, err = c2.Detect()
	if err != nil {
		t.Fatal(err)
	}
	if !found {
		t.Fatal("expected Detect to report true for an existing dir")
	}
}

func TestDirCacheCalculateSumsFileSizes(t *testing.T) {
	dir := t.TempDir()
	writeFixtureFile(t, dir, "a.bin", 1000)
	writeFixtureFile(t, dir, "b.bin", 2000)

	c := NewDirCache("yarn", dir, true)
	finding, err := c.Calculate()
	if err != nil {
		t.Fatal(err)
	}
	if finding.SizeBytes != 3000 {
		t.Fatalf("expected 3000 bytes, got %d", finding.SizeBytes)
	}
	if finding.ItemCount != 2 {
		t.Fatalf("expected 2 items, got %d", finding.ItemCount)
	}
}

func TestDirCacheCleanDryRunDoesNotRemove(t *testing.T) {
	orig := core.DefaultMinCleanBytes
	core.DefaultMinCleanBytes = 0
	defer func() { core.DefaultMinCleanBytes = orig }()

	dir := t.TempDir()
	writeFixtureFile(t, dir, "a.bin", 1000)

	c := NewDirCache("yarn", dir, true)
	result, err := c.Clean(true)
	if err != nil {
		t.Fatal(err)
	}
	if !result.DryRun {
		t.Fatal("expected DryRun to be true")
	}
	if result.FreedBytes != 1000 {
		t.Fatalf("expected FreedBytes to report 1000, got %d", result.FreedBytes)
	}
	if _, err := os.Stat(filepath.Join(dir, "a.bin")); err != nil {
		t.Fatalf("expected file to still exist after dry-run, stat error: %v", err)
	}
}

func TestDirCacheCleanRemovesContent(t *testing.T) {
	orig := core.DefaultMinCleanBytes
	core.DefaultMinCleanBytes = 0
	defer func() { core.DefaultMinCleanBytes = orig }()

	dir := t.TempDir()
	writeFixtureFile(t, dir, "a.bin", 1000)

	c := NewDirCache("yarn", dir, true)
	result, err := c.Clean(false)
	if err != nil {
		t.Fatal(err)
	}
	if result.FreedBytes != 1000 {
		t.Fatalf("expected FreedBytes to report 1000, got %d", result.FreedBytes)
	}
	if _, err := os.Stat(filepath.Join(dir, "a.bin")); !os.IsNotExist(err) {
		t.Fatalf("expected file to be removed, stat error: %v", err)
	}
}

func TestDirCacheCleanSkipsBelowThreshold(t *testing.T) {
	dir := t.TempDir()
	writeFixtureFile(t, dir, "a.bin", 1000) // well below default 200MB threshold

	c := NewDirCache("yarn", dir, true)
	result, err := c.Clean(false)
	if err != nil {
		t.Fatal(err)
	}
	if !result.Skipped {
		t.Fatal("expected small cache to be skipped")
	}
	if _, err := os.Stat(filepath.Join(dir, "a.bin")); err != nil {
		t.Fatalf("expected file to still exist when skipped, stat error: %v", err)
	}
}

func TestDirCacheCalculateSkipsUnreadableSubdir(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("root ignores permission bits, can't exercise this case")
	}

	dir := t.TempDir()
	writeFixtureFile(t, dir, "readable.bin", 1000)

	blocked := filepath.Join(dir, "partial")
	if err := os.Mkdir(blocked, 0o755); err != nil {
		t.Fatal(err)
	}
	writeFixtureFile(t, blocked, "inside.bin", 5000)
	if err := os.Chmod(blocked, 0o000); err != nil {
		t.Fatal(err)
	}
	defer os.Chmod(blocked, 0o755) // let t.TempDir() clean up afterward

	c := NewDirCache("apt", dir, false)
	finding, err := c.Calculate()
	if err != nil {
		t.Fatalf("expected unreadable subdir to be skipped, not fail the scan: %v", err)
	}
	if finding.SizeBytes != 1000 {
		t.Fatalf("expected only the readable file's 1000 bytes, got %d", finding.SizeBytes)
	}
}

func TestDirCacheGuardBlocksNeverDeletePath(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatal(err)
	}
	c := NewDirCache("docs", filepath.Join(home, "Documents"), true)
	if _, err := c.Clean(true); err == nil {
		t.Fatal("expected Clean to be rejected by the never-delete guard")
	}
}
