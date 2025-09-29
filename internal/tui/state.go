package tui

import (
	"strings"

	"github.com/fiffeek/hyprdynamicmonitors/internal/config"
)

type ViewMode int

const (
	MonitorsListView ViewMode = iota
	ConfigView
)

func (v ViewMode) String() string {
	switch v {
	case MonitorsListView:
		return "Monitors"
	case ConfigView:
		return "Config"
	default:
		return "Unknown"
	}
}

type AppState struct {
	EditingMonitor         bool
	Panning                bool
	Scaling                bool
	ModeSelection          bool
	MirrorSelection        bool
	Fullscreen             bool
	MonitorEditedListIndex int
	Snapping               bool
	ProfileNameRequested   bool
	ShowConfirmationPrompt bool
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
	return s.EditingMonitor || s.Fullscreen || s.ModeSelection || s.Scaling || s.Panning || s.MirrorSelection
}

func (s AppState) IsPanning() bool {
	return s.Panning
}

type RootState struct {
	CurrentViewIndex int
	State            AppState
	monitors         []*MonitorSpec
	viewModes        []ViewMode
	config           *config.Config
}

func NewState(monitors []*MonitorSpec, cfg *config.Config) *RootState {
	viewModes := []ViewMode{MonitorsListView}
	if cfg != nil {
		viewModes = append(viewModes, ConfigView)
	}
	return &RootState{
		CurrentViewIndex: 0,
		State: AppState{
			Snapping: true,
		},
		monitors:  monitors,
		config:    cfg,
		viewModes: viewModes,
	}
}

func (r *RootState) ToggleProfileNameRequested() {
	r.State.ProfileNameRequested = !r.State.ProfileNameRequested
}

func (r *RootState) ToggleFullscreen() {
	r.State.Fullscreen = !r.State.Fullscreen
}

func (r *RootState) ToggleSnapping() {
	r.State.Snapping = !r.State.Snapping
}

func (r *RootState) TogglePanning() {
	r.State.Panning = !r.State.Panning
}

func (r *RootState) ToggleConfirmationPrompt() {
	r.State.ShowConfirmationPrompt = !r.State.ShowConfirmationPrompt
}

func (r *RootState) SetMonitorEditState(msg MonitorBeingEdited) {
	r.State.EditingMonitor = true
	r.State.Scaling = msg.Scaling
	r.State.ModeSelection = msg.ModesEditor
	r.State.MonitorEditedListIndex = msg.ListIndex
	r.State.MirrorSelection = msg.MirroringMode
}

func (r *RootState) ClearMonitorEditState() {
	r.State.ModeSelection = false
	r.State.Scaling = false
	r.State.EditingMonitor = false
	r.State.MonitorEditedListIndex = -1
	r.State.MirrorSelection = false
}

func (r *RootState) CurrentView() ViewMode {
	return r.viewModes[r.CurrentViewIndex%len(r.viewModes)]
}

func (r *RootState) NextView() {
	r.CurrentViewIndex = (r.CurrentViewIndex + 1) % len(r.viewModes)
}

func (r *RootState) HasMoreThanOneView() bool {
	return len(r.viewModes) > 1
}
