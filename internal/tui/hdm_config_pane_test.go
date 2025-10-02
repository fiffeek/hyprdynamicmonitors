package tui_test

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/fiffeek/hyprdynamicmonitors/internal/config"
	"github.com/fiffeek/hyprdynamicmonitors/internal/matchers"
	"github.com/fiffeek/hyprdynamicmonitors/internal/power"
	"github.com/fiffeek/hyprdynamicmonitors/internal/testutils"
	"github.com/fiffeek/hyprdynamicmonitors/internal/tui"
	"github.com/fiffeek/hyprdynamicmonitors/internal/utils"
	"github.com/stretchr/testify/assert"
)

func TestHDMConfigPane_Update(t *testing.T) {
	tests := []struct {
		name               string
		msg                tea.Msg
		initialPowerState  power.PowerState
		expectedPowerState power.PowerState
		expectProfileReset bool
	}{
		{
			name:               "ConfigReloaded resets profile",
			msg:                tui.ConfigReloaded{},
			initialPowerState:  power.ACPowerState,
			expectedPowerState: power.ACPowerState,
			expectProfileReset: true,
		},
		{
			name:               "PowerStateChanged updates state and resets profile",
			msg:                tui.PowerStateChangedCmd(power.BatteryPowerState),
			initialPowerState:  power.ACPowerState,
			expectedPowerState: power.BatteryPowerState,
			expectProfileReset: true,
		},
		{
			name:               "CreateNewProfileCommand does not reset profile",
			msg:                tui.CreateNewProfileCommand{},
			initialPowerState:  power.ACPowerState,
			expectedPowerState: power.ACPowerState,
			expectProfileReset: false,
		},
		{
			name:               "n key does not reset profile",
			msg:                tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}},
			initialPowerState:  power.ACPowerState,
			expectedPowerState: power.ACPowerState,
			expectProfileReset: false,
		},
		{
			name:               "a key does not reset profile",
			msg:                tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}},
			initialPowerState:  power.ACPowerState,
			expectedPowerState: power.ACPowerState,
			expectProfileReset: false,
		},
		{
			name:               "e key does not reset profile",
			msg:                tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}},
			initialPowerState:  power.ACPowerState,
			expectedPowerState: power.ACPowerState,
			expectProfileReset: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := testutils.NewTestConfig(t).
				WithProfiles(map[string]*config.Profile{
					"test": {
						Name: "test",
						Conditions: &config.ProfileCondition{
							PowerState: utils.JustPtr(config.AC),
							RequiredMonitors: []*config.RequiredMonitor{
								{Name: utils.StringPtr("eDP-1")},
							},
						},
					},
				}).Get()
			matcher := matchers.NewMatcher()
			monitors := []*tui.MonitorSpec{
				{Name: "eDP-1"},
			}

			pane := tui.NewHDMConfigPane(cfg, matcher, monitors, tt.initialPowerState)

			pane.Update(nil)

			oldProfile := pane.GetProfile()

			pane.Update(tt.msg)

			assert.Equal(t, tt.expectedPowerState, pane.GetPowerState())

			if tt.expectProfileReset {
				profile := pane.GetProfile()
				if profile != nil {
					assert.Equal(t, cfg.Get().Profiles["test"], profile)
				}
			} else {
				assert.Equal(t, oldProfile, pane.GetProfile())
			}
		})
	}
}
