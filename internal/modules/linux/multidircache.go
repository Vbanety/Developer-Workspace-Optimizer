package linux

import (
	"os"
	"path/filepath"

	"github.com/vinidev/devopt/internal/core"
)

// ElectronCacheSubdirs are the subpaths shared by every Chromium/Electron
// app config dir (Cursor, VS Code) that are pure regenerable cache — safe
// to clear. Curated by inspecting real ~/.config/Cursor and ~/.config/Code
// with `du -sh`: excludes User/ (real settings/workspaceStorage), snapshots
// (undo history), Partitions/WebStorage/*Storage/Cookies* (login/session
// state), and other non-cache data. See .claude/contexts/architecture.md.
var ElectronCacheSubdirs = []string{
	"Cache",
	"Code Cache",
	"CachedData",
	"GPUCache",
	"DawnCache",
	"DawnGraphiteCache",
	"DawnWebGPUCache",
	"CachedExtensionVSIXs",
	"CachedExtensions",
	"CachedProfilesData",
	"Crashpad",
	"logs",
	"Service Worker",
	"Shared Dictionary",
}

// MultiDirCache is the generic module for an app whose cache is spread
// across several known-safe subdirectories under a root that also holds
// real user data (so the root itself must never be wiped wholesale — only
// the listed subpaths). Covers cursor/vscode; don't write a new type per
// app, construct another MultiDirCache instead (see cmd/devopt/main.go).
type MultiDirCache struct {
	name     string
	root     string
	subpaths []string
	safe     bool
}

func NewMultiDirCache(name, root string, subpaths []string, safe bool) *MultiDirCache {
	return &MultiDirCache{name: name, root: root, subpaths: subpaths, safe: safe}
}

func (c *MultiDirCache) Name() string { return c.name }

func (c *MultiDirCache) Safe() bool { return c.safe }

func (c *MultiDirCache) Detect() (bool, error) {
	for _, sub := range c.subpaths {
		info, err := os.Stat(filepath.Join(c.root, sub))
		if err == nil && info.IsDir() {
			return true, nil
		}
		if err != nil && !os.IsNotExist(err) {
			return false, err
		}
	}
	return false, nil
}

func (c *MultiDirCache) Calculate() (core.Finding, error) {
	var size int64
	var items int
	for _, sub := range c.subpaths {
		path := filepath.Join(c.root, sub)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			continue
		}
		s, n, err := dirSize(path)
		if err != nil {
			return core.Finding{}, err
		}
		size += s
		items += n
	}
	return core.Finding{Module: c.name, Path: c.root, SizeBytes: size, ItemCount: items}, nil
}

func (c *MultiDirCache) Clean(dryRun bool) (core.CleanResult, error) {
	finding, err := c.Calculate()
	if err != nil {
		return core.CleanResult{}, err
	}

	if core.ShouldSkipSmall(finding.SizeBytes) {
		return core.CleanResult{
			Module: c.name, Path: c.root, DryRun: dryRun,
			Skipped: true, SkipReason: "abaixo do limiar de 200 MB",
		}, nil
	}

	if dryRun {
		return core.CleanResult{
			Module: c.name, Path: c.root, FreedBytes: finding.SizeBytes, DryRun: true,
		}, nil
	}

	for _, sub := range c.subpaths {
		path := filepath.Join(c.root, sub)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			continue
		}
		if err := core.Guard(path); err != nil {
			return core.CleanResult{}, err
		}
		if err := emptyDir(path); err != nil {
			return core.CleanResult{}, err
		}
	}

	return core.CleanResult{
		Module: c.name, Path: c.root, FreedBytes: finding.SizeBytes, DryRun: false,
	}, nil
}
