package tui

import (
	"path/filepath"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/vinidev/devopt/internal/audit"
	"github.com/vinidev/devopt/internal/core"
)

// fakeModule is a deterministic in-memory core.Module for testing the
// model's state machine without touching real disk/docker/snap.
type fakeModule struct {
	name      string
	safe      bool
	sizeBytes int64
	cleanErr  error
}

func (f *fakeModule) Name() string { return f.name }
func (f *fakeModule) Safe() bool   { return f.safe }
func (f *fakeModule) Detect() (bool, error) {
	return true, nil
}
func (f *fakeModule) Calculate() (core.Finding, error) {
	return core.Finding{Module: f.name, Path: "/fake/" + f.name, SizeBytes: f.sizeBytes}, nil
}
func (f *fakeModule) Clean(dryRun bool) (core.CleanResult, error) {
	if f.cleanErr != nil {
		return core.CleanResult{}, f.cleanErr
	}
	return core.CleanResult{Module: f.name, FreedBytes: f.sizeBytes, DryRun: dryRun}, nil
}

func testRegistry() *core.Registry {
	reg := core.NewRegistry()
	reg.Register(&fakeModule{name: "yarn", safe: true, sizeBytes: 500 * 1024 * 1024})
	reg.Register(&fakeModule{name: "docker", safe: false, sizeBytes: 300 * 1024 * 1024})
	return reg
}

// runCmd synchronously executes a tea.Cmd the way the bubbletea runtime
// would, without needing a real Program/TTY — Cmd is just a func() tea.Msg.
func runCmd(t *testing.T, cmd tea.Cmd) tea.Msg {
	t.Helper()
	if cmd == nil {
		t.Fatal("expected a non-nil cmd")
	}
	return cmd()
}

func key(s string) tea.KeyMsg {
	switch s {
	case "up":
		return tea.KeyMsg{Type: tea.KeyUp}
	case "down":
		return tea.KeyMsg{Type: tea.KeyDown}
	case "enter":
		return tea.KeyMsg{Type: tea.KeyEnter}
	case " ":
		return tea.KeyMsg{Type: tea.KeySpace}
	case "esc":
		return tea.KeyMsg{Type: tea.KeyEsc}
	default:
		return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(s)}
	}
}

// afterScan drives a fresh model through Init + scan completion, landing on
// screenMenu, the way the real program does on startup.
func afterScan(t *testing.T) Model {
	t.Helper()
	m := NewModel(testRegistry())
	msg := runCmd(t, m.Init())
	// Init batches scanCmd + tickCmd; tea.Batch's returned Cmd runs both and
	// emits a tea.BatchMsg — resolve it down to the scanDoneMsg we want.
	switch v := msg.(type) {
	case tea.BatchMsg:
		for _, c := range v {
			if sd, ok := runCmd(t, c).(scanDoneMsg); ok {
				msg = sd
			}
		}
	}
	next, _ := m.Update(msg)
	got := next.(Model)
	if got.screen != screenMenu {
		t.Fatalf("expected screenMenu after scan, got %v", got.screen)
	}
	return got
}

func TestModelScanToMenu(t *testing.T) {
	m := afterScan(t)
	if len(m.actionable) != 2 {
		t.Fatalf("expected 2 actionable rows, got %d: %+v", len(m.actionable), m.actionable)
	}
}

func TestModelSafeCleanTargetsOnlySafeModules(t *testing.T) {
	m := afterScan(t)
	// cursor starts at 0 = "Limpeza segura"
	next, _ := m.Update(key("enter"))
	m = next.(Model)

	if m.screen != screenConfirm {
		t.Fatalf("expected screenConfirm, got %v", m.screen)
	}
	if len(m.confirmTarget) != 1 || m.confirmTarget[0] != "yarn" {
		t.Fatalf("expected only yarn (safe) targeted, got %v", m.confirmTarget)
	}
}

func TestModelDeepCleanTargetsAllModules(t *testing.T) {
	m := afterScan(t)
	next, _ := m.Update(key("down")) // cursor -> "Limpeza profunda"
	m = next.(Model)
	next, _ = m.Update(key("enter"))
	m = next.(Model)

	if m.screen != screenConfirm {
		t.Fatalf("expected screenConfirm, got %v", m.screen)
	}
	if len(m.confirmTarget) != 2 {
		t.Fatalf("expected both modules targeted, got %v", m.confirmTarget)
	}
}

func TestModelPickModulesToggleAndConfirm(t *testing.T) {
	m := afterScan(t)
	next, _ := m.Update(key("down"))
	m = next.(Model)
	next, _ = m.Update(key("down")) // cursor -> "Escolher módulos"
	m = next.(Model)
	next, _ = m.Update(key("enter"))
	m = next.(Model)

	if m.screen != screenModuleSelect {
		t.Fatalf("expected screenModuleSelect, got %v", m.screen)
	}

	next, _ = m.Update(key(" ")) // toggle first item
	m = next.(Model)
	next, _ = m.Update(key("enter"))
	m = next.(Model)

	if m.screen != screenConfirm {
		t.Fatalf("expected screenConfirm, got %v", m.screen)
	}
	if len(m.confirmTarget) != 1 {
		t.Fatalf("expected exactly 1 module targeted, got %v", m.confirmTarget)
	}
}

func TestModelConfirmNoGoesBackToMenu(t *testing.T) {
	m := afterScan(t)
	next, _ := m.Update(key("enter")) // -> confirm (safe clean)
	m = next.(Model)
	next, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("n")})
	m = next.(Model)
	if m.screen != screenMenu {
		t.Fatalf("expected back to screenMenu, got %v", m.screen)
	}
}

func TestModelConfirmYesRunsCleanAndReachesDone(t *testing.T) {
	defer audit.OverrideLogPathForTest(filepath.Join(t.TempDir(), "history.jsonl"))()

	m := afterScan(t)
	next, _ := m.Update(key("enter")) // -> confirm (safe clean: yarn only)
	m = next.(Model)

	next, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("y")})
	m = next.(Model)
	if m.screen != screenRunning {
		t.Fatalf("expected screenRunning, got %v", m.screen)
	}

	msg := runCmd(t, cmd)
	if batch, ok := msg.(tea.BatchMsg); ok {
		for _, c := range batch {
			if done, ok := runCmd(t, c).(cleanAllDoneMsg); ok {
				msg = done
			}
		}
	}

	next, _ = m.Update(msg)
	m = next.(Model)
	if m.screen != screenDone {
		t.Fatalf("expected screenDone, got %v", m.screen)
	}
	if len(m.outcomes) != 1 || m.outcomes[0].module != "yarn" {
		t.Fatalf("expected yarn outcome, got %+v", m.outcomes)
	}
	if m.outcomes[0].result.FreedBytes != 500*1024*1024 {
		t.Fatalf("expected 500MB freed, got %d", m.outcomes[0].result.FreedBytes)
	}
}
