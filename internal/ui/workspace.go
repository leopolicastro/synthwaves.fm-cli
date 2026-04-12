package ui

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/leo/synthwaves-cli/internal/ui/theme"
)

type WorkspaceProps struct {
	Width             int
	Height            int
	Nav               string
	Detail            string
	NavPreferredWidth int
	NavMinWidth       int
	DetailMinWidth    int
	StackAt           int
}

func RenderWorkspace(props WorkspaceProps) string {
	if props.Width <= 0 {
		return lipgloss.JoinVertical(lipgloss.Left, props.Nav, props.Detail)
	}
	if props.Height <= 0 {
		props.Height = 12
	}

	if props.NavMinWidth <= 0 {
		props.NavMinWidth = 24
	}
	if props.DetailMinWidth <= 0 {
		props.DetailMinWidth = 32
	}
	if props.StackAt <= 0 {
		props.StackAt = 86
	}

	gap := 2
	if props.Width < props.StackAt || props.Width < props.NavMinWidth+props.DetailMinWidth+gap {
		return lipgloss.JoinVertical(
			lipgloss.Left,
			panelStyle(true, props.Width, props.Height).Render(props.Nav),
			panelStyle(false, props.Width, props.Height).Render(props.Detail),
		)
	}

	navWidth := props.NavPreferredWidth
	if navWidth < props.NavMinWidth {
		navWidth = props.NavMinWidth
	}
	maxNav := props.Width - props.DetailMinWidth - gap
	if navWidth > maxNav {
		navWidth = maxNav
	}
	detailWidth := props.Width - navWidth - gap

	return lipgloss.JoinHorizontal(
		lipgloss.Top,
		panelStyle(true, navWidth, props.Height).Render(props.Nav),
		panelStyle(false, detailWidth, props.Height).Render(props.Detail),
	)
}

func panelStyle(nav bool, width, height int) lipgloss.Style {
	border := theme.ColorGridPurple
	if nav {
		border = theme.ColorHotPink
	}
	return lipgloss.NewStyle().
		Width(width).
		Height(height).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(border).
		Padding(0, 1)
}
