package tui

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
)

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
