package linux

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/vinidev/devopt/internal/core"
)

func writeFixtureFile(t *testing.T, dir, name string, size int) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), make([]byte, size), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestYarnDetect(t *testing.T) {
	dir := t.TempDir()
	y := NewYarn(filepath.Join(dir, "missing"))
	found, err := y.Detect()
	if err != nil {
		t.Fatal(err)
	}
	if found {
		t.Fatal("expected Detect to report false for a missing path")
	}

	y2 := NewYarn(dir)
	found, err = y2.Detect()
	if err != nil {
		t.Fatal(err)
	}
	if !found {
		t.Fatal("expected Detect to report true for an existing dir")
	}
}

func TestYarnCalculateSumsFileSizes(t *testing.T) {
	dir := t.TempDir()
	writeFixtureFile(t, dir, "a.bin", 1000)
	writeFixtureFile(t, dir, "b.bin", 2000)

	y := NewYarn(dir)
	finding, err := y.Calculate()
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

func TestYarnCleanDryRunDoesNotRemove(t *testing.T) {
	orig := core.DefaultMinCleanBytes
	core.DefaultMinCleanBytes = 0
	defer func() { core.DefaultMinCleanBytes = orig }()

	dir := t.TempDir()
	writeFixtureFile(t, dir, "a.bin", 1000)

	y := NewYarn(dir)
	result, err := y.Clean(true)
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

func TestYarnCleanRemovesContent(t *testing.T) {
	orig := core.DefaultMinCleanBytes
	core.DefaultMinCleanBytes = 0
	defer func() { core.DefaultMinCleanBytes = orig }()

	dir := t.TempDir()
	writeFixtureFile(t, dir, "a.bin", 1000)

	y := NewYarn(dir)
	result, err := y.Clean(false)
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

func TestYarnCleanSkipsBelowThreshold(t *testing.T) {
	dir := t.TempDir()
	writeFixtureFile(t, dir, "a.bin", 1000) // well below default 200MB threshold

	y := NewYarn(dir)
	result, err := y.Clean(false)
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
