package tui_test

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/fiffeek/hyprdynamicmonitors/internal/testutils"
	"github.com/fiffeek/hyprdynamicmonitors/internal/tui"
	"github.com/fiffeek/hyprdynamicmonitors/internal/utils"
	"github.com/stretchr/testify/assert"
)

func TestColorPicker_Update(t *testing.T) {
	tests := []struct {
		name         string
		key          tea.KeyMsg
		setupMonitor *tui.MonitorSpec
		expectedMsg  tea.Msg
	}{
		{
			name: "esc key closes color picker",
			key:  tea.KeyMsg{Type: tea.KeyEsc},
			setupMonitor: &tui.MonitorSpec{
				ID:          utils.IntPtr(1),
				Name:        "eDP-1",
				ColorPreset: tui.AutoColorPreset,
			},
			expectedMsg: tui.CloseColorPickerCommand{},
		},
		{
			name: "b key flips bitdepth",
			key:  tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'b'}},
			setupMonitor: &tui.MonitorSpec{
				ID:          utils.IntPtr(1),
				Name:        "eDP-1",
				ColorPreset: tui.AutoColorPreset,
				Bitdepth:    tui.DefaultBitdepth,
			},
			expectedMsg: tui.NextBitdepthCommand{MonitorID: 1},
		},
		{
			name: "up key triggers color preset preview through delegate",
			key:  tea.KeyMsg{Type: tea.KeyUp},
			setupMonitor: &tui.MonitorSpec{
				ID:          utils.IntPtr(1),
				Name:        "eDP-1",
				ColorPreset: tui.AutoColorPreset,
			},
			expectedMsg: tui.ChangeColorPresetCommand{Preset: tui.AutoColorPreset},
		},
		{
			name: "down key triggers color preset preview through delegate",
			key:  tea.KeyMsg{Type: tea.KeyDown},
			setupMonitor: &tui.MonitorSpec{
				ID:          utils.IntPtr(1),
				Name:        "eDP-1",
				ColorPreset: tui.AutoColorPreset,
			},
			expectedMsg: tui.ChangeColorPresetCommand{Preset: tui.SRGBColorPreset},
		},
		{
			name: "k key triggers color preset preview through delegate",
			key:  tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}},
			setupMonitor: &tui.MonitorSpec{
				ID:          utils.IntPtr(1),
				Name:        "eDP-1",
				ColorPreset: tui.AutoColorPreset,
			},
			expectedMsg: tui.ChangeColorPresetCommand{Preset: tui.AutoColorPreset},
		},
		{
			name: "j key triggers color preset preview through delegate",
			key:  tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}},
			setupMonitor: &tui.MonitorSpec{
				ID:          utils.IntPtr(1),
				Name:        "eDP-1",
				ColorPreset: tui.AutoColorPreset,
			},
			expectedMsg: tui.ChangeColorPresetCommand{Preset: tui.SRGBColorPreset},
		},
		{
			name: "enter key changes color preset through delegate",
			key:  tea.KeyMsg{Type: tea.KeyEnter},
			setupMonitor: &tui.MonitorSpec{
				ID:          utils.IntPtr(1),
				Name:        "eDP-1",
				ColorPreset: tui.AutoColorPreset,
			},
			expectedMsg: tui.ChangeColorPresetFinalCommand{Preset: tui.AutoColorPreset},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := testutils.NewTestConfig(t).Get()
			colors := tui.NewColorsManager(cfg)
			colorPicker := tui.NewColorPicker(colors)
			colorPicker.SetMonitor(tt.setupMonitor)

			cmd := colorPicker.Update(tt.key)

			if cmd == nil {
				t.Fatal("Expected command but got nil")
			}

			msg := cmd()
			assert.Equal(t, tt.expectedMsg, msg)
		})
	}
}

func TestColorPicker_SdrAdjustment(t *testing.T) {
	tests := []struct {
		name        string
		key         tea.KeyMsg
		preset      tui.ColorPreset
		initialVal  float64
		expectedMsg tea.Msg
	}{
		{
			name:       "r key increases SDR brightness for HDR preset",
			key:        tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}},
			preset:     tui.HDRColorPreset,
			initialVal: 1.0,
			expectedMsg: tui.AdjustSdrBrightnessCommand{
				MonitorID:     1,
				SdrBrightness: 1.01,
			},
		},
		{
			name:       "R key decreases SDR brightness for HDR preset",
			key:        tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'R'}},
			preset:     tui.HDRColorPreset,
			initialVal: 1.0,
			expectedMsg: tui.AdjustSdrBrightnessCommand{
				MonitorID:     1,
				SdrBrightness: 0.99,
			},
		},
		{
			name:       "t key increases SDR saturation for HDR preset",
			key:        tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'t'}},
			preset:     tui.HDRColorPreset,
			initialVal: 1.0,
			expectedMsg: tui.AdjustSdrSaturationCommand{
				MonitorID:  1,
				Saturation: 1.01,
			},
		},
		{
			name:       "T key decreases SDR saturation for HDR preset",
			key:        tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'T'}},
			preset:     tui.HDRColorPreset,
			initialVal: 1.0,
			expectedMsg: tui.AdjustSdrSaturationCommand{
				MonitorID:  1,
				Saturation: 0.99,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			monitor := &tui.MonitorSpec{
				ID:            utils.IntPtr(1),
				Name:          "eDP-1",
				ColorPreset:   tt.preset,
				SdrBrightness: tt.initialVal,
				SdrSaturation: tt.initialVal,
			}

			cfg := testutils.NewTestConfig(t).Get()
			colors := tui.NewColorsManager(cfg)
			colorPicker := tui.NewColorPicker(colors)
			colorPicker.SetMonitor(monitor)
			colorPicker.SetItems(monitor)

			cmd := colorPicker.Update(tt.key)
			assert.NotNil(t, cmd)

			msg := cmd()
			assert.Equal(t, tt.expectedMsg, msg)
		})
	}
}

func TestColorPicker_SetMonitor(t *testing.T) {
	monitor := &tui.MonitorSpec{
		ID:            utils.IntPtr(1),
		Name:          "eDP-1",
		ColorPreset:   tui.AutoColorPreset,
		SdrBrightness: 1.5,
		SdrSaturation: 1.2,
	}

	cfg := testutils.NewTestConfig(t).Get()
	colors := tui.NewColorsManager(cfg)
	colorPicker := tui.NewColorPicker(colors)
	colorPicker.SetMonitor(monitor)

	// Verify that the monitor was set by checking Update works
	cmd := colorPicker.Update(tea.KeyMsg{Type: tea.KeyEsc})
	assert.NotNil(t, cmd)
}

func TestColorPicker_Unset(t *testing.T) {
	monitor := &tui.MonitorSpec{
		ID:          utils.IntPtr(1),
		Name:        "eDP-1",
		ColorPreset: tui.AutoColorPreset,
	}

	cfg := testutils.NewTestConfig(t).Get()
	colors := tui.NewColorsManager(cfg)
	colorPicker := tui.NewColorPicker(colors)
	colorPicker.SetMonitor(monitor)

	cmd := colorPicker.Unset()
	assert.Nil(t, cmd)
}
