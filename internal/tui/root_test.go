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
	"github.com/fiffeek/hyprdynamicmonitors/internal/tui"
	"github.com/fiffeek/hyprdynamicmonitors/internal/utils"
	"github.com/stretchr/testify/require"
)

type step struct {
	msg                   tea.Msg
	waitFor               *time.Duration
	expectOutputToContain string
}

var (
	footer             = "p pan (move freely on the grid) • F fullscreen the preview"
	defaultWait        = 200 * time.Millisecond
	defaultMonitorData = "four.json"
)

func TestModel_Update_UserFlows(t *testing.T) {
	tests := []struct {
		name         string
		cfg          *config.Config
		monitorsData string
		powerState   power.PowerState
		steps        []step
		runFor       *time.Duration
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hyprMonitors := loadMonitorsFromTestdata(t, tt.monitorsData)
			pm := profilemaker.NewService(tt.cfg, nil)
			model := tui.NewModel(tt.cfg,
				hyprMonitors, pm, "test-version", tt.powerState, tt.runFor)
			tm := teatest.NewTestModel(t, model, teatest.WithInitialTermSize(160, 45))

			// wait for app to be `ready`, just check if the footer is up
			teatest.WaitFor(t, tm.Output(), func(bts []byte) bool {
				return bytes.Contains(bts, []byte(footer))
			}, teatest.WithCheckInterval(time.Millisecond*50),
				teatest.WithDuration(time.Millisecond*200))

			for _, step := range tt.steps {
				tm.Send(step.msg)
				if step.expectOutputToContain != "" {
					stepWait := defaultWait
					if step.waitFor != nil {
						stepWait = *step.waitFor
					}
					teatest.WaitFor(t, tm.Output(), func(bts []byte) bool {
						return bytes.Contains(bts, []byte(step.expectOutputToContain))
					}, teatest.WithCheckInterval(time.Millisecond*50), teatest.WithDuration(stepWait))
				}
			}
			tm.Send(tea.Quit())
			tm.WaitFinished(t, teatest.WithFinalTimeout(*tt.runFor))

			fm := tm.FinalModel(t)
			m, ok := fm.(tui.Model)
			require.True(t, ok, "the model should be of the same type")

			teatest.RequireEqualOutput(t, []byte(m.View()))
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
