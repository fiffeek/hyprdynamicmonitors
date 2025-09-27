package tui

type MonitorBeingEdited struct {
	Index   int
	Scaling bool
}

type MonitorUnselected struct{}

type StateChanged struct {
	state AppState
}
