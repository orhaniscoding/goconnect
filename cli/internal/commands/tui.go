package commands

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/orhaniscoding/goconnect/cli/internal/tui"
)

// RunTUIWithState launches the TUI with the given initial state
func RunTUIWithState(initialState tui.SessionState) {
	model := tui.NewModelWithState(initialState)
	p := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
