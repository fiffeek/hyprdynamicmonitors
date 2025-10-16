package tui_test

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/exp/teatest"
	"github.com/fiffeek/hyprdynamicmonitors/internal/config"
	"github.com/fiffeek/hyprdynamicmonitors/internal/hypr"
	"github.com/fiffeek/hyprdynamicmonitors/internal/power"
	"github.com/fiffeek/hyprdynamicmonitors/internal/profilemaker"
	"github.com/fiffeek/hyprdynamicmonitors/internal/testutils"
	"github.com/fiffeek/hyprdynamicmonitors/internal/tui"
	"github.com/fiffeek/hyprdynamicmonitors/internal/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type step struct {
	msg                   tea.Msg
	waitFor               *time.Duration
	expectOutputToContain string
	times                 *int
	sleepAfter            *time.Duration
	validateSideEffects   func(*config.Config)
}

var (
	footer             = "p pan (move freely on the grid) • F fullscreen the preview"
	defaultWait        = 200 * time.Millisecond
	defaultMonitorData = "four.json"
	headless           = "headless.json"
	twoMonitorsData    = "two.json"
)

func TestModel_Update_UserFlows(t *testing.T) {
	tests := []struct {
		name                string
		cfg                 *config.Config
		monitorsData        string
		powerState          power.PowerState
		lidState            power.LidState
		steps               []step
		runFor              *time.Duration
		validateSideEffects func(*config.Config, tui.Model)
	}{
		{
			name:         "rotate",
			monitorsData: defaultMonitorData,
			runFor:       utils.JustPtr(500 * time.Millisecond),
			steps: []step{
				{
					msg:                   tea.KeyMsg{Type: tea.KeyEnter},
					expectOutputToContain: "EDITING",
				},
				{
					msg:                   tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}},
					expectOutputToContain: "eDP-1→",
				},
			},
		},

		{
			name:         "scale_open",
			monitorsData: defaultMonitorData,
			runFor:       utils.JustPtr(500 * time.Millisecond),
			steps: []step{
				{
					msg:                   tea.KeyMsg{Type: tea.KeyEnter},
					expectOutputToContain: "EDITING",
				},
				{
					msg:                   tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}},
					expectOutputToContain: "Scale: 2.00",
				},
			},
		},

		{
			name:         "color_open",
			monitorsData: defaultMonitorData,
			runFor:       utils.JustPtr(500 * time.Millisecond),
			steps: []step{
				{
					msg:                   tea.KeyMsg{Type: tea.KeyEnter},
					expectOutputToContain: "EDITING",
				},
				{
					msg:                   tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'C'}},
					expectOutputToContain: "Adjust Colors",
				},
			},
		},

		{
			name:         "color_srgb",
			monitorsData: defaultMonitorData,
			runFor:       utils.JustPtr(500 * time.Millisecond),
			steps: []step{
				{
					msg:                   tea.KeyMsg{Type: tea.KeyEnter},
					expectOutputToContain: "EDITING",
				},
				{
					msg:                   tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'C'}},
					expectOutputToContain: "Adjust Colors",
				},
				{
					msg:                   tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}},
					expectOutputToContain: "► srgb",
				},
			},
		},

		{
			name:         "color_hdr",
			monitorsData: defaultMonitorData,
			runFor:       utils.JustPtr(500 * time.Millisecond),
			steps: []step{
				{
					msg:                   tea.KeyMsg{Type: tea.KeyEnter},
					expectOutputToContain: "EDITING",
				},
				{
					msg:                   tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'C'}},
					expectOutputToContain: "Adjust Colors",
				},
				{
					msg:                   tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}},
					times:                 utils.JustPtr(4),
					expectOutputToContain: "► hdr",
				},
				{
					msg:                   tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}},
					times:                 utils.JustPtr(1),
					expectOutputToContain: "SDR Brightness: 1.01",
				},
				{
					msg:                   tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'t'}},
					times:                 utils.JustPtr(1),
					expectOutputToContain: "SDR Saturation: 1.01",
				},
				{
					msg:                   tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'H'}},
					expectOutputToContain: "monitor = desc:BOE",
				},
				{
					msg:                   tea.KeyMsg{Type: tea.KeyEnter},
					expectOutputToContain: "EDITING",
				},
			},
		},

		{
			name:         "color_bitdepth",
			monitorsData: defaultMonitorData,
			runFor:       utils.JustPtr(500 * time.Millisecond),
			steps: []step{
				{
					msg:                   tea.KeyMsg{Type: tea.KeyEnter},
					expectOutputToContain: "EDITING",
				},
				{
					msg:                   tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'C'}},
					expectOutputToContain: "Adjust Colors",
				},
				{
					msg:                   tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'b'}},
					expectOutputToContain: "Bitdepth: 10",
				},
			},
		},

		{
			name:         "scale",
			monitorsData: defaultMonitorData,
			runFor:       utils.JustPtr(500 * time.Millisecond),
			steps: []step{
				{
					msg:                   tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}},
					expectOutputToContain: "► DP-1",
				},
				{
					msg:                   tea.KeyMsg{Type: tea.KeyEnter},
					expectOutputToContain: "EDITING",
				},
				{
					msg:                   tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}},
					expectOutputToContain: "Scale: 1.25",
				},
				{
					msg:                   tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}},
					expectOutputToContain: "Scale: 1.26",
				},
				{
					msg:                   tea.KeyMsg{Type: tea.KeyEnter},
					expectOutputToContain: "EDITING",
				},
			},
		},

		{
			name:         "mode_open",
			monitorsData: defaultMonitorData,
			runFor:       utils.JustPtr(500 * time.Millisecond),
			steps: []step{
				{
					msg:                   tea.KeyMsg{Type: tea.KeyEnter},
					expectOutputToContain: "EDITING",
				},
				{
					msg:                   tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'m'}},
					expectOutputToContain: "Select monitor mode",
				},
			},
		},

		{
			name:         "mode_select",
			monitorsData: defaultMonitorData,
			runFor:       utils.JustPtr(500 * time.Millisecond),
			steps: []step{
				{
					msg:                   tea.KeyMsg{Type: tea.KeyEnter},
					expectOutputToContain: "EDITING",
				},
				{
					msg:                   tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'m'}},
					expectOutputToContain: "Select monitor mode",
				},
				{
					msg:                   tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}},
					expectOutputToContain: "► 2880x1920@60.00Hz",
				},
				{
					msg:                   tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}},
					expectOutputToContain: "► 1920x1200@120.00Hz",
				},
				{
					msg:                   tea.KeyMsg{Type: tea.KeyEnter},
					expectOutputToContain: "► eDP-1 (BOE NE135A1M-NY...) [EDITING]",
				},
			},
		},

		{
			name:         "mirrors_open",
			monitorsData: defaultMonitorData,
			runFor:       utils.JustPtr(500 * time.Millisecond),
			steps: []step{
				{
					msg:                   tea.KeyMsg{Type: tea.KeyEnter},
					expectOutputToContain: "EDITING",
				},
				{
					msg:                   tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'i'}},
					expectOutputToContain: "Select a monitor mirror",
				},
			},
		},

		{
			name:         "mirror_select",
			monitorsData: defaultMonitorData,
			runFor:       utils.JustPtr(500 * time.Millisecond),
			steps: []step{
				{
					msg:                   tea.KeyMsg{Type: tea.KeyEnter},
					expectOutputToContain: "EDITING",
				},
				{
					msg:                   tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'i'}},
					expectOutputToContain: "Select a monitor mirror",
				},
				{
					msg:                   tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}},
					expectOutputToContain: "► DP-1",
				},
				{
					msg:                   tea.KeyMsg{Type: tea.KeyEnter},
					expectOutputToContain: "► eDP-1 (BOE NE135A1M-NY...) [EDITING]",
				},
			},
		},

		{
			name:         "monitors_picking",
			monitorsData: defaultMonitorData,
			runFor:       utils.JustPtr(500 * time.Millisecond),
			steps: []step{
				{
					msg:                   tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}},
					expectOutputToContain: "► DP-1 (Dell Inc. DELL ...)",
				},
				{
					msg:                   tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}},
					expectOutputToContain: "► DP-2 (Samsung Electri...)",
				},
				{
					msg:                   tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}},
					expectOutputToContain: "► HEADLESS-1 (Headless Virtua...)",
				},
			},
		},

		{
			name:         "monitors_select",
			monitorsData: defaultMonitorData,
			runFor:       utils.JustPtr(500 * time.Millisecond),
			steps: []step{
				{
					msg:                   tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}},
					expectOutputToContain: "► DP-1 (Dell Inc. DELL ...)",
				},
				{
					msg:                   tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}},
					expectOutputToContain: "► DP-2 (Samsung Electri...)",
				},
				{
					msg:                   tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}},
					expectOutputToContain: "► HEADLESS-1 (Headless Virtua...)",
				},
				{
					msg:                   tea.KeyMsg{Type: tea.KeyEnter},
					expectOutputToContain: "EDITING",
				},
			},
		},

		{
			name:         "snap_center",
			monitorsData: defaultMonitorData,
			runFor:       utils.JustPtr(500 * time.Millisecond),
			steps: []step{
				{
					msg:                   tea.KeyMsg{Type: tea.KeyEnter},
					expectOutputToContain: "EDITING",
				},
				{
					msg:   tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'h'}},
					times: utils.IntPtr(6),
				},
				{
					msg:                   tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}},
					times:                 utils.IntPtr(12),
					expectOutputToContain: "Position: 1536,384",
					sleepAfter:            utils.JustPtr(100 * time.Millisecond),
				},
			},
		},

		{
			name:         "snap",
			monitorsData: defaultMonitorData,
			runFor:       utils.JustPtr(500 * time.Millisecond),
			steps: []step{
				{
					msg:                   tea.KeyMsg{Type: tea.KeyEnter},
					expectOutputToContain: "EDITING",
				},
				{
					msg:   tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'h'}},
					times: utils.IntPtr(5),
				},
				{
					msg:                   tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}},
					times:                 utils.IntPtr(12),
					expectOutputToContain: "Position: 1632,1728",
					sleepAfter:            utils.JustPtr(100 * time.Millisecond),
				},
			},
		},

		{
			name:         "follow_monitor",
			monitorsData: defaultMonitorData,
			runFor:       utils.JustPtr(500 * time.Millisecond),
			steps: []step{
				{
					msg:                   tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'o'}},
					expectOutputToContain: "Follow ON",
				},
				{
					msg:                   tea.KeyMsg{Type: tea.KeyEnter},
					expectOutputToContain: "EDITING",
				},
				{
					msg:                   tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}},
					times:                 utils.IntPtr(9),
					expectOutputToContain: "Position: 1920,540",
					sleepAfter:            utils.JustPtr(100 * time.Millisecond),
				},
			},
		},

		{
			name:         "pan",
			monitorsData: defaultMonitorData,
			runFor:       utils.JustPtr(500 * time.Millisecond),
			steps: []step{
				{
					msg:                   tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'p'}},
					expectOutputToContain: "Panning",
				},
				{
					msg:   tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'l'}},
					times: utils.IntPtr(38),
				},
				{
					msg:                   tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}},
					times:                 utils.IntPtr(10),
					expectOutputToContain: "Center: (3800,1000)",
				},
			},
		},

		{
			name:         "fullscreen",
			monitorsData: defaultMonitorData,
			runFor:       utils.JustPtr(500 * time.Millisecond),
			steps: []step{
				{
					msg:                   tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'F'}},
					expectOutputToContain: "Fullscreen",
				},
			},
		},

		{
			name:         "move_fullscreen",
			monitorsData: defaultMonitorData,
			runFor:       utils.JustPtr(800 * time.Millisecond),
			steps: []step{
				{
					msg:                   tea.KeyMsg{Type: tea.KeyEnter},
					expectOutputToContain: "EDITING",
				},
				{
					msg:                   tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'F'}},
					expectOutputToContain: "Fullscreen",
				},
				{
					msg:   tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}},
					times: utils.IntPtr(5),
				},
				{
					msg:        tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'h'}},
					times:      utils.IntPtr(5),
					sleepAfter: utils.JustPtr(100 * time.Millisecond),
				},
			},
		},

		{
			name:         "disable_monitor",
			monitorsData: defaultMonitorData,
			runFor:       utils.JustPtr(500 * time.Millisecond),
			steps: []step{
				{
					msg:                   tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}},
					times:                 utils.IntPtr(2),
					expectOutputToContain: "► DP-2 (Samsung Electri...)",
				},
				{
					msg:                   tea.KeyMsg{Type: tea.KeyEnter},
					expectOutputToContain: "EDITING",
				},
				{
					msg:                   tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}},
					expectOutputToContain: "monitor = desc:Samsung Electric Company C27F390 HTHK500315,disable",
				},
			},
		},

		{
			name:         "rotate_disabled",
			monitorsData: defaultMonitorData,
			runFor:       utils.JustPtr(500 * time.Millisecond),
			steps: []step{
				{
					msg:                   tea.KeyMsg{Type: tea.KeyEnter},
					expectOutputToContain: "EDITING",
				},
				{
					msg:                   tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}},
					expectOutputToContain: "monitor = desc:BOE NE135A1M-NY1,disable",
				},
				{
					msg:                   tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}},
					expectOutputToContain: "Rotate Apply: monitor is disabled",
				},
			},
		},

		{
			name:         "zoom_in",
			monitorsData: defaultMonitorData,
			runFor:       utils.JustPtr(500 * time.Millisecond),
			steps: []step{
				{
					msg:                   tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'+'}},
					expectOutputToContain: "Virtual Area: 4000x4000",
				},
			},
		},

		{
			name:         "zoom_out",
			monitorsData: defaultMonitorData,
			runFor:       utils.JustPtr(500 * time.Millisecond),
			steps: []step{
				{
					msg:                   tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'-'}},
					expectOutputToContain: "Virtual Area: 16000x16000",
				},
			},
		},

		{
			name:         "pan_back",
			monitorsData: defaultMonitorData,
			runFor:       utils.JustPtr(500 * time.Millisecond),
			steps: []step{
				{
					msg:                   tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'p'}},
					expectOutputToContain: "Panning",
				},
				{
					msg:   tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'l'}},
					times: utils.IntPtr(10),
				},
				{
					msg:                   tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}},
					times:                 utils.IntPtr(10),
					expectOutputToContain: "Center: (1000,1000)",
				},
				{
					msg:                   tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}},
					expectOutputToContain: "Virtual Area: 8000x8000 | Snapping",
				},
				{
					msg:        tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'p'}},
					sleepAfter: utils.JustPtr(50 * time.Millisecond),
				},
			},
		},

		{
			name:         "new_profile_name_open",
			monitorsData: defaultMonitorData,
			runFor:       utils.JustPtr(500 * time.Millisecond),
			cfg:          testutils.NewTestConfig(t).Get(),
			steps: []step{
				{
					msg:                   tea.KeyMsg{Type: tea.KeyTab},
					expectOutputToContain: "No Matching Profile",
				},
				{
					msg:                   tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}},
					expectOutputToContain: "Type the profile name",
				},
				{
					msg:                   tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'h'}},
					expectOutputToContain: "h",
				},
				{
					msg:                   tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}},
					expectOutputToContain: "he",
				},
				{
					msg:                   tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}},
					expectOutputToContain: "hey",
				},
			},
		},

		{
			name:         "headless",
			monitorsData: headless,
			runFor:       utils.JustPtr(800 * time.Millisecond),
			cfg:          testutils.NewTestConfig(t).Get(),
			steps: []step{
				{
					msg:                   tea.KeyMsg{Type: tea.KeyEnter},
					expectOutputToContain: "[EDITING]",
				},
				{
					msg:                   tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'m'}},
					expectOutputToContain: "Select monitor mode",
				},
				{
					msg:                   tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}},
					expectOutputToContain: "► 1920x1080@0.06Hz",
				},
				{
					msg:                   tea.KeyMsg{Type: tea.KeyEnter},
					expectOutputToContain: "[EDITING]",
				},
				{
					msg:                   tea.KeyMsg{Type: tea.KeyTab},
					expectOutputToContain: "No Matching Profile",
				},
				{
					msg:                   tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}},
					expectOutputToContain: "Type the profile name",
				},
				{
					msg:                   tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'h'}},
					expectOutputToContain: "h",
				},
				{
					msg:        tea.KeyMsg{Type: tea.KeyEnter},
					sleepAfter: utils.JustPtr(100 * time.Millisecond),
					// mimic reload in the outer process
					validateSideEffects: func(cfg *config.Config) {
						require.NoError(t, cfg.Reload())
						raw := cfg.Get()
						assert.Equal(t, []*config.RequiredMonitor{
							{Name: utils.JustPtr("HEADLESS-2"), MonitorTag: utils.JustPtr("monitor1")},
						}, raw.Profiles["h"].Conditions.RequiredMonitors)
					},
				},
				// mimic config reload send event
				{
					msg:        tui.ConfigReloaded{},
					sleepAfter: utils.JustPtr(200 * time.Millisecond),
				},
			},
		},

		{
			name:         "new_profile",
			monitorsData: defaultMonitorData,
			runFor:       utils.JustPtr(800 * time.Millisecond),
			cfg:          testutils.NewTestConfig(t).Get(),
			steps: []step{
				{
					msg:                   tea.KeyMsg{Type: tea.KeyTab},
					expectOutputToContain: "No Matching Profile",
				},
				{
					msg:                   tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}},
					expectOutputToContain: "Type the profile name",
				},
				{
					msg:                   tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'h'}},
					expectOutputToContain: "h",
				},
				{
					msg:        tea.KeyMsg{Type: tea.KeyEnter},
					sleepAfter: utils.JustPtr(100 * time.Millisecond),
					// mimic reload in the outer process
					validateSideEffects: func(cfg *config.Config) {
						require.NoError(t, cfg.Reload())
						raw := cfg.Get()
						assert.Equal(t, []*config.RequiredMonitor{
							{Description: utils.JustPtr("BOE NE135A1M-NY1"), MonitorTag: utils.JustPtr("monitor0")},
							{Description: utils.JustPtr("Dell Inc. DELL U2723QE 5YNK3H3"), MonitorTag: utils.JustPtr("monitor1")},
							{Description: utils.JustPtr("Samsung Electric Company C27F390 HTHK500315"), MonitorTag: utils.JustPtr("monitor2")},
							{Description: utils.JustPtr("Headless Virtual Display"), MonitorTag: utils.JustPtr("monitor3")},
						}, raw.Profiles["h"].Conditions.RequiredMonitors)
					},
				},
				// mimic config reload send event
				{
					msg:        tui.ConfigReloaded{},
					sleepAfter: utils.JustPtr(200 * time.Millisecond),
				},
			},
		},

		{
			name:         "matching_profile_view",
			monitorsData: twoMonitorsData,
			runFor:       utils.JustPtr(500 * time.Millisecond),
			cfg: testutils.NewTestConfig(t).WithProfiles(map[string]*config.Profile{
				"two": {
					ConfigType: utils.JustPtr(config.Template),
					Conditions: &config.ProfileCondition{
						RequiredMonitors: []*config.RequiredMonitor{
							{
								Description: utils.StringPtr("BOE NE135A1M-NY1"),
							},
							{
								Description: utils.StringPtr("Dell Inc. DELL U2723QE 5YNK3H3"),
							},
						},
					},
				},
			}).Get(),
			steps: []step{
				{
					msg:                   tea.KeyMsg{Type: tea.KeyTab},
					expectOutputToContain: "Profile: two",
				},
			},
		},

		{
			name:         "maybe_matching_profile_view",
			monitorsData: twoMonitorsData,
			runFor:       utils.JustPtr(500 * time.Millisecond),
			cfg: testutils.NewTestConfig(t).WithProfiles(map[string]*config.Profile{
				"one": {
					ConfigType: utils.JustPtr(config.Template),
					Conditions: &config.ProfileCondition{
						RequiredMonitors: []*config.RequiredMonitor{
							{
								Description: utils.StringPtr("BOE NE135A1M-NY1"),
							},
						},
					},
				},
			}).Get(),
			steps: []step{
				{
					msg:                   tea.KeyMsg{Type: tea.KeyTab},
					expectOutputToContain: "Monitor Count Mismatch",
				},
			},
		},

		{
			name:         "no_matching_profile_view",
			monitorsData: twoMonitorsData,
			runFor:       utils.JustPtr(500 * time.Millisecond),
			cfg: testutils.NewTestConfig(t).WithProfiles(map[string]*config.Profile{
				"whatever": {
					ConfigType: utils.JustPtr(config.Template),
					Conditions: &config.ProfileCondition{
						RequiredMonitors: []*config.RequiredMonitor{
							{
								Description: utils.StringPtr("BOE NE135A1M-NY1 SAMOYEDS"),
							},
						},
					},
				},
			}).Get(),
			steps: []step{
				{
					msg:                   tea.KeyMsg{Type: tea.KeyTab},
					expectOutputToContain: "No configuration profile",
				},
			},
		},

		{
			name:         "no_matching_ne_profile_view",
			monitorsData: twoMonitorsData,
			runFor:       utils.JustPtr(500 * time.Millisecond),
			cfg: testutils.NewTestConfig(t).WithProfiles(map[string]*config.Profile{
				"whatever": {
					ConfigType: utils.JustPtr(config.Template),
					Conditions: &config.ProfileCondition{
						RequiredMonitors: []*config.RequiredMonitor{
							{
								Description: utils.StringPtr("BOE NE135A1M-NY1 SAMOYEDS"),
							},
						},
					},
				},
			}).Get(),
			steps: []step{
				{
					msg:                   tea.KeyMsg{Type: tea.KeyTab},
					expectOutputToContain: "No configuration profile",
				},
				{
					msg:                   tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}},
					expectOutputToContain: "Type the profile name",
				},
			},
		},

		{
			name:         "matching_profile_append",
			monitorsData: twoMonitorsData,
			runFor:       utils.JustPtr(800 * time.Millisecond),
			cfg: testutils.NewTestConfig(t).WithProfiles(map[string]*config.Profile{
				"two": {
					ConfigType: utils.JustPtr(config.Template),
					Conditions: &config.ProfileCondition{
						RequiredMonitors: []*config.RequiredMonitor{
							{
								Description: utils.StringPtr("BOE NE135A1M-NY1"),
							},
							{
								Description: utils.StringPtr("Dell Inc. DELL U2723QE 5YNK3H3"),
							},
						},
					},
				},
			}).Get(),
			steps: []step{
				{
					msg:                   tea.KeyMsg{Type: tea.KeyTab},
					expectOutputToContain: "Profile: two",
				},
				{
					msg:                   tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}},
					expectOutputToContain: "Apply edited settings to two profile?",
				},
				{
					msg:        tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'Y'}},
					sleepAfter: utils.JustPtr(100 * time.Millisecond),
					// mimic reload in the outer process
					validateSideEffects: func(cfg *config.Config) {
						require.NoError(t, cfg.Reload())
					},
				},
				// mimic event sent
				{
					msg:                   tui.ConfigReloaded{},
					expectOutputToContain: "TUI AUTO START",
				},
				// go back to editing
				{
					msg:                   tea.KeyMsg{Type: tea.KeyTab},
					expectOutputToContain: "► eDP-1 (BOE NE135A1M-NY...)",
				},
				{
					msg:                   tea.KeyMsg{Type: tea.KeyEnter},
					expectOutputToContain: "► eDP-1 (BOE NE135A1M-NY...) [EDITING]",
				},
				// move monitor right
				{
					msg:                   tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'l'}},
					times:                 utils.IntPtr(2),
					expectOutputToContain: "Position: 2020,1080",
				},
				// finish editing
				{
					msg:                   tea.KeyMsg{Type: tea.KeyEnter},
					expectOutputToContain: "► eDP-1 (BOE NE135A1M-NY...)",
				},
				// change back to the profiles view
				{
					msg:                   tea.KeyMsg{Type: tea.KeyTab},
					expectOutputToContain: "Profile: two",
				},
				// append again
				{
					msg:                   tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}},
					expectOutputToContain: "Apply edited settings to two profile?",
				},
				{
					msg:        tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'Y'}},
					sleepAfter: utils.JustPtr(100 * time.Millisecond),
					// mimic reload in the outer process
					validateSideEffects: func(cfg *config.Config) {
						require.NoError(t, cfg.Reload())
					},
				},
				// mimic event sent
				{
					msg:                   tui.ConfigReloaded{},
					expectOutputToContain: "2020x1080",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hyprMonitors := loadMonitorsFromTestdata(t, tt.monitorsData)
			pm := profilemaker.NewService(tt.cfg, nil)
			model := tui.NewModel(tt.cfg,
				hyprMonitors, pm, "test-version", tt.powerState, tt.runFor, true, tt.lidState)
			tm := teatest.NewTestModel(t, model, teatest.WithInitialTermSize(160, 45))

			// wait for app to be `ready`, just check if the footer is up
			teatest.WaitFor(t, tm.Output(), func(bts []byte) bool {
				return bytes.Contains(bts, []byte(footer))
			}, teatest.WithCheckInterval(time.Millisecond*50),
				teatest.WithDuration(time.Millisecond*200))

			for _, step := range tt.steps {
				if step.times == nil {
					step.times = utils.IntPtr(1)
				}
				for range *step.times {
					tm.Send(step.msg)
				}
				if step.expectOutputToContain != "" {
					stepWait := defaultWait
					if step.waitFor != nil {
						stepWait = *step.waitFor
					}
					teatest.WaitFor(t, tm.Output(), func(bts []byte) bool {
						return bytes.Contains(bts, []byte(step.expectOutputToContain))
					}, teatest.WithCheckInterval(time.Millisecond*50), teatest.WithDuration(stepWait))
				}
				if step.sleepAfter != nil {
					time.Sleep(*step.sleepAfter)
				}

				if step.validateSideEffects != nil {
					step.validateSideEffects(tt.cfg)
				}
			}
			tm.Send(tea.Quit())
			tm.WaitFinished(t, teatest.WithFinalTimeout(*tt.runFor))

			fm := tm.FinalModel(t)
			m, ok := fm.(tui.Model)
			require.True(t, ok, "the model should be of the same type")

			teatest.RequireEqualOutput(t, []byte(m.View()))

			if tt.validateSideEffects != nil {
				tt.validateSideEffects(tt.cfg, m)
			}
		})
	}
}

func loadMonitorsFromTestdata(t *testing.T, filename string) hypr.MonitorSpecs {
	t.Helper()
	path := filepath.Join("testdata", filename)
	// nolint:gosec
	data, err := os.ReadFile(path)
	require.NoError(t, err, "failed to read testdata file: %s", filename)

	var specs hypr.MonitorSpecs
	err = json.Unmarshal(data, &specs)
	require.NoError(t, err, "failed to unmarshal monitor specs from: %s", filename)

	return specs
}
