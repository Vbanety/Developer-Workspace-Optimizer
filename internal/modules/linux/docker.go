package linux

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/Vbanety/Developer-Workspace-Optimizer/internal/core"
)

// Docker is a bespoke module — unlike DirCache it doesn't walk a directory,
// it shells out to the docker CLI. There's no filesystem path for
// core.Guard() to protect here; safety comes from which subcommand/flags
// Clean chooses to run (system prune, never --volumes, never --all) — see
// .claude/contexts/architecture.md.
type Docker struct{}

func NewDocker() *Docker { return &Docker{} }

func (d *Docker) Name() string { return "docker" }

func (d *Docker) Safe() bool { return false }

func (d *Docker) Detect() (bool, error) {
	if _, err := exec.LookPath("docker"); err != nil {
		return false, nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := exec.CommandContext(ctx, "docker", "info").Run(); err != nil {
		return false, nil // CLI present but daemon unreachable — not applicable here
	}
	return true, nil
}

type dockerDFRow struct {
	Type        string `json:"Type"`
	Reclaimable string `json:"Reclaimable"`
}

// reclaimableTypes are the `docker system df` row types counted as
// recoverable. "Local Volumes" is deliberately excluded — never touched,
// per .claude/contexts/safety.md.
var reclaimableTypes = map[string]bool{
	"Images":      true,
	"Containers":  true,
	"Build Cache": true,
}

func (d *Docker) Calculate() (core.Finding, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	out, err := exec.CommandContext(ctx, "docker", "system", "df", "--format", "{{json .}}").Output()
	if err != nil {
		return core.Finding{}, fmt.Errorf("docker system df: %w", err)
	}

	var total int64
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		if line == "" {
			continue
		}
		var row dockerDFRow
		if err := json.Unmarshal([]byte(line), &row); err != nil {
			return core.Finding{}, fmt.Errorf("docker system df: parsing %q: %w", line, err)
		}
		if !reclaimableTypes[row.Type] {
			continue
		}
		n, err := parseDockerSize(row.Reclaimable)
		if err != nil {
			return core.Finding{}, err
		}
		total += n
	}

	return core.Finding{Module: d.Name(), Path: "docker (CLI)", SizeBytes: total}, nil
}

func (d *Docker) Clean(dryRun bool) (core.CleanResult, error) {
	finding, err := d.Calculate()
	if err != nil {
		return core.CleanResult{}, err
	}

	if core.ShouldSkipSmall(finding.SizeBytes) {
		return core.CleanResult{
			Module: d.Name(), Path: finding.Path, DryRun: dryRun,
			Skipped: true, SkipReason: "abaixo do limiar de 200 MB",
		}, nil
	}

	if dryRun {
		return core.CleanResult{
			Module: d.Name(), Path: finding.Path, FreedBytes: finding.SizeBytes, DryRun: true,
		}, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	// Deliberately no --volumes, no --all: never remove volumes or images still in use.
	if err := exec.CommandContext(ctx, "docker", "system", "prune", "-f").Run(); err != nil {
		return core.CleanResult{}, fmt.Errorf("docker system prune: %w", err)
	}

	return core.CleanResult{
		Module: d.Name(), Path: finding.Path, FreedBytes: finding.SizeBytes, DryRun: false,
	}, nil
}

// parseDockerSize parses docker's human-readable size strings (decimal
// units: B/kB/MB/GB/TB — note lowercase "kB"), optionally suffixed with a
// percentage like "1.37GB (70%)".
func parseDockerSize(s string) (int64, error) {
	if idx := strings.Index(s, " ("); idx != -1 {
		s = s[:idx]
	}
	s = strings.TrimSpace(s)

	units := []struct {
		suffix string
		mult   float64
	}{
		{"TB", 1e12},
		{"GB", 1e9},
		{"MB", 1e6},
		{"kB", 1e3},
		{"B", 1},
	}
	for _, u := range units {
		if strings.HasSuffix(s, u.suffix) {
			numStr := strings.TrimSpace(strings.TrimSuffix(s, u.suffix))
			val, err := strconv.ParseFloat(numStr, 64)
			if err != nil {
				return 0, fmt.Errorf("parseDockerSize: %q: %w", s, err)
			}
			return int64(val * u.mult), nil
		}
	}
	return 0, fmt.Errorf("parseDockerSize: unrecognized unit in %q", s)
}
