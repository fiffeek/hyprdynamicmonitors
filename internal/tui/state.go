package tui

import "strings"

type ViewMode int

const (
	MonitorsListView ViewMode = iota
	ProfileView
)

type AppState struct {
	EditingMonitor bool
	Panning        bool
	Scaling        bool
	ModeSelection  bool
	Fullscreen     bool
}

func (s AppState) String() string {
	modes := []string{}
	if s.Fullscreen {
		modes = append(modes, "Fullscreen")
	}
	if s.Panning {
		modes = append(modes, "Panning")
	}
	return strings.Join(modes, " ")
}

func (s AppState) Editing() bool {
	return s.EditingMonitor || s.Fullscreen || s.ModeSelection || s.Scaling || s.Panning
}

func (s AppState) IsPanning() bool {
	return s.Panning
}

type RootState struct {
	CurrentView ViewMode
	State       AppState
}

func NewState() *RootState {
	return &RootState{
		CurrentView: MonitorsListView,
		State:       AppState{},
	}
}

func (r *RootState) ToggleFullscreen() {
	r.State.Fullscreen = !r.State.Fullscreen
}

func (r *RootState) TogglePanning() {
	r.State.Panning = !r.State.Panning
}

func (r *RootState) SetMonitorEditState(msg MonitorBeingEdited) {
	r.State.EditingMonitor = true
	r.State.Scaling = msg.Scaling
}

func (r *RootState) ClearMonitorEditState() {
	r.State.Scaling = false
	r.State.EditingMonitor = false
}
