package tui

import (
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/stretchr/testify/assert"
)

// ==================== Color Tests ====================

func TestColors(t *testing.T) {
	t.Run("All Colors Are Defined", func(t *testing.T) {
		colors := []lipgloss.Color{
			ColorPrimary,
			ColorSecondary,
			ColorSuccess,
			ColorError,
			ColorWarning,
			ColorText,
			ColorSubtext,
			ColorMuted,
			ColorBg,
			ColorHighlight,
		}

		for i, color := range colors {
			assert.NotEmpty(t, string(color), "Color at index %d should not be empty", i)
		}
	})

	t.Run("Colors Are Valid Hex Values", func(t *testing.T) {
		// Colors should be hex values starting with #
		assert.Equal(t, "#7D56F4", string(ColorPrimary))
		assert.Equal(t, "#00ADD8", string(ColorSecondary))
		assert.Equal(t, "#23D18B", string(ColorSuccess))
		assert.Equal(t, "#F43F5E", string(ColorError))
		assert.Equal(t, "#FFB86C", string(ColorWarning))
	})
}

// ==================== Style Tests ====================

func TestStyles(t *testing.T) {
	t.Run("BaseStyle Is Defined", func(t *testing.T) {
		// BaseStyle should not panic when rendered
		result := BaseStyle.Render("test")
		assert.NotEmpty(t, result)
	})

	t.Run("HeaderStyle Is Defined", func(t *testing.T) {
		result := HeaderStyle.Render("Header")
		assert.NotEmpty(t, result)
		assert.Contains(t, result, "Header")
	})

	t.Run("StatusBarStyle Is Defined", func(t *testing.T) {
		result := StatusBarStyle.Render("Status")
		assert.NotEmpty(t, result)
	})

	t.Run("StatusKeyStyle Is Defined", func(t *testing.T) {
		result := StatusKeyStyle.Render("Key:")
		assert.NotEmpty(t, result)
	})

	t.Run("StatusValStyle Is Defined", func(t *testing.T) {
		result := StatusValStyle.Render("Value")
		assert.NotEmpty(t, result)
	})

	t.Run("ListItemStyle Is Defined", func(t *testing.T) {
		result := ListItemStyle.Render("Item")
		assert.NotEmpty(t, result)
	})

	t.Run("ListItemSelectedStyle Is Defined", func(t *testing.T) {
		result := ListItemSelectedStyle.Render("Selected Item")
		assert.NotEmpty(t, result)
	})

	t.Run("InputStyle Is Defined", func(t *testing.T) {
		result := InputStyle.Render("Input")
		assert.NotEmpty(t, result)
	})

	t.Run("InputFocusedStyle Is Defined", func(t *testing.T) {
		result := InputFocusedStyle.Render("Focused Input")
		assert.NotEmpty(t, result)
	})

	t.Run("CardStyle Is Defined", func(t *testing.T) {
		result := CardStyle.Render("Card Content")
		assert.NotEmpty(t, result)
	})

	t.Run("CardTitleStyle Is Defined", func(t *testing.T) {
		result := CardTitleStyle.Render("Card Title")
		assert.NotEmpty(t, result)
	})
}

// ==================== Helper Function Tests ====================

func TestRenderHeader(t *testing.T) {
	t.Run("Renders Header String", func(t *testing.T) {
		result := RenderHeader("Test Header")
		assert.NotEmpty(t, result)
		assert.Contains(t, result, "Test Header")
	})

	t.Run("Renders Empty Header", func(t *testing.T) {
		result := RenderHeader("")
		assert.NotEmpty(t, result) // Still has styling
	})

	t.Run("Renders Long Header", func(t *testing.T) {
		longTitle := "This is a very long header title that might wrap"
		result := RenderHeader(longTitle)
		assert.NotEmpty(t, result)
		assert.Contains(t, result, "very long header")
	})
}

func TestRenderStatusBar(t *testing.T) {
	t.Run("Renders Single Item", func(t *testing.T) {
		result := RenderStatusBar("Item 1")
		assert.NotEmpty(t, result)
	})

	t.Run("Renders Multiple Items", func(t *testing.T) {
		result := RenderStatusBar("Key:", "Value", " | ", "Status")
		assert.NotEmpty(t, result)
	})

	t.Run("Renders Empty Items", func(t *testing.T) {
		result := RenderStatusBar()
		assert.NotEmpty(t, result) // Still has styling wrapper
	})
}

// ==================== Style Composition Tests ====================

func TestStyleComposition(t *testing.T) {
	t.Run("Styles Can Be Combined", func(t *testing.T) {
		// Test that styles can be applied together
		combined := StatusKeyStyle.Render("Key: ") + StatusValStyle.Render("Value")
		assert.NotEmpty(t, combined)
	})

	t.Run("Styles Support Width", func(t *testing.T) {
		// CardStyle has Width capability
		result := CardStyle.Width(40).Render("Content")
		assert.NotEmpty(t, result)
	})

	t.Run("Styles Support Foreground Override", func(t *testing.T) {
		// StatusValStyle foreground can be overridden
		result := StatusValStyle.Foreground(ColorSuccess).Render("Success!")
		assert.NotEmpty(t, result)
	})
}
