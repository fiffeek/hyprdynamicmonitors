package tui_test

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/fiffeek/hyprdynamicmonitors/internal/testutils"
	"github.com/fiffeek/hyprdynamicmonitors/internal/tui"
	"github.com/stretchr/testify/assert"
)

func TestMonitorModeList_Update(t *testing.T) {
	tests := []struct {
		name        string
		key         tea.KeyMsg
		setupItems  bool
		expectedMsg tea.Msg
	}{
		{
			name:        "esc key closes mode list",
			key:         tea.KeyMsg{Type: tea.KeyEsc},
			setupItems:  false,
			expectedMsg: tui.CloseMonitorModeListCommand{},
		},
		{
			name:        "up key triggers preview through delegate",
			key:         tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}},
			setupItems:  true,
			expectedMsg: tui.ChangeModePreviewCommand{Mode: "1920x1080@60.00Hz"},
		},
		{
			name:        "down key triggers preview through delegate",
			key:         tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}},
			setupItems:  true,
			expectedMsg: tui.ChangeModePreviewCommand{Mode: "1366x768@60.00Hz"},
		},
		{
			name:        "enter key changes mode through delegate",
			key:         tea.KeyMsg{Type: tea.KeyEnter},
			setupItems:  true,
			expectedMsg: tui.ChangeModeCommand{Mode: "1920x1080@60.00Hz"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			monitor := &tui.MonitorSpec{
				Name:           "eDP-1",
				Width:          1920,
				Height:         1080,
				RefreshRate:    60.00,
				AvailableModes: []string{"1920x1080@60.00Hz", "1366x768@60.00Hz"},
			}

			cfg := testutils.NewTestConfig(t).Get()
			colors := tui.NewColorsManager(cfg)
			monitors := []*tui.MonitorSpec{monitor}
			modeList := tui.NewMonitorModeList(monitors, colors)

			if tt.setupItems {
				modeList.SetItems(monitor)
			}

			cmd := modeList.Update(tt.key)

			assert.NotNil(t, cmd)
			msg := cmd()
			assert.Equal(t, tt.expectedMsg, msg)
		})
	}
}
