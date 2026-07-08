// Package config holds default cache paths per module, per OS — the Go
// equivalent of the original design's rules.json. Paths are embedded in the
// binary and can be overridden by ~/.config/devopt/rules.json.
package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
)

// Rules maps a module name (e.g. "yarn") to its default target path.
type Rules struct {
	Paths map[string]string
}

// defaultLinux returns the built-in Linux paths. This is the only OS with
// real entries in v0.1 — windows/darwin are placeholders for future phases
// (see .claude/contexts/roadmap.md).
func defaultLinux(home string) Rules {
	return Rules{Paths: map[string]string{
		"yarn":       filepath.Join(home, ".cache", "yarn"),
		"npm":        filepath.Join(home, ".npm", "_cacache"),
		"pnpm":       filepath.Join(home, ".local", "share", "pnpm", "store"),
		"gradle":     filepath.Join(home, ".gradle", "caches"),
		"composer":   filepath.Join(home, ".cache", "composer"),
		"playwright": filepath.Join(home, ".cache", "ms-playwright"),
		"puppeteer":  filepath.Join(home, ".cache", "puppeteer"),
		"trash":      filepath.Join(home, ".local", "share", "Trash"),
		"apt":        "/var/cache/apt/archives",
		"cursor":     filepath.Join(home, ".config", "Cursor"),
		"vscode":     filepath.Join(home, ".config", "Code"),
	}}
}

// defaultWindows is a placeholder for the v0.4 phase — not implemented yet.
func defaultWindows(_ string) Rules { return Rules{Paths: map[string]string{}} }

// defaultDarwin is a placeholder for the v0.5 phase — not implemented yet.
func defaultDarwin(_ string) Rules { return Rules{Paths: map[string]string{}} }

// Load returns the default rules for the current OS, then applies any
// overrides found in ~/.config/devopt/rules.json on top (per-key merge, not
// a full replace).
func Load() (Rules, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return Rules{}, err
	}

	var rules Rules
	switch runtime.GOOS {
	case "windows":
		rules = defaultWindows(home)
	case "darwin":
		rules = defaultDarwin(home)
	default:
		rules = defaultLinux(home)
	}

	overridePath := filepath.Join(home, ".config", "devopt", "rules.json")
	data, err := os.ReadFile(overridePath)
	if err != nil {
		if os.IsNotExist(err) {
			return rules, nil
		}
		return rules, err
	}

	var overrides map[string]string
	if err := json.Unmarshal(data, &overrides); err != nil {
		return rules, err
	}
	for name, path := range overrides {
		rules.Paths[name] = path
	}
	return rules, nil
}
