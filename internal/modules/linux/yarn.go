// Package linux holds Linux cache-cleaning module implementations.
package linux

import (
	"io/fs"
	"os"
	"path/filepath"

	"github.com/vinidev/devopt/internal/core"
)

// Yarn is the reference module implementation — real Detect/Calculate/Clean.
// New modules should copy this pattern; see .claude/contexts/architecture.md
// and .claude/commands/add-module.md.
type Yarn struct {
	Path string // e.g. ~/.cache/yarn, injected from config.Rules
}

func NewYarn(path string) *Yarn {
	return &Yarn{Path: path}
}

func (y *Yarn) Name() string { return "yarn" }

func (y *Yarn) Detect() (bool, error) {
	info, err := os.Stat(y.Path)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return info.IsDir(), nil
}

func (y *Yarn) Calculate() (core.Finding, error) {
	var size int64
	var items int
	err := filepath.WalkDir(y.Path, func(_ string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
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
	if err != nil {
		return core.Finding{}, err
	}
	return core.Finding{Module: y.Name(), Path: y.Path, SizeBytes: size, ItemCount: items}, nil
}

func (y *Yarn) Safe() bool { return true }

func (y *Yarn) Clean(dryRun bool) (core.CleanResult, error) {
	if err := core.Guard(y.Path); err != nil {
		return core.CleanResult{}, err
	}

	finding, err := y.Calculate()
	if err != nil {
		return core.CleanResult{}, err
	}

	if core.ShouldSkipSmall(finding.SizeBytes) {
		return core.CleanResult{
			Module: y.Name(), Path: y.Path, DryRun: dryRun,
			Skipped: true, SkipReason: "abaixo do limiar de 200 MB",
		}, nil
	}

	if dryRun {
		return core.CleanResult{
			Module: y.Name(), Path: y.Path, FreedBytes: finding.SizeBytes, DryRun: true,
		}, nil
	}

	entries, err := os.ReadDir(y.Path)
	if err != nil {
		return core.CleanResult{}, err
	}
	for _, entry := range entries {
		if err := os.RemoveAll(filepath.Join(y.Path, entry.Name())); err != nil {
			return core.CleanResult{}, err
		}
	}

	return core.CleanResult{
		Module: y.Name(), Path: y.Path, FreedBytes: finding.SizeBytes, DryRun: false,
	}, nil
}
