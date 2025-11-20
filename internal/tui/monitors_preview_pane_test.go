package tui_test

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/fiffeek/hyprdynamicmonitors/internal/testutils"
	"github.com/fiffeek/hyprdynamicmonitors/internal/tui"
	"github.com/stretchr/testify/assert"
)

func TestMonitorsPreviewPane_Update(t *testing.T) {
	x100 := 100
	y200 := 200

	tests := []struct {
		name                  string
		setupMsg              tea.Cmd
		msg                   tea.Msg
		expectedSelectedIndex int
		expectGridX           bool
		expectGridY           bool
		expectedPanning       bool
		expectedSnapping      bool
		expectedFollowMode    bool
	}{
		{
			name:                  "ShowGridLineCommand sets grid lines",
			msg:                   tui.ShowGridLineCmd(&x100, &y200)(),
			expectedSelectedIndex: -1,
			expectGridX:           true,
			expectGridY:           true,
			expectedPanning:       false,
			expectedSnapping:      true,
			expectedFollowMode:    false,
		},
		{
			name:                  "MoveMonitorCommand clears grid lines",
			setupMsg:              tui.ShowGridLineCmd(&x100, &y200),
			msg:                   tui.MoveMonitorCmd(0, tui.DeltaMore, tui.DeltaNone)(),
			expectedSelectedIndex: -1,
			expectGridX:           false,
			expectGridY:           false,
			expectedPanning:       false,
			expectedSnapping:      true,
			expectedFollowMode:    false,
		},
		{
			name:                  "MonitorBeingEdited selects monitor",
			msg:                   tui.MonitorBeingEdited{ListIndex: 1},
			expectedSelectedIndex: 1,
			expectGridX:           false,
			expectGridY:           false,
			expectedPanning:       false,
			expectedSnapping:      true,
			expectedFollowMode:    false,
		},
		{
			name:                  "MonitorUnselected clears selection",
			msg:                   tui.MonitorUnselected{},
			expectedSelectedIndex: -1,
			expectGridX:           false,
			expectGridY:           false,
			expectedPanning:       false,
			expectedSnapping:      true,
			expectedFollowMode:    false,
		},
		{
			name:                  "StateChanged updates panning and snapping",
			setupMsg:              tui.StateChangedCmd(tui.AppState{Panning: true, Snapping: false, MonitorFollowMode: true}),
			expectedSelectedIndex: -1,
			expectGridX:           false,
			expectGridY:           false,
			expectedPanning:       true,
			expectedSnapping:      false,
			expectedFollowMode:    true,
		},
		{
			name:                  "other message does nothing",
			msg:                   tea.WindowSizeMsg{},
			expectedSelectedIndex: -1,
			expectGridX:           false,
			expectGridY:           false,
			expectedPanning:       false,
			expectedSnapping:      true,
			expectedFollowMode:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			monitors := []*tui.MonitorSpec{
				{Name: "eDP-1", X: 0, Y: 0, Width: 1920, Height: 1080, Scale: 1.0},
				{Name: "HDMI-1", X: 1920, Y: 0, Width: 1920, Height: 1080, Scale: 1.0},
			}
			cfg := testutils.NewTestConfig(t).Get()
			colors := tui.NewColorsManager(cfg)
			pane := tui.NewMonitorsPreviewPane(monitors, colors)

			if tt.setupMsg != nil {
				pane.Update(tt.setupMsg())
			}

			if tt.msg != nil {
				pane.Update(tt.msg)
			}

			assert.Equal(t, tt.expectedSelectedIndex, pane.GetSelectedIndex())
			assert.Equal(t, tt.expectGridX, pane.GetSnapGridX() != nil)
			assert.Equal(t, tt.expectGridY, pane.GetSnapGridY() != nil)
			assert.Equal(t, tt.expectedPanning, pane.GetPanning())
			assert.Equal(t, tt.expectedSnapping, pane.GetSnapping())
			assert.Equal(t, tt.expectedFollowMode, pane.GetFollowMonitor())
		})
	}
}

func TestMonitorsPreviewPane_KeyboardControls(t *testing.T) {
	tests := []struct {
		name             string
		setupCmd         tea.Cmd
		key              tea.KeyMsg
		expectedPanX     int
		expectedPanY     int
		expectedVirtualW int
		expectedVirtualH int
	}{
		{
			name:             "up key moves pan up when panning",
			setupCmd:         tui.StateChangedCmd(tui.AppState{Panning: true}),
			key:              tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}},
			expectedPanX:     960,
			expectedPanY:     440,
			expectedVirtualW: 3456,
			expectedVirtualH: 3456,
		},
		{
			name:             "down key moves pan down when panning",
			setupCmd:         tui.StateChangedCmd(tui.AppState{Panning: true}),
			key:              tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}},
			expectedPanX:     960,
			expectedPanY:     640,
			expectedVirtualW: 3456,
			expectedVirtualH: 3456,
		},
		{
			name:             "left key moves pan left when panning",
			setupCmd:         tui.StateChangedCmd(tui.AppState{Panning: true}),
			key:              tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'h'}},
			expectedPanX:     860,
			expectedPanY:     540,
			expectedVirtualW: 3456,
			expectedVirtualH: 3456,
		},
		{
			name:             "right key moves pan right when panning",
			setupCmd:         tui.StateChangedCmd(tui.AppState{Panning: true}),
			key:              tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'l'}},
			expectedPanX:     1060,
			expectedPanY:     540,
			expectedVirtualW: 3456,
			expectedVirtualH: 3456,
		},
		{
			name:             "up key does nothing when not panning",
			key:              tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}},
			expectedPanX:     960,
			expectedPanY:     540,
			expectedVirtualW: 3456,
			expectedVirtualH: 3456,
		},
		{
			name:             "center key resets pan to 0,0",
			setupCmd:         tui.StateChangedCmd(tui.AppState{Panning: true}),
			key:              tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}},
			expectedPanX:     0,
			expectedPanY:     0,
			expectedVirtualW: 3456,
			expectedVirtualH: 3456,
		},
		{
			name:             "zoom in decreases virtual dimensions",
			key:              tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'+'}},
			expectedPanX:     960,
			expectedPanY:     540,
			expectedVirtualW: 3141,
			expectedVirtualH: 3141,
		},
		{
			name:             "zoom out increases virtual dimensions",
			key:              tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'-'}},
			expectedPanX:     960,
			expectedPanY:     540,
			expectedVirtualW: 3801,
			expectedVirtualH: 3801,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			monitors := []*tui.MonitorSpec{
				{Name: "eDP-1", X: 0, Y: 0, Width: 1920, Height: 1080, Scale: 1.0},
			}
			cfg := testutils.NewTestConfig(t).Get()
			colors := tui.NewColorsManager(cfg)
			pane := tui.NewMonitorsPreviewPane(monitors, colors)

			if tt.setupCmd != nil {
				pane.Update(tt.setupCmd())
			}

			pane.Update(tt.key)

			assert.Equal(t, tt.expectedPanX, pane.GetPanX())
			assert.Equal(t, tt.expectedPanY, pane.GetPanY())
			assert.Equal(t, tt.expectedVirtualW, pane.GetVirtualWidth())
			assert.Equal(t, tt.expectedVirtualH, pane.GetVirtualHeight())
		})
	}
}
