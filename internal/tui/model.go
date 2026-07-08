// Package tui is devopt's interactive menu: scan → action menu (safe
// clean / deep clean / pick modules / write report / quit) → confirm →
// run → final report. cmd/devopt's report/clean subcommands remain the
// scriptable, non-interactive path — this is only wired up when devopt is
// invoked with no subcommand.
package tui

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/vinidev/devopt/internal/core"
	"github.com/vinidev/devopt/internal/report"
)

type screen int

const (
	screenScanning screen = iota
	screenMenu
	screenModuleSelect
	screenConfirm
	screenRunning
	screenDone
)

var actionMenuLabels = []string{
	"Limpeza segura",
	"Limpeza profunda",
	"Escolher módulos",
	"Gerar relatório",
	"Sair",
}

const (
	actionSafeClean = iota
	actionDeepClean
	actionPickModules
	actionWriteReport
	actionQuit
)

// cleanOutcome is the result of running Clean() on one module during the
// "running" screen.
type cleanOutcome struct {
	module string
	result core.CleanResult
	err    error
}

type Model struct {
	reg  *core.Registry
	rows []report.Row
	// actionable is rows filtered down to modules Clean() can actually act
	// on (excludes NotImpl and below-threshold/error rows).
	actionable []report.Row

	screen screen
	status string // transient one-line status (e.g. "relatório salvo em ...")

	menu   selectList
	modSel selectList

	confirmTarget []string // module names about to be cleaned
	outcomes      []cleanOutcome

	freeBefore int64
	freeAfter  int64

	spinnerIdx int
}

func NewModel(reg *core.Registry) Model {
	return Model{reg: reg, screen: screenScanning}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(scanCmd(m.reg), tickCmd())
}

// --- messages ---

type scanDoneMsg struct{ rows []report.Row }
type tickMsg time.Time
type cleanAllDoneMsg struct{ results []cleanOutcome }

func scanCmd(reg *core.Registry) tea.Cmd {
	return func() tea.Msg {
		return scanDoneMsg{rows: report.Scan(reg)}
	}
}

func tickCmd() tea.Cmd {
	return tea.Tick(120*time.Millisecond, func(t time.Time) tea.Msg { return tickMsg(t) })
}

// cleanAllCmd runs Clean() on every target module sequentially. Kept as one
// batch Cmd rather than a per-step channel pipeline — Clean() calls here are
// local FS ops or a couple of CLI invocations, not slow enough to justify
// the extra complexity of incremental progress messages.
func cleanAllCmd(targets []core.Module, dryRun bool) tea.Cmd {
	return func() tea.Msg {
		outcomes := make([]cleanOutcome, 0, len(targets))
		for _, mod := range targets {
			res, err := mod.Clean(dryRun)
			outcomes = append(outcomes, cleanOutcome{module: mod.Name(), result: res, err: err})
		}
		return cleanAllDoneMsg{results: outcomes}
	}
}

// --- update ---

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKey(msg)
	case scanDoneMsg:
		m.rows = msg.rows
		m.actionable = actionableRows(m.rows)
		m.menu = newSelectList(actionMenuLabels, false)
		m.screen = screenMenu
		m.status = ""
		return m, nil
	case tickMsg:
		m.spinnerIdx++
		if m.screen == screenScanning || m.screen == screenRunning {
			return m, tickCmd()
		}
		return m, nil
	case cleanAllDoneMsg:
		m.outcomes = msg.results
		m.freeAfter = freeBytes(homeDir())
		m.screen = screenDone
		return m, nil
	}
	return m, nil
}

func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if msg.String() == "ctrl+c" {
		return m, tea.Quit
	}

	switch m.screen {
	case screenMenu:
		switch msg.String() {
		case "up", "k":
			m.menu.up()
		case "down", "j":
			m.menu.down()
		case "enter":
			return m.chooseMenuAction()
		case "q":
			return m, tea.Quit
		}
	case screenModuleSelect:
		switch msg.String() {
		case "up", "k":
			m.modSel.up()
		case "down", "j":
			m.modSel.down()
		case " ":
			m.modSel.toggle()
		case "enter":
			return m.confirmModuleSelection()
		case "esc":
			m.screen = screenMenu
		}
	case screenConfirm:
		switch msg.String() {
		case "y", "enter":
			return m.startClean()
		case "n", "esc":
			m.screen = screenMenu
		}
	case screenDone:
		switch msg.String() {
		case "enter":
			m.screen = screenScanning
			return m, tea.Batch(scanCmd(m.reg), tickCmd())
		case "q":
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m Model) chooseMenuAction() (tea.Model, tea.Cmd) {
	switch m.menu.cursor {
	case actionSafeClean:
		m.confirmTarget = moduleNames(m.actionable, func(r report.Row) bool {
			mod := m.reg.Get(r.Module)
			return mod != nil && mod.Safe()
		})
		m.screen = screenConfirm
	case actionDeepClean:
		m.confirmTarget = moduleNames(m.actionable, func(report.Row) bool { return true })
		m.screen = screenConfirm
	case actionPickModules:
		labels := make([]string, len(m.actionable))
		for i, r := range m.actionable {
			labels[i] = fmt.Sprintf("%s (%s)", r.Module, report.HumanSize(r.SizeBytes))
		}
		m.modSel = newSelectList(labels, true)
		m.screen = screenModuleSelect
	case actionWriteReport:
		path, err := writeReportFile(m.rows)
		if err != nil {
			m.status = "erro ao gerar relatório: " + err.Error()
		} else {
			m.status = "relatório salvo em " + path
		}
	case actionQuit:
		return m, tea.Quit
	}
	return m, nil
}

func (m Model) confirmModuleSelection() (tea.Model, tea.Cmd) {
	var target []string
	for _, i := range m.modSel.selectedIndices() {
		target = append(target, m.actionable[i].Module)
	}
	if len(target) == 0 {
		m.screen = screenMenu
		return m, nil
	}
	m.confirmTarget = target
	m.screen = screenConfirm
	return m, nil
}

func (m Model) startClean() (tea.Model, tea.Cmd) {
	var targets []core.Module
	for _, name := range m.confirmTarget {
		if mod := m.reg.Get(name); mod != nil {
			targets = append(targets, mod)
		}
	}
	m.freeBefore = freeBytes(homeDir())
	m.screen = screenRunning
	return m, tea.Batch(cleanAllCmd(targets, false), tickCmd())
}

// --- helpers ---

// actionableRows filters out modules Clean() shouldn't be invoked for:
// not-yet-implemented or already flagged with a skip reason (error, below
// the size threshold).
func actionableRows(rows []report.Row) []report.Row {
	var out []report.Row
	for _, r := range rows {
		if r.NotImpl || r.SkipReason != "" {
			continue
		}
		out = append(out, r)
	}
	return out
}

func moduleNames(rows []report.Row, include func(report.Row) bool) []string {
	var out []string
	for _, r := range rows {
		if include(r) {
			out = append(out, r.Module)
		}
	}
	return out
}

func writeReportFile(rows []report.Row) (string, error) {
	name := fmt.Sprintf("devopt-report-%s.json", time.Now().Format("20060102-150405"))
	data, err := json.MarshalIndent(rows, "", "  ")
	if err != nil {
		return "", err
	}
	if err := os.WriteFile(name, data, 0o644); err != nil {
		return "", err
	}
	return name, nil
}

func homeDir() string {
	h, err := os.UserHomeDir()
	if err != nil {
		return "/"
	}
	return h
}

var spinnerFrames = []string{"|", "/", "-", "\\"}

func spinnerFrame(i int) string {
	return spinnerFrames[i%len(spinnerFrames)]
}

// --- view ---

func (m Model) View() string {
	switch m.screen {
	case screenScanning:
		return spinnerFrame(m.spinnerIdx) + " escaneando caches...\n"

	case screenMenu:
		var b strings.Builder
		b.WriteString(report.Render(m.rows))
		b.WriteString("\n")
		b.WriteString(m.menu.View())
		if m.status != "" {
			b.WriteString("\n" + m.status + "\n")
		}
		b.WriteString("\n↑/↓ navega, enter escolhe, q sai\n")
		return b.String()

	case screenModuleSelect:
		var b strings.Builder
		b.WriteString("Escolher módulos (espaço alterna, enter confirma, esc volta):\n\n")
		b.WriteString(m.modSel.View())
		return b.String()

	case screenConfirm:
		var b strings.Builder
		b.WriteString("Confirma limpeza dos módulos:\n\n")
		var total int64
		for _, name := range m.confirmTarget {
			for _, r := range m.actionable {
				if r.Module == name {
					total += r.SizeBytes
					b.WriteString(fmt.Sprintf("  %s (%s)\n", name, report.HumanSize(r.SizeBytes)))
				}
			}
		}
		b.WriteString(fmt.Sprintf("\nTotal: %s\n\n(y) confirmar   (n) cancelar\n", report.HumanSize(total)))
		return b.String()

	case screenRunning:
		return spinnerFrame(m.spinnerIdx) + fmt.Sprintf(" limpando %d módulo(s)...\n", len(m.confirmTarget))

	case screenDone:
		var b strings.Builder
		b.WriteString("Limpeza concluída\n\n")
		var total int64
		for _, o := range m.outcomes {
			switch {
			case o.err != nil:
				b.WriteString(fmt.Sprintf("  ✗ %s: erro (%v)\n", o.module, o.err))
			case o.result.Skipped:
				b.WriteString(fmt.Sprintf("  … %s: pulado (%s)\n", o.module, o.result.SkipReason))
			default:
				b.WriteString(fmt.Sprintf("  ✔ %s: %s liberado\n", o.module, report.HumanSize(o.result.FreedBytes)))
				total += o.result.FreedBytes
			}
		}
		b.WriteString(fmt.Sprintf("\nTotal liberado: %s\n", report.HumanSize(total)))
		if m.freeBefore > 0 && m.freeAfter > 0 {
			b.WriteString(fmt.Sprintf(
				"\nEspaço livre antes: %s\nEspaço livre agora: %s\n",
				report.HumanSize(m.freeBefore), report.HumanSize(m.freeAfter),
			))
		}
		b.WriteString("\n(enter) voltar ao menu   (q) sair\n")
		return b.String()
	}
	return ""
}
