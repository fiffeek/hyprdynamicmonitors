package tui

type MonitorSelected struct {
	Index int
}

type MonitorUnselected struct{}

type WindowResized struct {
	height int
	width  int
}
