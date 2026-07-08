// Package report scans modules and renders results as the terminal table
// described in the original design doc (checkmarks, dotted alignment, total
// recoverable). Shared by cmd/devopt's CLI commands and internal/tui so
// neither has to re-implement the scan loop.
package report

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/vinidev/devopt/internal/core"
)

var (
	checkStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("42"))  // green
	warnStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("214")) // orange
	totalStyle = lipgloss.NewStyle().Bold(true)
)

// Row is a single reportable line: a real finding, a not-yet-implemented
// module, or one skipped below the size threshold.
type Row struct {
	Module     string `json:"module"`
	SizeBytes  int64  `json:"sizeBytes"`
	NotImpl    bool   `json:"notImplemented,omitempty"`
	SkipReason string `json:"skipReason,omitempty"`
}

const nameColumnWidth = 20

// Render prints the scan table + total recoverable size.
func Render(rows []Row) string {
	var b strings.Builder
	b.WriteString("==========================================\n")
	b.WriteString(" devopt — relatório de cache\n")
	b.WriteString("==========================================\n")

	var total int64
	for _, r := range rows {
		label := padDots(capitalize(r.Module), nameColumnWidth)
		switch {
		case r.NotImpl:
			b.WriteString(warnStyle.Render("… "+label+"não implementado") + "\n")
		case r.SkipReason != "":
			b.WriteString(warnStyle.Render("… "+label+r.SkipReason) + "\n")
		default:
			b.WriteString(checkStyle.Render("✔ "+label+HumanSize(r.SizeBytes)) + "\n")
			total += r.SizeBytes
		}
	}

	b.WriteString("\n" + totalStyle.Render("Total possível: "+HumanSize(total)) + "\n")
	return b.String()
}

func padDots(s string, width int) string {
	if len(s) >= width {
		return s + " "
	}
	return s + strings.Repeat(".", width-len(s))
}

func capitalize(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

// HumanSize formats a byte count as a human-readable string (e.g. "11.2 GB").
// Exported so cmd/devopt can reuse it for clean-command output instead of
// duplicating the logic.
func HumanSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// Scan runs Detect+Calculate across every registered module, turning each
// into a Row. Never touches disk. Shared by cmd/devopt (report/clean
// commands) and internal/tui so the scan loop isn't duplicated.
func Scan(reg *core.Registry) []Row {
	var rows []Row
	for _, m := range reg.All() {
		found, err := m.Detect()
		if err != nil || !found {
			continue
		}

		finding, err := m.Calculate()
		switch {
		case err == core.ErrNotImplemented:
			rows = append(rows, Row{Module: m.Name(), NotImpl: true})
		case err != nil:
			rows = append(rows, Row{Module: m.Name(), SkipReason: "erro: " + err.Error()})
		case core.ShouldSkipSmall(finding.SizeBytes):
			rows = append(rows, Row{Module: m.Name(), SkipReason: "abaixo do limiar de 200 MB"})
		default:
			rows = append(rows, Row{Module: m.Name(), SizeBytes: finding.SizeBytes})
		}
	}
	return rows
}
