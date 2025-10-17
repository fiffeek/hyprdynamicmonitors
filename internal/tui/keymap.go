package tui

import "github.com/charmbracelet/bubbles/key"

type keyMap struct {
	Tab               key.Binding
	Enter             key.Binding
	Quit              key.Binding
	Up                key.Binding
	Center            key.Binding
	Down              key.Binding
	Left              key.Binding
	Right             key.Binding
	ShowFullHelp      key.Binding
	CloseFullHelp     key.Binding
	Pan               key.Binding
	Fullscreen        key.Binding
	ZoomIn            key.Binding
	ZoomOut           key.Binding
	NextPage          key.Binding
	ToggleSnapping    key.Binding
	ApplyHypr         key.Binding
	Back              key.Binding
	EditHDMConfig     key.Binding
	FollowMonitor     key.Binding
	ExpandHyprPreview key.Binding
	ResetZoom         key.Binding
}

var rootKeyMap = keyMap{
	ToggleSnapping: key.NewBinding(
		key.WithKeys("S"),
		key.WithHelp("S", "toggle snapping"),
	),
	ApplyHypr: key.NewBinding(
		key.WithKeys("A"),
		key.WithHelp("A", "apply (ephemeral)"),
	),
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
		key.WithKeys("F"),
		key.WithHelp("F", "fullscreen the preview"),
	),
	Center: key.NewBinding(
		key.WithKeys("c"),
		key.WithHelp("c", "moves the grid to 0, 0"),
	),
	ZoomIn: key.NewBinding(
		key.WithKeys("+"),
		key.WithHelp("+", "zoom in"),
	),
	ZoomOut: key.NewBinding(
		key.WithKeys("-"),
		key.WithHelp("-", "zoom out"),
	),
	NextPage: key.NewBinding(
		key.WithKeys("right", "l", "pgdown", "f", "d"),
		key.WithHelp("→/l/pgdn", "next page"),
	),
	Back: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "close"),
	),
	EditHDMConfig: key.NewBinding(
		key.WithKeys("C"),
		key.WithHelp("C", "edit HyprDynamicMonitors config"),
	),
	FollowMonitor: key.NewBinding(
		key.WithKeys("o"),
		key.WithHelp("o", "follow monitor"),
	),
	ExpandHyprPreview: key.NewBinding(
		key.WithKeys("H"),
		key.WithHelp("H", "expand/collapse hypr preview"),
	),
	ResetZoom: key.NewBinding(
		key.WithKeys("R"),
		key.WithHelp("R", "reset zoom"),
	),
}
