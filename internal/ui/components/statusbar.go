package components

import (
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/lipgloss"
	"github.com/leo/synthwaves-cli/internal/ui/theme"
)

type StatusBar struct {
	Breadcrumbs []string
	KeyHints    []key.Binding
	Width       int
}

func NewStatusBar() StatusBar {
	return StatusBar{}
}

func (s StatusBar) View() string {
	var parts []string

	if len(s.Breadcrumbs) > 0 {
		crumbStyle := lipgloss.NewStyle().Foreground(theme.ColorNeonCyan).Bold(true)
		sepStyle := lipgloss.NewStyle().Foreground(theme.ColorGridPurple)
		var crumbs []string
		for _, c := range s.Breadcrumbs {
			crumbs = append(crumbs, crumbStyle.Render(c))
		}
		parts = append(parts, strings.Join(crumbs, sepStyle.Render(" > ")))
	}

	if len(s.KeyHints) > 0 {
		var hints []string
		keyStyle := lipgloss.NewStyle().Foreground(theme.ColorHotPink).Bold(true)
		descStyle := lipgloss.NewStyle().Foreground(theme.ColorMutedLavender)
		for _, k := range s.KeyHints {
			help := k.Help()
			hints = append(hints, keyStyle.Render(help.Key)+descStyle.Render(":"+help.Desc))
		}
		parts = append(parts, strings.Join(hints, "  "))
	}

	content := strings.Join(parts, "  ")
	return theme.StatusBarStyle.Width(s.Width).Render(content)
}
