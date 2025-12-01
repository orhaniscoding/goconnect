package tui

import (
	"github.com/charmbracelet/lipgloss"
)

// Color Palette
var (
	ColorPrimary   = lipgloss.Color("#7D56F4") // Purple
	ColorSecondary = lipgloss.Color("#00ADD8") // Cyan
	ColorSuccess   = lipgloss.Color("#23D18B") // Green
	ColorError     = lipgloss.Color("#F43F5E") // Red
	ColorWarning   = lipgloss.Color("#FFB86C") // Orange
	ColorText      = lipgloss.Color("#F8F8F2") // White
	ColorSubtext   = lipgloss.Color("#6272A4") // Gray
	ColorMuted     = lipgloss.Color("#44475A") // Muted Gray
	ColorBg        = lipgloss.Color("#282A36") // Dark Background
	ColorHighlight = lipgloss.Color("#44475A") // Selection Background
)

// Styles
var (
	// Base styles
	BaseStyle = lipgloss.NewStyle().
			Padding(1, 2).
			Foreground(ColorText)

	// Header styles
	HeaderStyle = lipgloss.NewStyle().
			Foreground(ColorPrimary).
			Bold(true).
			Padding(0, 1).
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(ColorPrimary).
			MarginBottom(1)

	// Status Bar styles
	StatusBarStyle = lipgloss.NewStyle().
			Foreground(ColorText).
			Background(ColorHighlight).
			Padding(0, 1)

	StatusKeyStyle = lipgloss.NewStyle().
			Foreground(ColorPrimary).
			Bold(true)

	StatusValStyle = lipgloss.NewStyle().
			Foreground(ColorSubtext).
			PaddingRight(2)

	// List styles
	ListItemStyle = lipgloss.NewStyle().
			PaddingLeft(2)

	ListItemSelectedStyle = lipgloss.NewStyle().
				PaddingLeft(0).
				Foreground(ColorSecondary).
				Border(lipgloss.NormalBorder(), false, false, false, true).
				BorderForeground(ColorSecondary).
				Bold(true)

	// Form styles
	InputStyle = lipgloss.NewStyle().
			Foreground(ColorSecondary).
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(ColorSubtext).
			Padding(0, 1)

	InputFocusedStyle = lipgloss.NewStyle().
				Foreground(ColorPrimary).
				BorderStyle(lipgloss.RoundedBorder()).
				BorderForeground(ColorPrimary).
				Padding(0, 1)

	// Card styles (for dashboard)
	CardStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorSubtext).
			Padding(1).
			Margin(0, 1)

	CardTitleStyle = lipgloss.NewStyle().
			Foreground(ColorSecondary).
			Bold(true).
			MarginBottom(1)
)

// Helper functions for common UI elements
func RenderHeader(title string) string {
	return HeaderStyle.Render(title)
}

func RenderStatusBar(items ...string) string {
	return StatusBarStyle.Width(80).Render(lipgloss.JoinHorizontal(lipgloss.Top, items...))
}
