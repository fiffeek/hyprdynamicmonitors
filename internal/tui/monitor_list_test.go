package tui_test

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/fiffeek/hyprdynamicmonitors/internal/testutils"
	"github.com/fiffeek/hyprdynamicmonitors/internal/tui"
	"github.com/fiffeek/hyprdynamicmonitors/internal/utils"
	"github.com/stretchr/testify/assert"
)

func TestMonitorList_Update(t *testing.T) {
	tests := []struct {
		name        string
		msg         tea.Msg
		setupState  *tui.AppState
		expectedMsg tea.Msg
	}{
		{
			name: "MonitorBeingEdited message",
			msg: tui.MonitorBeingEdited{
				ListIndex:     0,
				Scaling:       true,
				MonitorID:     1,
				ModesEditor:   false,
				MirroringMode: false,
			},
			expectedMsg: nil,
		},
		{
			name:        "MonitorUnselected message",
			msg:         tui.MonitorUnselected{},
			expectedMsg: nil,
		},
		{
			name: "StateChanged message",
			msg: tui.StateChanged{
				State: tui.AppState{EditingMonitor: true},
			},
			expectedMsg: nil,
		},
		{
			name: "arrow key in panning mode returns nil",
			msg:  tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}},
			setupState: &tui.AppState{
				Panning: true,
			},
			expectedMsg: nil,
		},
		{
			name: "arrow up in editing mode triggers move",
			msg:  tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}},
			setupState: &tui.AppState{
				EditingMonitor: true,
			},
			expectedMsg: tui.MoveMonitorCommand{
				MonitorID: 1,
				StepX:     tui.DeltaNone,
				StepY:     tui.DeltaLess,
			},
		},
		{
			name: "arrow down in editing mode triggers move",
			msg:  tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}},
			setupState: &tui.AppState{
				EditingMonitor: true,
			},
			expectedMsg: tui.MoveMonitorCommand{
				MonitorID: 1,
				StepX:     tui.DeltaNone,
				StepY:     tui.DeltaMore,
			},
		},
		{
			name: "arrow left in editing mode triggers move",
			msg:  tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'h'}},
			setupState: &tui.AppState{
				EditingMonitor: true,
			},
			expectedMsg: tui.MoveMonitorCommand{
				MonitorID: 1,
				StepX:     tui.DeltaLess,
				StepY:     tui.DeltaNone,
			},
		},
		{
			name: "arrow right in editing mode triggers move",
			msg:  tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'l'}},
			setupState: &tui.AppState{
				EditingMonitor: true,
			},
			expectedMsg: tui.MoveMonitorCommand{
				MonitorID: 1,
				StepX:     tui.DeltaMore,
				StepY:     tui.DeltaNone,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			monitor := &tui.MonitorSpec{
				Name: "eDP-1",
				ID:   utils.IntPtr(1),
			}

			cfg := testutils.NewTestConfig(t).Get()
			colors := tui.NewColorsManager(cfg)
			monitors := []*tui.MonitorSpec{monitor}
			monitorList := tui.NewMonitorList(monitors, colors)

			if tt.setupState != nil {
				monitorList.Update(tui.StateChanged{State: *tt.setupState})
				if tt.setupState.EditingMonitor {
					monitorList.Update(tui.MonitorBeingEdited{
						ListIndex: 0,
						MonitorID: 1,
					})
				}
			}

			cmd := monitorList.Update(tt.msg)

			if tt.expectedMsg == nil {
				return
			}

			assert.NotNil(t, cmd)
			msg := cmd()
			assert.Equal(t, tt.expectedMsg, msg)
		})
	}
}

func TestMonitorList_DelegateUpdate(t *testing.T) {
	tests := []struct {
		name         string
		key          tea.KeyMsg
		setupEditing bool
		expectedMsg  tea.Msg
	}{
		{
			name:         "enter key selects monitor",
			key:          tea.KeyMsg{Type: tea.KeyEnter},
			setupEditing: false,
			expectedMsg: tui.MonitorBeingEdited{
				ListIndex:     0,
				Scaling:       false,
				MonitorID:     1,
				ModesEditor:   false,
				MirroringMode: false,
			},
		},
		{
			name:         "enter key unselects monitor",
			key:          tea.KeyMsg{Type: tea.KeyEnter},
			setupEditing: true,
			expectedMsg:  tui.MonitorUnselected{},
		},
		{
			name:         "r key rotates monitor",
			key:          tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}},
			setupEditing: true,
			expectedMsg:  tui.RotateMonitorCommand{MonitorID: 1},
		},
		{
			name:         "v key toggles VRR",
			key:          tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'v'}},
			setupEditing: true,
			expectedMsg:  tui.ToggleMonitorVRRCommand{MonitorID: 1},
		},
		{
			name:         "e key toggles monitor",
			key:          tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}},
			setupEditing: true,
			expectedMsg:  tui.ToggleMonitorCommand{MonitorID: 1},
		},
		{
			name:         "s key enters scale mode",
			key:          tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}},
			setupEditing: true,
			expectedMsg: tui.MonitorBeingEdited{
				ListIndex:     0,
				Scaling:       true,
				MonitorID:     1,
				ModesEditor:   false,
				MirroringMode: false,
			},
		},
		{
			name:         "m key enters mode selection",
			key:          tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'m'}},
			setupEditing: true,
			expectedMsg: tui.MonitorBeingEdited{
				ListIndex:     0,
				Scaling:       false,
				MonitorID:     1,
				ModesEditor:   true,
				MirroringMode: false,
			},
		},
		{
			name:         "i key enters mirroring mode",
			key:          tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'i'}},
			setupEditing: true,
			expectedMsg: tui.MonitorBeingEdited{
				ListIndex:     0,
				Scaling:       false,
				MonitorID:     1,
				ModesEditor:   false,
				MirroringMode: true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			monitor := &tui.MonitorSpec{
				Name: "eDP-1",
				ID:   utils.IntPtr(1),
			}

			cfg := testutils.NewTestConfig(t).Get()
			colors := tui.NewColorsManager(cfg)
			monitors := []*tui.MonitorSpec{monitor}
			monitorList := tui.NewMonitorList(monitors, colors)

			if tt.setupEditing {
				monitorList.Update(tea.KeyMsg{Type: tea.KeyEnter})
			}

			cmd := monitorList.Update(tt.key)

			assert.NotNil(t, cmd)
			msg := cmd()
			assert.Equal(t, tt.expectedMsg, msg)
		})
	}
}
