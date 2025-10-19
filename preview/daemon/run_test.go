package daemon_test

import (
	"context"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/fiffeek/hyprdynamicmonitors/internal/config"
	"github.com/fiffeek/hyprdynamicmonitors/internal/hypr"
	"github.com/fiffeek/hyprdynamicmonitors/internal/power"
	"github.com/fiffeek/hyprdynamicmonitors/internal/testutils"
	"github.com/fiffeek/hyprdynamicmonitors/internal/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test__Preview(t *testing.T) {
	tests := []struct {
		name           string
		config         *testutils.TestConfig
		tapeName       string
		contextTimeout time.Duration

		hyprMonitorResponseFiles []string

		hyprEvents             []string
		initialHyprEventsSleep *time.Duration
		sleepBetweenHyprEvents time.Duration

		disablePowerEvents      bool
		powerEvents             []power.PowerState
		initialPowerEventsSleep time.Duration
		sleepBetweenPowerEvents time.Duration

		connectToSessionBus bool

		enableLidEvents       bool
		lidEvents             []power.LidState
		initialLidEventsSleep time.Duration
		sleepBetweenLidEvents time.Duration

		tui bool
	}{
		{
			name: "preview lid tui",
			config: testutils.NewTestConfig(t).WithNotifications(&config.Notifications{
				Disabled: utils.JustPtr(true),
			}).WithProfiles(
				map[string]*config.Profile{
					"lid_closed": {
						ConfigType: utils.JustPtr(config.Template),
						Conditions: &config.ProfileCondition{
							LidState: utils.JustPtr(config.ClosedLidStateType),
							RequiredMonitors: []*config.RequiredMonitor{
								{
									Name:       utils.StringPtr("eDP-1"),
									MonitorTag: utils.JustPtr("laptop"),
								},
								{
									Name:       utils.StringPtr("DP-11"),
									MonitorTag: utils.JustPtr("external"),
								},
							},
						},
					},
					"lid_opened": {
						ConfigType: utils.JustPtr(config.Template),
						Conditions: &config.ProfileCondition{
							LidState: utils.JustPtr(config.OpenedLidStateType),
							RequiredMonitors: []*config.RequiredMonitor{
								{
									Name:       utils.StringPtr("eDP-1"),
									MonitorTag: utils.JustPtr("laptop"),
								},
								{
									Name:       utils.StringPtr("DP-11"),
									MonitorTag: utils.JustPtr("external"),
								},
							},
						},
					},
				},
			).FillProfileConfigFile("lid_opened", "testdata/templates/lid_opened.go.tmpl").
				FillProfileConfigFile("lid_closed", "testdata/templates/lid_closed.go.tmpl"),
			tapeName:       "lid_tui",
			contextTimeout: 60 * time.Second,

			hyprMonitorResponseFiles: []string{
				"testdata/hypr/basic.json",
			},

			disablePowerEvents: true,

			connectToSessionBus:   true,
			enableLidEvents:       true,
			lidEvents:             []power.LidState{power.ClosedLidState},
			initialLidEventsSleep: 5 * time.Second,
			sleepBetweenLidEvents: 100 * time.Millisecond,

			tui: true,
		},

		{
			name: "preview lid events",
			config: testutils.NewTestConfig(t).WithNotifications(&config.Notifications{
				Disabled: utils.JustPtr(true),
			}).WithProfiles(
				map[string]*config.Profile{
					"lid_closed": {
						ConfigType: utils.JustPtr(config.Template),
						Conditions: &config.ProfileCondition{
							LidState: utils.JustPtr(config.ClosedLidStateType),
							RequiredMonitors: []*config.RequiredMonitor{
								{
									Name:       utils.StringPtr("eDP-1"),
									MonitorTag: utils.JustPtr("laptop"),
								},
								{
									Name:       utils.StringPtr("DP-11"),
									MonitorTag: utils.JustPtr("external"),
								},
							},
						},
					},
					"lid_opened": {
						ConfigType: utils.JustPtr(config.Template),
						Conditions: &config.ProfileCondition{
							LidState: utils.JustPtr(config.OpenedLidStateType),
							RequiredMonitors: []*config.RequiredMonitor{
								{
									Name:       utils.StringPtr("eDP-1"),
									MonitorTag: utils.JustPtr("laptop"),
								},
								{
									Name:       utils.StringPtr("DP-11"),
									MonitorTag: utils.JustPtr("external"),
								},
							},
						},
					},
				},
			).FillProfileConfigFile("lid_opened", "testdata/templates/lid_state.go.tmpl").
				FillProfileConfigFile("lid_closed", "testdata/templates/lid_state.go.tmpl"),
			tapeName:       "lid_events",
			contextTimeout: 60 * time.Second,

			hyprMonitorResponseFiles: []string{
				"testdata/hypr/basic.json",
			},

			disablePowerEvents: true,

			connectToSessionBus:   true,
			enableLidEvents:       true,
			lidEvents:             []power.LidState{power.ClosedLidState},
			initialLidEventsSleep: 4 * time.Second,
			sleepBetweenLidEvents: 100 * time.Millisecond,
		},

		{
			name: "preview power events",
			config: testutils.NewTestConfig(t).WithNotifications(&config.Notifications{
				Disabled: utils.JustPtr(true),
			}).WithProfiles(
				map[string]*config.Profile{
					"ac": {
						ConfigType: utils.JustPtr(config.Template),
						Conditions: &config.ProfileCondition{
							PowerState: utils.JustPtr(config.AC),
							RequiredMonitors: []*config.RequiredMonitor{
								{
									Name:       utils.StringPtr("eDP-1"),
									MonitorTag: utils.JustPtr("laptop"),
								},
							},
						},
					},
					"bat": {
						ConfigType: utils.JustPtr(config.Template),
						Conditions: &config.ProfileCondition{
							PowerState: utils.JustPtr(config.BAT),
							RequiredMonitors: []*config.RequiredMonitor{
								{
									Name:       utils.StringPtr("eDP-1"),
									MonitorTag: utils.JustPtr("laptop"),
								},
							},
						},
					},
				},
			).FillProfileConfigFile("ac", "testdata/templates/power_state.go.tmpl").
				FillProfileConfigFile("bat", "testdata/templates/power_state.go.tmpl"),
			tapeName:       "power_events",
			contextTimeout: 60 * time.Second,

			hyprMonitorResponseFiles: []string{
				"testdata/hypr/basic.json",
			},

			disablePowerEvents:      false,
			connectToSessionBus:     true,
			powerEvents:             []power.PowerState{power.ACPowerState},
			initialPowerEventsSleep: 4 * time.Second,
			sleepBetweenPowerEvents: 100 * time.Millisecond,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			binaryStartingChan := make(chan struct{})
			ctx, cancel := context.WithTimeout(context.Background(), tt.contextTimeout)
			defer cancel()

			// hypr: vars
			xdgRuntimeDir, signature := testutils.SetupHyprEnvVars(t)

			// hypr: fake ipc server
			listener, teardown := testutils.SetupHyprSocket(ctx, t, xdgRuntimeDir, signature, hypr.GetHyprSocket)
			defer teardown()
			responses := [][]byte{}
			expectedCommands := []string{}
			for _, file := range tt.hyprMonitorResponseFiles {
				// nolint:gosec
				responseData, err := os.ReadFile(file)
				require.NoError(t, err, "Failed to read test response file %s: %w", file, err)
				responses = append(responses, responseData)
				expectedCommands = append(expectedCommands, "j/monitors all")
			}
			_ = testutils.SetupFakeHyprIPCWriter(t,
				listener, responses, expectedCommands, true)

			// hypr: fake ipc events server
			if !tt.tui {
				eventsListener, teardownEvents := testutils.SetupHyprSocket(ctx, t,
					xdgRuntimeDir, signature, hypr.GetHyprEventsSocket)
				defer teardownEvents()
				_ = testutils.SetupFakeHyprEventsServerWithSleep(ctx, t,
					eventsListener, tt.hyprEvents, tt.initialHyprEventsSleep, tt.sleepBetweenHyprEvents)
			}

			// power: fake dbus session
			if !tt.disablePowerEvents {
				dbusService, testBusName, testObjectPath, cleanup := testutils.SetupTestDbusService(t)
				dbusService.SetProperty(power.BatteryPowerState)
				defer cleanup()
				tt.config = tt.config.WithPowerSection(
					testutils.CreatePowerConfig(testBusName, testObjectPath))
				_ = testutils.SetupFakeDbusEventsServer(t, dbusService,
					tt.powerEvents, tt.initialPowerEventsSleep,
					tt.sleepBetweenPowerEvents, binaryStartingChan)
			}

			// lid: fake dbus session
			if tt.enableLidEvents {
				dbusService, testBusName, testObjectPath, cleanup := testutils.SetupTestDbusService(t)
				// initial state to open
				dbusService.SetLidProperty(power.OpenedLidState)
				defer cleanup()
				tt.config = tt.config.WithLidSection(
					testutils.CreateLidConfig(testBusName, testObjectPath))
				_ = testutils.SetupFakeDbusLidEventsServer(t, dbusService,
					tt.lidEvents, tt.initialLidEventsSleep, tt.sleepBetweenLidEvents, binaryStartingChan)
			}

			flags := []string{}
			if tt.disablePowerEvents {
				flags = append(flags, "--disable-power-events")
			}
			if tt.enableLidEvents {
				flags = append(flags, "--enable-lid-events")
			}
			if tt.connectToSessionBus {
				flags = append(flags, "--connect-to-session-bus")
			}
			flagsString := strings.Join(flags, " ")

			// nolint:gosec
			cmd := exec.CommandContext(
				ctx,
				"vhs",
				"./preview/tapes/"+tt.tapeName+".tape")
			cmd.Dir = basepath
			cmd.Env = append(os.Environ(), "HDM_CONFIG="+tt.config.Get().Get().ConfigPath)
			cmd.Env = append(cmd.Env, "HDM_DESTINATION="+
				*tt.config.Get().Get().General.Destination)
			cmd.Env = append(cmd.Env, "HDM_EXTRA_FLAGS="+flagsString)

			cmd.Cancel = func() error {
				return cmd.Process.Signal(os.Interrupt)
			}
			cmd.WaitDelay = 200 * time.Millisecond

			go func() {
				// give time to warm up
				time.Sleep(100 * time.Millisecond)
				close(binaryStartingChan)
			}()
			out, err := cmd.CombinedOutput()
			assert.NoError(t, err, "vhs should succeed:\n %s", string(out))
			testutils.Logf(t, "Logs: %s", string(out))
		})
	}
}
