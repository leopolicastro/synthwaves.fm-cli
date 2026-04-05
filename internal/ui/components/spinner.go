package components

import (
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/leo/synthwaves-cli/internal/ui/theme"
)

type Spinner struct {
	spinner spinner.Model
	Label   string
}

func NewSpinner(label string) Spinner {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(theme.ColorHotPink)
	return Spinner{spinner: s, Label: label}
}

func (s Spinner) Init() tea.Cmd {
	return s.spinner.Tick
}

func (s Spinner) Update(msg tea.Msg) (Spinner, tea.Cmd) {
	var cmd tea.Cmd
	s.spinner, cmd = s.spinner.Update(msg)
	return s, cmd
}

func (s Spinner) View() string {
	return s.spinner.View() + theme.MutedStyle.Render(" "+s.Label)
}
