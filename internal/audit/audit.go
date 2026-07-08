// Package audit records every real Clean() invocation to a JSONL log so
// users can see what devopt actually did over time — module, path, bytes
// freed, and whether it was a dry-run, skip, or error.
package audit

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"github.com/Vbanety/Developer-Workspace-Optimizer/internal/core"
)

// Entry is one Clean() invocation, successful or not.
type Entry struct {
	Time       time.Time `json:"time"`
	Module     string    `json:"module"`
	Path       string    `json:"path"`
	FreedBytes int64     `json:"freedBytes"`
	DryRun     bool      `json:"dryRun"`
	Skipped    bool      `json:"skipped,omitempty"`
	SkipReason string    `json:"skipReason,omitempty"`
	Error      string    `json:"error,omitempty"`
}

// FromCleanResult builds an Entry from a module's Clean() outcome — shared
// by cmd/devopt and internal/tui so the field mapping isn't duplicated at
// each call site.
func FromCleanResult(module string, res core.CleanResult, cleanErr error) Entry {
	e := Entry{
		Time:       time.Now(),
		Module:     module,
		Path:       res.Path,
		FreedBytes: res.FreedBytes,
		DryRun:     res.DryRun,
		Skipped:    res.Skipped,
		SkipReason: res.SkipReason,
	}
	if cleanErr != nil {
		e.Error = cleanErr.Error()
	}
	return e
}

// logPath is a var, not a const, so tests can point it at a temp file
// instead of the real ~/.local/share/devopt/history.jsonl.
var logPath = defaultLogPath()

// OverrideLogPathForTest points the audit log at a custom path for the
// duration of a test — used by other packages' tests (e.g. internal/tui)
// so exercising a real Clean() during `go test` doesn't write into the
// user's actual history.jsonl. Not for production use.
func OverrideLogPathForTest(path string) (restore func()) {
	orig := logPath
	logPath = path
	return func() { logPath = orig }
}

func defaultLogPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".local", "share", "devopt", "history.jsonl")
}

// Record appends one entry to the audit log. Logging is best-effort — a
// failure here (e.g. unwritable disk) shouldn't be treated as fatal by
// callers, since the clean operation itself already happened.
func Record(e Entry) error {
	if logPath == "" {
		return nil // no home dir resolvable, nothing sane to do
	}
	if err := os.MkdirAll(filepath.Dir(logPath), 0o755); err != nil {
		return err
	}
	f, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()

	data, err := json.Marshal(e)
	if err != nil {
		return err
	}
	_, err = f.Write(append(data, '\n'))
	return err
}

// Read returns up to n most recent entries, oldest first (n <= 0 means
// all). A missing log file isn't an error — it just means nothing has
// been recorded yet.
func Read(n int) ([]Entry, error) {
	if logPath == "" {
		return nil, nil
	}
	f, err := os.Open(logPath)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var entries []Entry
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}
		var e Entry
		if err := json.Unmarshal(line, &e); err != nil {
			continue // skip malformed lines instead of failing the whole read
		}
		entries = append(entries, e)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	if n > 0 && len(entries) > n {
		entries = entries[len(entries)-n:]
	}
	return entries, nil
}
