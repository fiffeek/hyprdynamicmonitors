package tui

import "github.com/charmbracelet/bubbles/key"

type keyMap struct {
	Tab           key.Binding
	Enter         key.Binding
	Quit          key.Binding
	Up            key.Binding
	Center        key.Binding
	Down          key.Binding
	Left          key.Binding
	Right         key.Binding
	ShowFullHelp  key.Binding
	CloseFullHelp key.Binding
	Pan           key.Binding
	Fullscreen    key.Binding
}

var rootKeyMap = keyMap{
	Tab: key.NewBinding(
		key.WithKeys("tab"),
		key.WithHelp("tab", "switch view"),
	),
	Enter: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "select"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "ctrl+c"),
		key.WithHelp("q", "quit"),
	),
	Up: key.NewBinding(
		key.WithKeys("up", "k"),
		key.WithHelp("↑/k", "up"),
	),
	Down: key.NewBinding(
		key.WithKeys("down", "j"),
		key.WithHelp("↓/j", "down"),
	),
	Left: key.NewBinding(
		key.WithKeys("left", "h"),
		key.WithHelp("←/h", "left"),
	),
	Right: key.NewBinding(
		key.WithKeys("right", "l"),
		key.WithHelp("→/l", "right"),
	),
	ShowFullHelp: key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "more"),
	),
	CloseFullHelp: key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "close help"),
	),
	Pan: key.NewBinding(
		key.WithKeys("p"),
		key.WithHelp("p", "pan (move freely on the grid)"),
	),
	Fullscreen: key.NewBinding(
		key.WithKeys("f"),
		key.WithHelp("f", "fullscreen the preview"),
	),
	Center: key.NewBinding(
		key.WithKeys("c"),
		key.WithHelp("c", "moves the grid to 0, 0"),
	),
}
