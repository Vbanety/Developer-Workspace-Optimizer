package core

import (
	"os"
	"path/filepath"
	"strings"
)

// DefaultMinCleanBytes is the skip threshold: caches smaller than this are
// left alone automatically (not worth the risk/prompt). A var, not a const,
// so tests can override it instead of writing multi-hundred-MB fixtures.
var DefaultMinCleanBytes int64 = 200 * 1024 * 1024 // 200 MB

// projectMarkers are files whose presence at a directory's root mean "this
// is a project, not disposable cache" — never delete it wholesale.
var projectMarkers = []string{".git", "package.json", "go.mod", "composer.json", "Cargo.toml"}

// ErrUnsafePath is returned when a candidate path fails the guard.
type ErrUnsafePath struct {
	Path   string
	Reason string
}

func (e *ErrUnsafePath) Error() string {
	return "unsafe path: " + e.Path + " (" + e.Reason + ")"
}

// neverDeletePaths are absolute path prefixes Clean must never touch,
// regardless of what a module computes. A hard floor beneath each module's
// own path selection.
func neverDeletePaths() []string {
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return nil
	}
	return []string{
		filepath.Join(home, "Documents"),
		filepath.Join(home, "Desktop"),
		filepath.Join(home, ".config"),
		filepath.Join(home, ".ssh"),
	}
}

// Guard validates a candidate path before any Clean touches disk. Every
// module's Clean must call this first and abort on error.
func Guard(path string) error {
	abs, err := filepath.Abs(path)
	if err != nil {
		return err
	}
	for _, forbidden := range neverDeletePaths() {
		if abs == forbidden || strings.HasPrefix(abs, forbidden+string(filepath.Separator)) {
			return &ErrUnsafePath{Path: abs, Reason: "in never-delete list"}
		}
	}
	if looksLikeProject(abs) {
		return &ErrUnsafePath{Path: abs, Reason: "looks like a project directory"}
	}
	return nil
}

func looksLikeProject(path string) bool {
	for _, marker := range projectMarkers {
		if _, err := os.Stat(filepath.Join(path, marker)); err == nil {
			return true
		}
	}
	return false
}

// ShouldSkipSmall reports whether a finding is below the clean-worthiness
// threshold and should be skipped automatically.
func ShouldSkipSmall(sizeBytes int64) bool {
	return sizeBytes < DefaultMinCleanBytes
}
