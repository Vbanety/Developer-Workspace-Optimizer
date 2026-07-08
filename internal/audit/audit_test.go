package audit

import (
	"errors"
	"path/filepath"
	"testing"

	"github.com/vinidev/devopt/internal/core"
)

func withTempLog(t *testing.T) {
	t.Helper()
	orig := logPath
	logPath = filepath.Join(t.TempDir(), "history.jsonl")
	t.Cleanup(func() { logPath = orig })
}

func TestRecordAndReadRoundTrip(t *testing.T) {
	withTempLog(t)

	if err := Record(Entry{Module: "yarn", Path: "/fake/yarn", FreedBytes: 100}); err != nil {
		t.Fatal(err)
	}
	if err := Record(Entry{Module: "docker", Path: "/fake/docker", FreedBytes: 200}); err != nil {
		t.Fatal(err)
	}

	entries, err := Read(0)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d: %+v", len(entries), entries)
	}
	if entries[0].Module != "yarn" || entries[1].Module != "docker" {
		t.Fatalf("unexpected order/content: %+v", entries)
	}
}

func TestReadMissingFileReturnsEmpty(t *testing.T) {
	withTempLog(t)

	entries, err := Read(0)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 0 {
		t.Fatalf("expected no entries for a missing log, got %d", len(entries))
	}
}

func TestReadLimitReturnsMostRecent(t *testing.T) {
	withTempLog(t)

	for _, name := range []string{"a", "b", "c", "d"} {
		if err := Record(Entry{Module: name}); err != nil {
			t.Fatal(err)
		}
	}

	entries, err := Read(2)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 2 || entries[0].Module != "c" || entries[1].Module != "d" {
		t.Fatalf("expected [c d], got %+v", entries)
	}
}

func TestFromCleanResult(t *testing.T) {
	res := core.CleanResult{Path: "/fake/yarn", FreedBytes: 500, DryRun: true}
	e := FromCleanResult("yarn", res, nil)
	if e.Module != "yarn" || e.Path != "/fake/yarn" || e.FreedBytes != 500 || !e.DryRun || e.Error != "" {
		t.Fatalf("unexpected entry: %+v", e)
	}

	e2 := FromCleanResult("apt", core.CleanResult{}, errors.New("permission denied"))
	if e2.Error != "permission denied" {
		t.Fatalf("expected error captured, got %+v", e2)
	}
}
