package core

import "errors"

// ErrNotImplemented is returned by stub modules whose Calculate/Clean
// logic hasn't been written yet. Report/clean flows must handle it by
// listing the module as "não implementado" instead of failing.
var ErrNotImplemented = errors.New("module not implemented yet")

// Module is the contract every cache-cleaning module implements.
// Flow is always Detect -> Calculate -> (report) -> safety guard -> Clean.
type Module interface {
	// Name is the module identifier used in CLI flags/reports (e.g. "yarn").
	Name() string

	// Detect reports whether this tool's cache/artifacts exist on this system.
	Detect() (bool, error)

	// Calculate scans the target path and reports its size and item count.
	// Only called after Detect returns true.
	Calculate() (Finding, error)

	// Safe reports whether this module is eligible for "safe clean" without
	// extra confirmation. Only true for pure, provably orphaned caches.
	Safe() bool

	// Clean removes the target path's contents. When dryRun is true, it must
	// not touch disk — it reports what would happen instead.
	Clean(dryRun bool) (CleanResult, error)
}
