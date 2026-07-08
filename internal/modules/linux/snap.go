package linux

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/Vbanety/Developer-Workspace-Optimizer/internal/core"
)

// Snap is a bespoke module — like Docker, it shells out to a CLI rather
// than walking a directory.
type Snap struct{}

func NewSnap() *Snap { return &Snap{} }

func (s *Snap) Name() string { return "snap" }

func (s *Snap) Safe() bool { return false }

func (s *Snap) Detect() (bool, error) {
	if _, err := exec.LookPath("snap"); err != nil {
		return false, nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := exec.CommandContext(ctx, "snap", "version").Run(); err != nil {
		return false, nil
	}
	return true, nil
}

// snapRevision is a disabled (superseded) revision of an installed snap —
// safe to remove without affecting the currently active version.
type snapRevision struct {
	name      string
	revision  string
	path      string
	sizeBytes int64
}

// snapsDir is a var, not a const, so tests can point it at a temp dir
// instead of touching the real /var/lib/snapd/snaps.
var snapsDir = "/var/lib/snapd/snaps"

// listDisabledRevisions parses `snap list --all` and stats the on-disk
// squashfs file for every revision marked "disabled" in the Notes column.
// Shared by Calculate and Clean so the table is only parsed once per call.
func listDisabledRevisions() ([]snapRevision, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	out, err := exec.CommandContext(ctx, "snap", "list", "--all").Output()
	if err != nil {
		return nil, fmt.Errorf("snap list --all: %w", err)
	}
	return parseSnapListDisabled(string(out))
}

// parseSnapListDisabled parses `snap list --all` output. Always 6
// whitespace-separated fields (Name, Version, Rev, Tracking, Publisher,
// Notes) — verified against a real machine. Notes may combine multiple
// flags like "classic,disabled", so check Contains, not equality. Rows
// that don't match the expected shape or whose file is gone are skipped
// rather than failing the whole scan.
func parseSnapListDisabled(output string) ([]snapRevision, error) {
	lines := strings.Split(strings.TrimRight(output, "\n"), "\n")
	if len(lines) <= 1 {
		return nil, nil // just the header, or empty
	}

	var revisions []snapRevision
	for _, line := range lines[1:] { // skip header
		fields := strings.Fields(line)
		if len(fields) != 6 {
			continue
		}
		name, rev, notes := fields[0], fields[2], fields[5]
		if !strings.Contains(notes, "disabled") {
			continue
		}
		path := filepath.Join(snapsDir, fmt.Sprintf("%s_%s.snap", name, rev))
		info, err := os.Stat(path)
		if err != nil {
			continue
		}
		revisions = append(revisions, snapRevision{name: name, revision: rev, path: path, sizeBytes: info.Size()})
	}
	return revisions, nil
}

func (s *Snap) Calculate() (core.Finding, error) {
	revisions, err := listDisabledRevisions()
	if err != nil {
		return core.Finding{}, err
	}
	var total int64
	for _, r := range revisions {
		total += r.sizeBytes
	}
	return core.Finding{Module: s.Name(), Path: snapsDir, SizeBytes: total, ItemCount: len(revisions)}, nil
}

// Clean removes each disabled revision independently — unlike DirCache,
// this isn't an all-or-nothing directory removal, so one failure (e.g. no
// root privileges) shouldn't block the others. FreedBytes only counts
// revisions actually removed; a combined error lists what failed and why.
func (s *Snap) Clean(dryRun bool) (core.CleanResult, error) {
	revisions, err := listDisabledRevisions()
	if err != nil {
		return core.CleanResult{}, err
	}

	var total int64
	for _, r := range revisions {
		total += r.sizeBytes
	}

	if core.ShouldSkipSmall(total) {
		return core.CleanResult{
			Module: s.Name(), Path: snapsDir, DryRun: dryRun,
			Skipped: true, SkipReason: "abaixo do limiar de 200 MB",
		}, nil
	}

	if dryRun {
		return core.CleanResult{Module: s.Name(), Path: snapsDir, FreedBytes: total, DryRun: true}, nil
	}

	var freed int64
	var failures []string
	for _, r := range revisions {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		err := exec.CommandContext(ctx, "snap", "remove", "--revision="+r.revision, r.name).Run()
		cancel()
		if err != nil {
			failures = append(failures, fmt.Sprintf("%s rev %s: %v", r.name, r.revision, err))
			continue
		}
		freed += r.sizeBytes
	}

	result := core.CleanResult{Module: s.Name(), Path: snapsDir, FreedBytes: freed, DryRun: false}
	if len(failures) > 0 {
		return result, fmt.Errorf(
			"%d/%d revisões removidas, falhas: %s",
			len(revisions)-len(failures), len(revisions), strings.Join(failures, "; "),
		)
	}
	return result, nil
}
