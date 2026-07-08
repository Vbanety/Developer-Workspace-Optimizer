// Package linux holds Linux cache-cleaning module implementations.
package linux

import (
	"io/fs"
	"os"
	"path/filepath"

	"github.com/Vbanety/Developer-Workspace-Optimizer/internal/core"
)

// DirCache is the generic module for a cache directory that can be cleared
// wholesale: walk it for total size, then remove its children. Covers every
// v0.1 module (yarn/npm/pnpm/gradle/composer/playwright/puppeteer/trash/apt)
// — they're all the exact same shape, just a different path and Safe value.
// Don't write a new type per module; construct another DirCache instead
// (see cmd/devopt/main.go's buildRegistry). Only reach for a bespoke type
// when a module needs logic this can't express (e.g. talking to the Docker
// daemon instead of walking a directory).
type DirCache struct {
	name string
	path string
	safe bool
}

func NewDirCache(name, path string, safe bool) *DirCache {
	return &DirCache{name: name, path: path, safe: safe}
}

func (c *DirCache) Name() string { return c.name }

func (c *DirCache) Safe() bool { return c.safe }

func (c *DirCache) Detect() (bool, error) {
	info, err := os.Stat(c.path)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return info.IsDir(), nil
}

func (c *DirCache) Calculate() (core.Finding, error) {
	size, items, err := dirSize(c.path)
	if err != nil {
		return core.Finding{}, err
	}
	return core.Finding{Module: c.name, Path: c.path, SizeBytes: size, ItemCount: items}, nil
}

func (c *DirCache) Clean(dryRun bool) (core.CleanResult, error) {
	if err := core.Guard(c.path); err != nil {
		return core.CleanResult{}, err
	}

	finding, err := c.Calculate()
	if err != nil {
		return core.CleanResult{}, err
	}

	if core.ShouldSkipSmall(finding.SizeBytes) {
		return core.CleanResult{
			Module: c.name, Path: c.path, DryRun: dryRun,
			Skipped: true, SkipReason: "abaixo do limiar de 200 MB",
		}, nil
	}

	if dryRun {
		return core.CleanResult{
			Module: c.name, Path: c.path, FreedBytes: finding.SizeBytes, DryRun: true,
		}, nil
	}

	if err := emptyDir(c.path); err != nil {
		return core.CleanResult{}, err
	}

	return core.CleanResult{
		Module: c.name, Path: c.path, FreedBytes: finding.SizeBytes, DryRun: false,
	}, nil
}

// dirSize walks path summing file sizes and counting items. Subtrees it
// can't read (e.g. root-only /var/cache/apt/archives/partial) are skipped
// instead of failing the whole walk. Shared by DirCache and MultiDirCache.
func dirSize(path string) (size int64, items int, err error) {
	err = filepath.WalkDir(path, func(_ string, d fs.DirEntry, err error) error {
		if err != nil {
			if d != nil && d.IsDir() {
				return fs.SkipDir
			}
			return nil
		}
		if !d.IsDir() {
			info, err := d.Info()
			if err != nil {
				return err
			}
			size += info.Size()
			items++
		}
		return nil
	})
	return size, items, err
}

// emptyDir removes every direct child of path, leaving path itself in place.
// Shared by DirCache and MultiDirCache.
func emptyDir(path string) error {
	entries, err := os.ReadDir(path)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if err := os.RemoveAll(filepath.Join(path, entry.Name())); err != nil {
			return err
		}
	}
	return nil
}
