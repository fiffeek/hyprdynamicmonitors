package tui_test

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/fiffeek/hyprdynamicmonitors/internal/testutils"
	"github.com/fiffeek/hyprdynamicmonitors/internal/tui"
	"github.com/stretchr/testify/assert"
)

func TestMirrorList_Update(t *testing.T) {
	tests := []struct {
		name        string
		key         tea.KeyMsg
		setupItems  bool
		expectedMsg tea.Msg
	}{
		{
			name:        "esc key closes mirror list",
			key:         tea.KeyMsg{Type: tea.KeyEsc},
			setupItems:  false,
			expectedMsg: tui.CloseMonitorMirrorListCommand{},
		},
		{
			name:        "up key triggers preview through delegate",
			key:         tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}},
			setupItems:  true,
			expectedMsg: tui.ChangeMirrorPreviewCommand{MirrorOf: "none"},
		},
		{
			name:        "down key triggers preview through delegate",
			key:         tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}},
			setupItems:  true,
			expectedMsg: tui.ChangeMirrorPreviewCommand{MirrorOf: "DP-1"},
		},
		{
			name:        "enter key changes mirror through delegate",
			key:         tea.KeyMsg{Type: tea.KeyEnter},
			setupItems:  true,
			expectedMsg: tui.ChangeMirrorCommand{MirrorOf: "HDMI-1"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			monitor := &tui.MonitorSpec{
				Name:   "eDP-1",
				Mirror: "HDMI-1",
			}

			monitors := []*tui.MonitorSpec{
				monitor,
				{Name: "HDMI-1"},
				{Name: "DP-1"},
			}
			cfg := testutils.NewTestConfig(t).Get()
			colors := tui.NewColorsManager(cfg)
			mirrorList := tui.NewMirrorList(monitors, colors)

			if tt.setupItems {
				mirrorList.SetItems(monitor)
			}

			cmd := mirrorList.Update(tt.key)

			assert.NotNil(t, cmd)
			msg := cmd()
			assert.Equal(t, tt.expectedMsg, msg)
		})
	}
}
