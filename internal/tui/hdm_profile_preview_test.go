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
	"github.com/stretchr/testify/require"
)

func TestHDMProfilePreview_Update(t *testing.T) {
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
			msg:                tui.PowerStateChangedCmd(power.ACPowerState),
			initialPowerState:  power.BatteryPowerState,
			expectedPowerState: power.ACPowerState,
			expectProfileReset: true,
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
			cfgProfile := cfg.Get().Profiles["test"]
			content := "hello"
			require.NoError(t, utils.WriteAtomic(cfgProfile.ConfigFile, []byte(content)))
			matcher := matchers.NewMatcher()
			monitors := []*tui.MonitorSpec{
				{Name: "eDP-1"},
			}

			preview := tui.NewHDMProfilePreview(cfg, matcher, monitors, tt.initialPowerState)
			oldProfile := preview.GetProfile()

			preview.Update(tt.msg)

			assert.Equal(t, tt.expectedPowerState, preview.GetPowerState())

			if tt.expectProfileReset {
				assert.Equal(t, cfgProfile, preview.GetProfile())
				assert.Equal(t, content, preview.GetText())
			} else {
				assert.Equal(t, oldProfile, preview.GetProfile())
			}
		})
	}
}
