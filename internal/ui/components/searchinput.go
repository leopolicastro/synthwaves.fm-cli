package components

import (
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/leo/synthwaves-cli/internal/ui/theme"
)

type SearchQueryMsg struct {
	Query string
}

type debounceTickMsg struct {
	id int
}

type SearchInput struct {
	input   textinput.Model
	tickID  int
	Focused bool
}

func NewSearchInput(placeholder string) SearchInput {
	ti := textinput.New()
	ti.Placeholder = placeholder
	ti.PromptStyle = theme.TitleStyle
	ti.Prompt = "/ "
	ti.TextStyle = theme.GhostStyle
	ti.PlaceholderStyle = theme.MutedStyle
	ti.CharLimit = 100

	return SearchInput{input: ti}
}

func (s SearchInput) Update(msg tea.Msg) (SearchInput, tea.Cmd) {
	if !s.Focused {
		return s, nil
	}

	switch msg := msg.(type) {
	case debounceTickMsg:
		if msg.id == s.tickID {
			return s, func() tea.Msg {
				return SearchQueryMsg{Query: s.input.Value()}
			}
		}
	}

	var cmd tea.Cmd
	prev := s.input.Value()
	s.input, cmd = s.input.Update(msg)

	if s.input.Value() != prev {
		s.tickID++
		id := s.tickID
		return s, tea.Batch(cmd, tea.Tick(300*time.Millisecond, func(time.Time) tea.Msg {
			return debounceTickMsg{id: id}
		}))
	}

	return s, cmd
}

func (s SearchInput) View() string {
	return s.input.View()
}

func (s SearchInput) Focus() SearchInput {
	s.Focused = true
	s.input.Focus()
	return s
}

func (s SearchInput) Blur() SearchInput {
	s.Focused = false
	s.input.Blur()
	return s
}

func (s SearchInput) Value() string {
	return s.input.Value()
}

func (s SearchInput) SetValue(v string) SearchInput {
	s.input.SetValue(v)
	return s
}
