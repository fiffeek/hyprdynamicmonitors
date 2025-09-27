package tui

type MonitorSelected struct {
	Index int
}

type MonitorUnselected struct{}

type StateChanged struct {
	state AppState
}
