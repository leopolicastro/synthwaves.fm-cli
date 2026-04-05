package components

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/leo/synthwaves-cli/internal/ui/theme"
)

type NavItem struct {
	Label string
	Key   string
}

type Nav struct {
	Items    []NavItem
	Selected int
	Focused  bool
	Height   int
}

func NewNav(items []NavItem) Nav {
	return Nav{
		Items:   items,
		Focused: true,
	}
}

type NavSelectMsg struct {
	Key string
}

func (n Nav) Update(msg tea.Msg) (Nav, tea.Cmd) {
	if !n.Focused {
		return n, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if n.Selected > 0 {
				n.Selected--
			}
		case "down", "j":
			if n.Selected < len(n.Items)-1 {
				n.Selected++
			}
		case "enter":
			return n, func() tea.Msg {
				return NavSelectMsg{Key: n.Items[n.Selected].Key}
			}
		}
	}

	return n, nil
}

func (n Nav) View() string {
	var b strings.Builder

	b.WriteString(theme.RenderSmallLogo())
	b.WriteString("\n\n")

	for i, item := range n.Items {
		if i == n.Selected && n.Focused {
			b.WriteString(theme.NavActivePrefix)
			b.WriteString(" ")
			b.WriteString(theme.NavActiveStyle.Render(item.Label))
		} else if i == n.Selected {
			b.WriteString("  ")
			b.WriteString(lipgloss.NewStyle().
				Foreground(theme.ColorGhostWhite).
				Render(item.Label))
		} else {
			b.WriteString(theme.NavItemStyle.Render(item.Label))
		}
		b.WriteString("\n")
	}

	return theme.SidebarStyle.Height(n.Height).Render(b.String())
}
