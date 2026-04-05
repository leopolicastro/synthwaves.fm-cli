package theme

import (
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
)

// SynthwaveHuhTheme returns a huh form theme drenched in neon.
func SynthwaveHuhTheme() *huh.Theme {
	t := huh.ThemeCharm()

	// Focused state -- bright neon
	t.Focused.Title = t.Focused.Title.
		Foreground(ColorNeonCyan).
		Bold(true)
	t.Focused.Description = t.Focused.Description.
		Foreground(ColorMutedLavender)
	t.Focused.TextInput.Cursor = t.Focused.TextInput.Cursor.
		Foreground(ColorHotPink)
	t.Focused.TextInput.Prompt = t.Focused.TextInput.Prompt.
		Foreground(ColorHotPink)
	t.Focused.TextInput.Text = t.Focused.TextInput.Text.
		Foreground(ColorGhostWhite)
	t.Focused.SelectSelector = t.Focused.SelectSelector.
		Foreground(ColorHotPink)
	t.Focused.SelectedOption = t.Focused.SelectedOption.
		Foreground(ColorNeonCyan)
	t.Focused.UnselectedOption = t.Focused.UnselectedOption.
		Foreground(ColorMutedLavender)
	t.Focused.FocusedButton = t.Focused.FocusedButton.
		Foreground(lipgloss.Color("#0a0a1a")).
		Background(ColorHotPink).
		Bold(true)
	t.Focused.BlurredButton = t.Focused.BlurredButton.
		Foreground(ColorMutedLavender).
		Background(ColorGridPurple)
	t.Focused.Base = t.Focused.Base.
		BorderForeground(ColorNeonPurple)

	// Blurred state -- dimmed
	t.Blurred.Title = t.Blurred.Title.
		Foreground(ColorMutedLavender)
	t.Blurred.TextInput.Text = t.Blurred.TextInput.Text.
		Foreground(ColorMutedLavender)

	return t
}

// FormContainerStyle wraps a form in a neon-bordered box.
var FormContainerStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.DoubleBorder()).
	BorderForeground(ColorNeonPurple).
	Padding(1, 3).
	MarginTop(1)
