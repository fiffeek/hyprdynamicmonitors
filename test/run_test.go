package test

import (
	"context"
	"os"
	"path/filepath"
	"syscall"
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

func Test__Run_Binary(t *testing.T) {
	tmpDir := t.TempDir()
	tests := []struct {
		name                     string
		description              string
		config                   *testutils.TestConfig
		extraArgs                []string
		expectError              bool
		expectErrorContains      string
		expectLogs               []utils.LogID
		hyprMonitorResponseFiles []string
		hyprEvents               []string
		powerEvents              []power.PowerState
		lidEvents                []power.LidState
		// waitForSideEffects ensures tests run faster if the side effects conditions are met
		// it will kill the current binary execution
		waitForSideEffects  func(context.Context, *testing.T, *config.RawConfig)
		validateSideEffects func(*testing.T, *config.RawConfig)
		disableHotReload    bool
		disablePowerEvents  bool
		enableLidEvents     bool
		connectToSessionBus bool
		runOnce             bool
		configUpdates       []*testutils.TestConfig
		sendSignal          *syscall.Signal
		// disablePassingArgs works for options like disableHotReload, which additionally to changing the test behavior
		// would add "--disable-auto-hot-reload" cli arg; in the context of anything but "run" it does not make sense, so this
		// is an explicit knob to change this behavior
		disablePassingArgs bool
	}{
		{
			name:                     "dry run should succeed",
			description:              "dry-run without power events should succeed by querying monitor data directly",
			config:                   testutils.NewTestConfig(t),
			extraArgs:                []string{"--dry-run"},
			expectLogs:               []utils.LogID{utils.DryRunLogID},
			hyprMonitorResponseFiles: []string{"testdata/hypr/server/basic_monitors.json"},
			disablePowerEvents:       true,
			disableHotReload:         true,
			runOnce:                  true,
		},

		{
			name:                     "basic templating",
			description:              "when hypr returns the same monitors as defined in the configuration, the template should match the golden file",
			config:                   createBasicTestConfig(t),
			hyprMonitorResponseFiles: []string{"testdata/hypr/server/basic_monitors.json"},
			validateSideEffects: func(t *testing.T, cfg *config.RawConfig) {
				testutils.AssertFileExists(t, *cfg.General.Destination)
				compareWithFixture(t, *cfg.General.Destination,
					"testdata/app/fixtures/basic_both.conf")
			},
			disablePowerEvents: true,
			disableHotReload:   true,
			runOnce:            true,
		},

		{
			name:        "disable templating extra",
			description: "when hypr returns more monitors that are defined in the profile the template can use templating to reason about the state",
			config: createBasicTestConfig(t).WithProfiles(
				map[string]*config.Profile{
					"one": {
						ConfigType: utils.JustPtr(config.Template),
						Conditions: &config.ProfileCondition{
							RequiredMonitors: []*config.RequiredMonitor{
								{
									Name: utils.StringPtr("eDP-1"),
								},
							},
						},
					},
				},
			).FillProfileConfigFile("one", "testdata/app/templates/disable_templating.go.tmpl"),
			hyprMonitorResponseFiles: []string{"testdata/hypr/server/basic_monitors.json"},
			validateSideEffects: func(t *testing.T, cfg *config.RawConfig) {
				testutils.AssertFileExists(t, *cfg.General.Destination)
				compareWithFixture(t, *cfg.General.Destination,
					"testdata/app/fixtures/disable_templating_extra_one.conf")
			},
			disablePowerEvents: true,
			disableHotReload:   true,
			runOnce:            true,
		},

		{
			name:        "disable templating none",
			description: "when hypr returns exadtly monitors that are defined in the profile the template can use templating to reason about the state",
			config: createBasicTestConfig(t).FillProfileConfigFile("both",
				"testdata/app/templates/disable_templating.go.tmpl"),
			hyprMonitorResponseFiles: []string{"testdata/hypr/server/basic_monitors.json"},
			validateSideEffects: func(t *testing.T, cfg *config.RawConfig) {
				testutils.AssertFileExists(t, *cfg.General.Destination)
				compareWithFixture(t, *cfg.General.Destination,
					"testdata/app/fixtures/disable_templating_extra_none.conf")
			},
			disablePowerEvents: true,
			disableHotReload:   true,
			runOnce:            true,
		},

		{
			name:                     "basic templating",
			description:              "when hypr returns disabled monitor it should still match the profile",
			config:                   createBasicTestConfig(t),
			hyprMonitorResponseFiles: []string{"testdata/hypr/server/basic_monitors_disabled.json"},
			validateSideEffects: func(t *testing.T, cfg *config.RawConfig) {
				testutils.AssertFileExists(t, *cfg.General.Destination)
				compareWithFixture(t, *cfg.General.Destination,
					"testdata/app/fixtures/basic_both_disabled.conf")
			},
			disablePowerEvents: true,
			disableHotReload:   true,
			runOnce:            true,
		},

		{
			name:        "no matches",
			description: "when no monitors match the configuration and there is no fallback, the service does nothing",
			config: createBasicTestConfig(t).WithProfiles(
				map[string]*config.Profile{
					"random_monitor": {
						ConfigType: utils.JustPtr(config.Static),
						Conditions: &config.ProfileCondition{
							RequiredMonitors: []*config.RequiredMonitor{
								{
									Name: utils.StringPtr("nonexistentmonitor"),
								},
							},
						},
					},
				},
			),
			hyprMonitorResponseFiles: []string{"testdata/hypr/server/basic_monitors.json"},
			validateSideEffects: func(t *testing.T, cfg *config.RawConfig) {
				testutils.AssertFileDoesNotExist(t, *cfg.General.Destination)
			},
			disablePowerEvents: true,
			disableHotReload:   true,
			runOnce:            true,
		},

		{
			name:        "user execs called",
			description: "when user defines callbacks, they should be called pre and post config application",
			config: createBasicTestConfig(t).WithPreExec(
				"touch " + filepath.Join(tmpDir, "pre.exec")).
				WithPostExec("touch " + filepath.Join(tmpDir, "post.exec")),
			hyprMonitorResponseFiles: []string{"testdata/hypr/server/basic_monitors.json"},
			expectLogs:               []utils.LogID{utils.PreExecLogID, utils.PostExecLogID},
			validateSideEffects: func(t *testing.T, cfg *config.RawConfig) {
				testutils.AssertFileExists(t, filepath.Join(tmpDir, "pre.exec"))
				testutils.AssertFileExists(t, filepath.Join(tmpDir, "post.exec"))
			},
			disablePowerEvents: true,
			disableHotReload:   true,
			runOnce:            true,
		},

		{
			name:        "user execs fail",
			description: "when user callbacks fail the service should operate as normal",
			config: createBasicTestConfig(t).WithPreExec("whatevercommandplaceholder").
				WithPostExec("whatevercommandplaceholder"),
			hyprMonitorResponseFiles: []string{"testdata/hypr/server/basic_monitors.json"},
			expectLogs:               []utils.LogID{utils.PreExecLogID, utils.PostExecLogID},
			validateSideEffects: func(t *testing.T, cfg *config.RawConfig) {
				testutils.AssertFileExists(t, *cfg.General.Destination)
				compareWithFixture(t, *cfg.General.Destination,
					"testdata/app/fixtures/basic_both.conf")
			},
			disablePowerEvents: true,
			disableHotReload:   true,
			runOnce:            true,
		},

		{
			name:        "basic templating mon events",
			description: "when monitor events are received, the service should regenerate configuration",
			config:      createBasicTestConfig(t),
			hyprMonitorResponseFiles: []string{
				"testdata/hypr/server/basic_monitors.json",
				"testdata/hypr/server/basic_monitors_one.json",
			},
			validateSideEffects: func(t *testing.T, cfg *config.RawConfig) {
				testutils.AssertFileExists(t, *cfg.General.Destination)
				compareWithFixture(t, *cfg.General.Destination, "testdata/app/fixtures/basic_one.conf")
			},
			disablePowerEvents: true,
			disableHotReload:   true,
			hyprEvents: []string{
				"monitorremovedv2>>2,DP-11,LG Electronics LG SDQHD 301NTBKDU037",
			},
			waitForSideEffects: func(ctx context.Context, t *testing.T, cfg *config.RawConfig) {
				funcs := []func() error{
					func() error {
						return testutils.ContentSameAsFixture(t, *cfg.General.Destination, "testdata/app/fixtures/basic_one.conf")
					},
				}
				waitTillHolds(ctx, t, funcs, 400*time.Millisecond)
			},
		},

		{
			name:        "power events templating",
			description: "when power events are enabled, dbus should be queried and return the state used for templating",
			config:      createBasicTestConfig(t).RequirePower(config.BAT),
			extraArgs:   []string{},
			hyprMonitorResponseFiles: []string{
				"testdata/hypr/server/basic_monitors_one.json",
			},
			validateSideEffects: func(t *testing.T, cfg *config.RawConfig) {
				testutils.AssertFileExists(t, *cfg.General.Destination)
				compareWithFixture(t, *cfg.General.Destination,
					"testdata/app/fixtures/basic_one_bat.conf")
			},
			disableHotReload:    true,
			runOnce:             true,
			connectToSessionBus: true,
		},

		{
			name:        "lid events templating",
			description: "when lid events are enabled, dbus should be queried and return the state used for templating",
			config:      createBasicTestConfig(t).RequireLid(config.OpenedLidStateType),
			extraArgs:   []string{},
			hyprMonitorResponseFiles: []string{
				"testdata/hypr/server/basic_monitors_one.json",
			},
			validateSideEffects: func(t *testing.T, cfg *config.RawConfig) {
				testutils.AssertFileExists(t, *cfg.General.Destination)
				compareWithFixture(t, *cfg.General.Destination,
					"testdata/app/fixtures/basic_one_lid_opened.conf")
			},
			disableHotReload:    true,
			runOnce:             true,
			disablePowerEvents:  true,
			enableLidEvents:     true,
			connectToSessionBus: true,
		},

		{
			name:        "power events triggers",
			description: "receiving power events should results in a configuration update",
			config:      createBasicTestConfig(t).RequirePower(config.AC),
			extraArgs:   []string{},
			hyprMonitorResponseFiles: []string{
				"testdata/hypr/server/basic_monitors_one.json",
			},
			waitForSideEffects: func(ctx context.Context, t *testing.T, cfg *config.RawConfig) {
				funcs := []func() error{
					func() error {
						return testutils.ContentSameAsFixture(t, *cfg.General.Destination,
							"testdata/app/fixtures/basic_one_ac.conf")
					},
				}
				waitTillHolds(ctx, t, funcs, 2000*time.Millisecond)
			},
			validateSideEffects: func(t *testing.T, cfg *config.RawConfig) {
				testutils.AssertFileExists(t, *cfg.General.Destination)
				compareWithFixture(t, *cfg.General.Destination,
					"testdata/app/fixtures/basic_one_ac.conf")
			},
			powerEvents:         []power.PowerState{power.ACPowerState},
			disableHotReload:    true,
			connectToSessionBus: true,
		},

		{
			name:        "lid events triggers",
			description: "receiving lid events should results in a configuration update",
			config:      createBasicTestConfig(t).RequireLid(config.ClosedLidStateType),
			extraArgs:   []string{},
			hyprMonitorResponseFiles: []string{
				"testdata/hypr/server/basic_monitors_one.json",
			},
			waitForSideEffects: func(ctx context.Context, t *testing.T, cfg *config.RawConfig) {
				funcs := []func() error{
					func() error {
						return testutils.ContentSameAsFixture(t, *cfg.General.Destination,
							"testdata/app/fixtures/basic_one_lid_closed.conf")
					},
				}
				waitTillHolds(ctx, t, funcs, 300*time.Millisecond)
			},
			validateSideEffects: func(t *testing.T, cfg *config.RawConfig) {
				testutils.AssertFileExists(t, *cfg.General.Destination)
				compareWithFixture(t, *cfg.General.Destination,
					"testdata/app/fixtures/basic_one_lid_closed.conf")
			},
			lidEvents:           []power.LidState{power.ClosedLidState},
			disableHotReload:    true,
			connectToSessionBus: true,
			disablePowerEvents:  true,
			enableLidEvents:     true,
		},

		{
			name:        "power mon events",
			description: "receiving both power and hypr events should results in a configuration update",
			config:      createBasicTestConfig(t).RequirePower(config.BAT),
			extraArgs:   []string{},
			hyprMonitorResponseFiles: []string{
				"testdata/hypr/server/basic_monitors_one.json",
				"testdata/hypr/server/basic_monitors.json",
			},
			hyprEvents: []string{
				"monitoraddedv2>>2,DP-11,LG Electronics LG SDQHD 301NTBKDU037",
			},
			waitForSideEffects: func(ctx context.Context, t *testing.T, cfg *config.RawConfig) {
				funcs := []func() error{
					func() error {
						return testutils.ContentSameAsFixture(t, *cfg.General.Destination,
							"testdata/app/fixtures/basic_both_bat.conf")
					},
				}
				waitTillHolds(ctx, t, funcs, 200*time.Millisecond)
			},
			validateSideEffects: func(t *testing.T, cfg *config.RawConfig) {
				testutils.AssertFileExists(t, *cfg.General.Destination)
				compareWithFixture(t, *cfg.General.Destination,
					"testdata/app/fixtures/basic_both_bat.conf")
			},
			powerEvents:         []power.PowerState{power.BatteryPowerState},
			disableHotReload:    true,
			connectToSessionBus: true,
		},

		{
			name:        "power mon lid",
			description: "receiving power, lid and hypr events should results in a configuration update",
			config:      createBasicTestConfig(t).RequirePower(config.BAT).RequireLid(config.ClosedLidStateType),
			extraArgs:   []string{},
			hyprMonitorResponseFiles: []string{
				"testdata/hypr/server/basic_monitors_one.json",
				"testdata/hypr/server/basic_monitors.json",
			},
			hyprEvents: []string{
				"monitoraddedv2>>2,DP-11,LG Electronics LG SDQHD 301NTBKDU037",
			},
			waitForSideEffects: func(ctx context.Context, t *testing.T, cfg *config.RawConfig) {
				funcs := []func() error{
					func() error {
						return testutils.ContentSameAsFixture(t, *cfg.General.Destination,
							"testdata/app/fixtures/basic_both_bat_closed.conf")
					},
				}
				waitTillHolds(ctx, t, funcs, 800*time.Millisecond)
			},
			validateSideEffects: func(t *testing.T, cfg *config.RawConfig) {
				testutils.AssertFileExists(t, *cfg.General.Destination)
				compareWithFixture(t, *cfg.General.Destination,
					"testdata/app/fixtures/basic_both_bat_closed.conf")
			},
			powerEvents:         []power.PowerState{power.BatteryPowerState},
			lidEvents:           []power.LidState{power.ClosedLidState},
			disableHotReload:    true,
			connectToSessionBus: true,
			enableLidEvents:     true,
		},

		{
			name:                     "reload cfg",
			description:              "configuration should hot-reload on the running service",
			config:                   createBasicTestConfig(t),
			hyprMonitorResponseFiles: []string{"testdata/hypr/server/basic_monitors.json"},
			validateSideEffects: func(t *testing.T, cfg *config.RawConfig) {
				testutils.AssertFileExists(t, *cfg.General.Destination)
				compareWithFixture(t, *cfg.General.Destination,
					"testdata/app/fixtures/basic_reloaded.conf")
			},
			waitForSideEffects: func(ctx context.Context, t *testing.T, cfg *config.RawConfig) {
				funcs := []func() error{
					func() error {
						return testutils.ContentSameAsFixture(t, *cfg.General.Destination,
							"testdata/app/fixtures/basic_reloaded.conf")
					},
				}
				waitTillHolds(ctx, t, funcs, 1500*time.Millisecond)
			},
			disablePowerEvents: true,
			configUpdates: []*testutils.TestConfig{
				createBasicTestConfig(t).WithProfiles(map[string]*config.Profile{
					"both_reloaded": {
						ConfigType: utils.JustPtr(config.Template),
						Conditions: &config.ProfileCondition{
							RequiredMonitors: []*config.RequiredMonitor{
								{
									Name:       utils.StringPtr("eDP-1"),
									MonitorTag: utils.StringPtr("EDP"),
								},
								{
									Name:       utils.StringPtr("DP-11"),
									MonitorTag: utils.StringPtr("DP"),
								},
							},
						},
					},
				}).FillProfileConfigFile("both_reloaded", "testdata/app/templates/basic_reloaded.toml"),
			},
		},

		{
			name:        "fallback profile",
			description: "fallback profile should be used when no profile matches",
			config: createBasicTestConfig(t).WithProfiles(map[string]*config.Profile{
				"random_monitor": {
					ConfigType: utils.JustPtr(config.Static),
					Conditions: &config.ProfileCondition{
						RequiredMonitors: []*config.RequiredMonitor{
							{
								Name: utils.StringPtr("nonexistentmonitor"),
							},
						},
					},
				},
			},
			).WithFallbackProfile(
				&config.Profile{
					ConfigType: utils.JustPtr(config.Static),
				},
			).FillFallbackProfileConfigFile("testdata/app/static/fallback.conf"),
			hyprMonitorResponseFiles: []string{"testdata/hypr/server/basic_monitors.json"},
			validateSideEffects: func(t *testing.T, cfg *config.RawConfig) {
				testutils.AssertFileExists(t, *cfg.General.Destination)
				testutils.AssertIsSymlink(t, *cfg.General.Destination)
				compareWithFixture(t, *cfg.General.Destination,
					"testdata/app/fixtures/fallback.conf")
			},
			disablePowerEvents: true,
			disableHotReload:   true,
			runOnce:            true,
		},

		{
			name:        "freeze fails for existing",
			description: "when freezing the current settings the profile with the same name must not exist",
			config: createBasicTestConfig(t).WithProfiles(map[string]*config.Profile{
				"one": {
					ConfigType: utils.JustPtr(config.Static),
					Conditions: &config.ProfileCondition{
						RequiredMonitors: []*config.RequiredMonitor{
							{
								Name: utils.StringPtr("nonexistentmonitor"),
							},
						},
					},
				},
			},
			),
			extraArgs:                []string{"freeze", "--profile-name", "one"},
			hyprMonitorResponseFiles: []string{"testdata/hypr/server/basic_monitors.json"},
			expectError:              true,
			disablePowerEvents:       true,
			disableHotReload:         true,
			runOnce:                  true,
			disablePassingArgs:       true,
			expectErrorContains:      "a profile with this name already exist",
		},

		{
			name:                     "freeze current settings as profile",
			description:              "when freezing the current profile we should see it appended to the current config",
			config:                   testutils.NewTestConfig(t),
			extraArgs:                []string{"freeze", "--profile-name", "hello"},
			hyprMonitorResponseFiles: []string{"testdata/hypr/server/basic_monitors.json"},
			validateSideEffects: func(t *testing.T, cfg *config.RawConfig) {
				// reread the config since it changed
				cfg, err := config.Load(cfg.ConfigPath)
				require.NoError(t, err, "the new config should be parseable")
				// hard to compare the golden files here since they rely on /tmp files
				assert.Len(t, cfg.Profiles, 2, "a new profile should be added")
				profile := cfg.Profiles["hello"]
				assert.NotNil(t, profile, "hello profile should exist")
				assert.Equal(t, utils.JustPtr(config.Template), profile.ConfigType)
				assert.Equal(t, &config.ProfileCondition{
					PowerState: nil,
					RequiredMonitors: []*config.RequiredMonitor{
						{
							Description: utils.StringPtr("BOE NE135A1M-NY1"),
							MonitorTag:  utils.StringPtr("monitor0"),
						},
						{
							Description: utils.StringPtr("LG Electronics LG SDQHD 301NTBKDU037"),
							MonitorTag:  utils.StringPtr("monitor1"),
						},
					},
				}, profile.Conditions)
				assert.NotNil(t, profile.ConfigFile, "the config file should be set")
				testutils.AssertFileExists(t, profile.ConfigFile)
				compareWithFixture(t, profile.ConfigFile,
					"testdata/app/fixtures/frozen_current.conf")
			},
			expectError:        false,
			disablePowerEvents: true,
			disableHotReload:   true,
			runOnce:            true,
			disablePassingArgs: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			binaryStartingChan := make(chan struct{})
			ctx, cancel := context.WithTimeout(context.Background(), 2000*time.Millisecond)
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
			fakeHyprServerDone := testutils.SetupFakeHyprIPCWriter(t,
				listener, responses, expectedCommands, true)

			// hypr: fake ipc events server, only needs to be run when the app runs
			var fakeHyprEventServerDone chan struct{}
			if !tt.runOnce {
				eventsListener, teardownEvents := testutils.SetupHyprSocket(ctx, t,
					xdgRuntimeDir, signature, hypr.GetHyprEventsSocket)
				defer teardownEvents()
				fakeHyprEventServerDone = testutils.SetupFakeHyprEventsServer(ctx, t, eventsListener, tt.hyprEvents)
			}

			// power: fake dbus session
			var dbusDone chan struct{}
			if !tt.disablePowerEvents {
				dbusService, testBusName, testObjectPath, cleanup := testutils.SetupTestDbusService(t)
				defer cleanup()
				tt.config = tt.config.WithPowerSection(
					testutils.CreatePowerConfig(testBusName, testObjectPath))
				dbusDone = testutils.SetupFakeDbusEventsServer(t, dbusService,
					tt.powerEvents, 100*time.Millisecond, 50*time.Millisecond, binaryStartingChan)
			}

			// lid: fake dbus session
			var lidDbusDone chan struct{}
			if tt.enableLidEvents {
				dbusService, testBusName, testObjectPath, cleanup := testutils.SetupTestDbusService(t)
				// initial state to open
				dbusService.SetLidProperty(power.OpenedLidState)
				defer cleanup()
				tt.config = tt.config.WithLidSection(
					testutils.CreateLidConfig(testBusName, testObjectPath))
				lidDbusDone = testutils.SetupFakeDbusLidEventsServer(t, dbusService,
					tt.lidEvents, 100*time.Millisecond, 50*time.Millisecond, binaryStartingChan)
			}

			// materialize config
			rawConfig := tt.config.Get().Get()

			// filewatcher
			var filewatcherDone chan struct{}
			if len(tt.configUpdates) > 0 {
				filewatcherDone = testutils.SetupFakeConfigUpdater(t,
					tt.configUpdates, 100*time.Millisecond, 50*time.Millisecond,
					binaryStartingChan, rawConfig.ConfigPath)
			}

			args := append([]string{
				"--config", rawConfig.ConfigPath,
				"--enable-json-logs-format",
			}, tt.extraArgs...)
			if !tt.disablePassingArgs {
				if tt.disableHotReload {
					args = append(args, "--disable-auto-hot-reload")
				}
				if tt.disablePowerEvents {
					args = append(args, "--disable-power-events")
				}
				if tt.enableLidEvents {
					args = append(args, "--enable-lid-events")
				}
				if tt.runOnce {
					args = append(args, "--run-once")
				}
				if tt.connectToSessionBus {
					args = append(args, "--connect-to-session-bus")
				}
				if *debug {
					args = append(args, "--debug")
				}

			}
			done := make(chan struct{})
			var out []byte
			var binaryErr error

			go func() {
				defer close(done)
				cmd := prepBinaryRun(ctx, args)
				go func() {
					// give time to warm up
					time.Sleep(100 * time.Millisecond)
					close(binaryStartingChan)
				}()
				out, binaryErr = cmd.CombinedOutput()
			}()

			// wait for filewatcher to write all files
			waitFor(t, filewatcherDone)
			if len(tt.configUpdates) > 0 {
				// save the last config, its fine if a file materializes here
				// since filewatcherDone means filewatcher finished writing anyway
				rawConfig = tt.configUpdates[len(tt.configUpdates)-1].Get().Get()
			}

			// speed up correct tests by explicitly waiting for side effects
			// instead of relying on auto context cancellation
			if tt.waitForSideEffects != nil {
				t.Log("Starting waitForSideEffects")
				tt.waitForSideEffects(ctx, t, rawConfig)
				// this will kill the running binary
				t.Log("waitForSideEffects returned, calling cancel()")
				cancel()
			}

			// wait on fakes to finish
			waitFor(t, dbusDone)
			waitFor(t, lidDbusDone)
			waitFor(t, fakeHyprServerDone)
			waitFor(t, fakeHyprEventServerDone)

			select {
			case <-time.After(3000 * time.Millisecond):
				assert.NoError(t, ctx.Err(), "timeout while running, out: %s", string(out))
				// explicitly kill the binary, then get the output
				cancel()
				assert.True(t, false, "timeout while running, out: %s", string(out))
			case <-done:
				t.Log(string(out))
				if tt.expectError {
					assert.Error(t, binaryErr, "expected run to fail but it succeeded. Output: %s", string(out))
					assert.Contains(t, string(out), tt.expectErrorContains,
						"error message should contain expected substring. Got: %s", string(out))
				} else {
					if tt.runOnce {
						assert.NoError(t, binaryErr, "expected run to not fail but it did. Output: %s", string(out))
					} else {
						assert.Error(t, binaryErr, "expected the program to be killed")
					}
					testutils.AssertLogsPresent(t, out, tt.expectLogs)
				}
				if tt.validateSideEffects != nil {
					tt.validateSideEffects(t, rawConfig)
				}
			}
		})
	}
}

func waitFor(t *testing.T, server chan struct{}) {
	if server == nil {
		return
	}

	select {
	case <-server:
	case <-time.After(800 * time.Millisecond):
		assert.True(t, false, "Hypr fake server didn't finish in time")
	}
}

func createBasicTestConfig(t *testing.T) *testutils.TestConfig {
	return testutils.NewTestConfig(t).WithServiceDebounceTime(50).WithFilewatcherDebounceTime(50).
		WithNotifications(&config.Notifications{Disabled: utils.BoolPtr(true)}).
		WithProfiles(map[string]*config.Profile{
			"both": {
				ConfigType: utils.JustPtr(config.Template),
				Conditions: &config.ProfileCondition{
					RequiredMonitors: []*config.RequiredMonitor{
						{
							Name:       utils.StringPtr("eDP-1"),
							MonitorTag: utils.StringPtr("EDP"),
						},
						{
							Name:       utils.StringPtr("DP-11"),
							MonitorTag: utils.StringPtr("DP"),
						},
					},
				},
			},
			"one": {
				ConfigType: utils.JustPtr(config.Template),
				Conditions: &config.ProfileCondition{
					RequiredMonitors: []*config.RequiredMonitor{
						{
							Name:       utils.StringPtr("eDP-1"),
							MonitorTag: utils.StringPtr("EDP"),
						},
					},
				},
			},
		}).FillProfileConfigFile("both", "testdata/app/templates/basic.toml").
		FillProfileConfigFile("one", "testdata/app/templates/basic.toml")
}
