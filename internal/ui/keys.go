package ui

import "github.com/charmbracelet/bubbles/key"

// Global key bindings shared across all views.
type KeyMap struct {
	Quit      key.Binding
	Back      key.Binding
	Tab       key.Binding
	Search    key.Binding
	New       key.Binding
	Delete    key.Binding
	Refresh   key.Binding
	Help      key.Binding
	NextPage  key.Binding
	PrevPage  key.Binding
	Up        key.Binding
	Down      key.Binding
	Enter     key.Binding
	Favorite  key.Binding

	// Player controls
	PlayPause   key.Binding
	NextTrack   key.Binding
	PrevTrack   key.Binding
	Shuffle     key.Binding
	Repeat      key.Binding
	Stop        key.Binding
	VolumeUp    key.Binding
	VolumeDown  key.Binding
	SeekForward key.Binding
	SeekBack    key.Binding
}

var Keys = KeyMap{
	Quit: key.NewBinding(
		key.WithKeys("q", "ctrl+c"),
		key.WithHelp("q", "quit"),
	),
	Back: key.NewBinding(
		key.WithKeys("esc", "backspace"),
		key.WithHelp("esc", "back"),
	),
	Tab: key.NewBinding(
		key.WithKeys("tab"),
		key.WithHelp("tab", "switch focus"),
	),
	Search: key.NewBinding(
		key.WithKeys("/"),
		key.WithHelp("/", "search"),
	),
	New: key.NewBinding(
		key.WithKeys("n"),
		key.WithHelp("n", "new"),
	),
	Delete: key.NewBinding(
		key.WithKeys("d"),
		key.WithHelp("d", "delete"),
	),
	Refresh: key.NewBinding(
		key.WithKeys("r"),
		key.WithHelp("r", "refresh"),
	),
	Help: key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "help"),
	),
	NextPage: key.NewBinding(
		key.WithKeys("right", "l"),
		key.WithHelp("->", "next page"),
	),
	PrevPage: key.NewBinding(
		key.WithKeys("left", "h"),
		key.WithHelp("<-", "prev page"),
	),
	Up: key.NewBinding(
		key.WithKeys("up", "k"),
		key.WithHelp("k", "up"),
	),
	Down: key.NewBinding(
		key.WithKeys("down", "j"),
		key.WithHelp("j", "down"),
	),
	Enter: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "select"),
	),
	Favorite: key.NewBinding(
		key.WithKeys("f"),
		key.WithHelp("f", "favorite"),
	),
	PlayPause: key.NewBinding(
		key.WithKeys(" "),
		key.WithHelp("space", "play/pause"),
	),
	NextTrack: key.NewBinding(
		key.WithKeys("]"),
		key.WithHelp("]", "next"),
	),
	PrevTrack: key.NewBinding(
		key.WithKeys("["),
		key.WithHelp("[", "prev"),
	),
	Shuffle: key.NewBinding(
		key.WithKeys("s"),
		key.WithHelp("s", "shuffle"),
	),
	Repeat: key.NewBinding(
		key.WithKeys("R"),
		key.WithHelp("R", "repeat"),
	),
	Stop: key.NewBinding(
		key.WithKeys("x"),
		key.WithHelp("x", "stop"),
	),
	VolumeUp: key.NewBinding(
		key.WithKeys("+", "="),
		key.WithHelp("+", "vol up"),
	),
	VolumeDown: key.NewBinding(
		key.WithKeys("-"),
		key.WithHelp("-", "vol down"),
	),
	SeekForward: key.NewBinding(
		key.WithKeys(".", ">"),
		key.WithHelp(".", "seek fwd"),
	),
	SeekBack: key.NewBinding(
		key.WithKeys(",", "<"),
		key.WithHelp(",", "seek back"),
	),
}
