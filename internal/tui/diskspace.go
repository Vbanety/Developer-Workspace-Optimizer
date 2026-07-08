package tui

import "golang.org/x/sys/unix"

// freeBytes returns available disk space, in bytes, for the filesystem
// containing path. Best-effort — returns 0 on error so the "done" screen
// can just omit the before/after comparison instead of crashing over it.
func freeBytes(path string) int64 {
	var stat unix.Statfs_t
	if err := unix.Statfs(path, &stat); err != nil {
		return 0
	}
	return int64(stat.Bavail) * int64(stat.Bsize)
}
