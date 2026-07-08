package linux

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/vinidev/devopt/internal/core"
)

func TestMultiDirCacheDetectFindsAnyExistingSubpath(t *testing.T) {
	root := t.TempDir()
	if err := os.Mkdir(filepath.Join(root, "Cache"), 0o755); err != nil {
		t.Fatal(err)
	}

	c := NewMultiDirCache("vscode", root, []string{"Cache", "GPUCache", "missing"}, true)
	found, err := c.Detect()
	if err != nil {
		t.Fatal(err)
	}
	if !found {
		t.Fatal("expected Detect to report true when at least one subpath exists")
	}
}

func TestMultiDirCacheDetectFalseWhenNoneExist(t *testing.T) {
	root := t.TempDir()
	c := NewMultiDirCache("vscode", root, []string{"Cache", "GPUCache"}, true)
	found, err := c.Detect()
	if err != nil {
		t.Fatal(err)
	}
	if found {
		t.Fatal("expected Detect to report false when no subpath exists")
	}
}

func TestMultiDirCacheCalculateSumsExistingSubpathsOnly(t *testing.T) {
	root := t.TempDir()
	mustMkdir(t, filepath.Join(root, "Cache"))
	mustMkdir(t, filepath.Join(root, "GPUCache"))
	writeFixtureFile(t, filepath.Join(root, "Cache"), "a.bin", 1000)
	writeFixtureFile(t, filepath.Join(root, "GPUCache"), "b.bin", 2000)
	// "logs" subpath intentionally not created — Calculate must not error on it.

	c := NewMultiDirCache("vscode", root, []string{"Cache", "GPUCache", "logs"}, true)
	finding, err := c.Calculate()
	if err != nil {
		t.Fatal(err)
	}
	if finding.SizeBytes != 3000 {
		t.Fatalf("expected 3000 bytes, got %d", finding.SizeBytes)
	}
}

func TestMultiDirCacheCleanDryRunDoesNotRemove(t *testing.T) {
	orig := core.DefaultMinCleanBytes
	core.DefaultMinCleanBytes = 0
	defer func() { core.DefaultMinCleanBytes = orig }()

	root := t.TempDir()
	mustMkdir(t, filepath.Join(root, "Cache"))
	writeFixtureFile(t, filepath.Join(root, "Cache"), "a.bin", 1000)

	c := NewMultiDirCache("vscode", root, []string{"Cache"}, true)
	result, err := c.Clean(true)
	if err != nil {
		t.Fatal(err)
	}
	if !result.DryRun || result.FreedBytes != 1000 {
		t.Fatalf("unexpected dry-run result: %+v", result)
	}
	if _, err := os.Stat(filepath.Join(root, "Cache", "a.bin")); err != nil {
		t.Fatalf("expected file to still exist after dry-run: %v", err)
	}
}

func TestMultiDirCacheCleanRemovesOnlyListedSubpaths(t *testing.T) {
	orig := core.DefaultMinCleanBytes
	core.DefaultMinCleanBytes = 0
	defer func() { core.DefaultMinCleanBytes = orig }()

	root := t.TempDir()
	mustMkdir(t, filepath.Join(root, "Cache"))
	mustMkdir(t, filepath.Join(root, "User")) // not in subpaths — must survive
	writeFixtureFile(t, filepath.Join(root, "Cache"), "a.bin", 1000)
	writeFixtureFile(t, filepath.Join(root, "User"), "settings.json", 50)

	c := NewMultiDirCache("vscode", root, []string{"Cache"}, true)
	if _, err := c.Clean(false); err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(filepath.Join(root, "Cache", "a.bin")); !os.IsNotExist(err) {
		t.Fatalf("expected Cache/a.bin to be removed, stat error: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, "User", "settings.json")); err != nil {
		t.Fatalf("expected User/settings.json (not in subpaths) to survive: %v", err)
	}
}

func mustMkdir(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatal(err)
	}
}
