package linux

import (
	"os"

	"github.com/vinidev/devopt/internal/core"
)

// Stub is a placeholder module: Detect is real (checks the path exists),
// Calculate/Clean report core.ErrNotImplemented instead of guessing at
// unwritten logic. When a module graduates from stub to full support,
// replace its Stub registration with a real type (see yarn.go) — don't
// duplicate this type per module, just register more instances of it.
type Stub struct {
	name string
	path string
}

func NewStub(name, path string) *Stub {
	return &Stub{name: name, path: path}
}

func (s *Stub) Name() string { return s.name }

func (s *Stub) Detect() (bool, error) {
	_, err := os.Stat(s.path)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func (s *Stub) Safe() bool { return false }

func (s *Stub) Calculate() (core.Finding, error) {
	return core.Finding{}, core.ErrNotImplemented
}

func (s *Stub) Clean(_ bool) (core.CleanResult, error) {
	return core.CleanResult{}, core.ErrNotImplemented
}
