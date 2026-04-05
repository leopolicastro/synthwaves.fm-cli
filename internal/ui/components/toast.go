package components

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/leo/synthwaves-cli/internal/ui/theme"
)

type ToastMsg struct {
	Message string
	IsError bool
}

type toastClearMsg struct{}

type Toast struct {
	message string
	isError bool
	visible bool
}

func NewToast() Toast {
	return Toast{}
}

func (t Toast) Update(msg tea.Msg) (Toast, tea.Cmd) {
	switch msg := msg.(type) {
	case ToastMsg:
		t.message = msg.Message
		t.isError = msg.IsError
		t.visible = true
		return t, tea.Tick(3*time.Second, func(time.Time) tea.Msg {
			return toastClearMsg{}
		})
	case toastClearMsg:
		t.visible = false
	}
	return t, nil
}

func (t Toast) View() string {
	if !t.visible {
		return ""
	}
	style := lipgloss.NewStyle().Padding(0, 2).Bold(true)
	if t.isError {
		style = style.Foreground(theme.ColorRed).
			BorderStyle(lipgloss.RoundedBorder()).BorderForeground(theme.ColorRed)
	} else {
		style = style.Foreground(theme.ColorGreen).
			BorderStyle(lipgloss.RoundedBorder()).BorderForeground(theme.ColorGreen)
	}
	return style.Render(t.message)
}
