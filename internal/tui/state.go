package tui

import "fmt"

type ViewMode int

const (
	MonitorsListView ViewMode = iota
	ProfileView
)

type AppState int

const (
	StateNavigating AppState = iota
	StateEditingMonitor
	StatePanning
	StateScaling
	StateModeSelection
	StateFullscreen
)

func (s AppState) String() string {
	switch s {
	case StateNavigating:
		return "Navigating"
	case StateEditingMonitor:
		return "Editing Monitor"
	case StatePanning:
		return "Panning"
	case StateScaling:
		return "Scaling"
	case StateModeSelection:
		return "Mode Selection"
	case StateFullscreen:
		return "Fullscreen"
	default:
		return fmt.Sprintf("Unknown (%d)", s)
	}
}

type RootState struct {
	CurrentView ViewMode
	State       AppState
}

func NewState() *RootState {
	return &RootState{
		CurrentView: MonitorsListView,
		State:       StateNavigating,
	}
}

func (r *RootState) ChangeState(state AppState) {
	r.State = state
}

func (r *RootState) ToogleState(state AppState) bool {
	if r.State == state {
		r.State = StateNavigating
		return false
	}
	r.ChangeState(state)
	return true
}
