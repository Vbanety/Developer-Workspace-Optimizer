package tui

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/vinidev/devopt/internal/core"
)

// Run launches the interactive menu against reg until the user quits.
func Run(reg *core.Registry) error {
	_, err := tea.NewProgram(NewModel(reg)).Run()
	return err
}
