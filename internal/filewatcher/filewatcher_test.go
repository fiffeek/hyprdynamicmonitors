package filewatcher_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/fiffeek/hyprdynamicmonitors/internal/config"
	"github.com/fiffeek/hyprdynamicmonitors/internal/filewatcher"
	"github.com/fiffeek/hyprdynamicmonitors/internal/testutils"
	"github.com/fiffeek/hyprdynamicmonitors/internal/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupFilewatcherTest(t *testing.T) *config.Config {
	tempDir := t.TempDir()
	configDir := filepath.Join(tempDir, "config")
	profile1Dir := filepath.Join(tempDir, "profile1")
	profile2Dir := filepath.Join(tempDir, "profile2")
	require.NoError(t, os.MkdirAll(configDir, 0o750))
	require.NoError(t, os.MkdirAll(profile1Dir, 0o750))
	require.NoError(t, os.MkdirAll(profile2Dir, 0o750))

	configFile1 := filepath.Join(profile1Dir, "hypr1.conf")
	configFile2 := filepath.Join(profile2Dir, "hypr2.conf")
	require.NoError(t, os.WriteFile(configFile1, []byte("monitor=eDP-1,1920x1080@60,0x0,1"), 0o600))
	require.NoError(t, os.WriteFile(configFile2, []byte("monitor=DP-1,2560x1440@60,1920x0,1"), 0o600))

	return testutils.NewTestConfig(t).
		WithProfiles(map[string]*config.Profile{
			"profile1": {
				Name:       "profile1",
				ConfigFile: configFile1,
				Conditions: &config.ProfileCondition{
					PowerState: utils.JustPtr(config.BAT),
					RequiredMonitors: []*config.RequiredMonitor{
						{Name: utils.StringPtr("eDP-1")},
					},
				},
			},
			"profile2": {
				Name:       "profile2",
				ConfigFile: configFile2,
				Conditions: &config.ProfileCondition{
					PowerState: utils.JustPtr(config.AC),
					RequiredMonitors: []*config.RequiredMonitor{
						{Name: utils.StringPtr("DP-1")},
					},
				},
			},
		}).WithConfigDir(configDir).WithHotReload(&config.HotReloadSection{
		UpdateDebounceTimer: utils.JustPtr(50), // 50ms debounce for faster tests
	}).
		Get()
}

func TestFilewatcher_Integration(t *testing.T) {
	tests := []struct {
		name          string
		setupChange   func(*config.Config) error
		expectEvent   bool
		changeDesc    string
		validateAfter func(*config.Config, *filewatcher.Service, <-chan interface{}) error
	}{
		{
			name: "receives events when files in profile directories change",
			setupChange: func(cfg *config.Config) error {
				profile1Dir := cfg.Get().Profiles["profile1"].ConfigFileDir
				testFile := filepath.Join(profile1Dir, "test_file.txt")
				return os.WriteFile(testFile, []byte("test content"), 0o600)
			},
			expectEvent: true,
			changeDesc:  "new file in profile directory",
		},
		{
			name: "receives events when config file changes in flight",
			setupChange: func(cfg *config.Config) error {
				configFile1 := cfg.Get().Profiles["profile1"].ConfigFile
				newContent := []byte("monitor=eDP-1,1920x1080@60,0x0,1.5")
				return os.WriteFile(configFile1, newContent, 0o600)
			},
			expectEvent: true,
			changeDesc:  "modified existing config file",
		},
		{
			name: "receives events when files in config directory change",
			setupChange: func(cfg *config.Config) error {
				configDir := cfg.Get().ConfigDirPath
				configTestFile := filepath.Join(configDir, "config_test.toml")
				return os.WriteFile(configTestFile, []byte("[general]\npath=\"/tmp/test\""), 0o600)
			},
			expectEvent: true,
			changeDesc:  "new file in config directory",
		},

		{
			name: "validate config after changing it",
			setupChange: func(cfg *config.Config) error {
				profile1Dir := cfg.Get().Profiles["profile1"].ConfigFileDir
				testFile := filepath.Join(profile1Dir, "test_file.txt")
				return os.WriteFile(testFile, []byte("test content"), 0o600)
			},
			expectEvent: true,
			changeDesc:  "config edited",
			validateAfter: func(cfg *config.Config, svc *filewatcher.Service, events <-chan interface{}) error {
				drainEvents(t, events)
				configDirPath := cfg.Get().ConfigDirPath

				profile3Dir := filepath.Join(t.TempDir(), "profile3")
				require.NoError(t, os.MkdirAll(profile3Dir, 0o750))
				configFile3 := filepath.Join(profile3Dir, "hypr3.conf")
				require.NoError(t, os.WriteFile(configFile3, []byte("placeholder"), 0o600))

				testutils.NewTestConfig(t).WithConfigDir(configDirPath).WithProfiles(map[string]*config.Profile{
					"profile3": {
						Name:       "profile3",
						ConfigFile: configFile3,
						Conditions: &config.ProfileCondition{
							PowerState: utils.JustPtr(config.BAT),
							RequiredMonitors: []*config.RequiredMonitor{
								{Name: utils.StringPtr("eDP-1")},
							},
						},
					},
				}).WithHotReload(cfg.Get().HotReload).FillDefaults().SaveToFile()
				drainEvents(t, events)

				require.NoError(t, cfg.Reload(), "cant reload configuration")
				require.NoError(t, svc.Update(), "cant update service")
				drainEvents(t, events)

				newContent := []byte("monitor=eDP-1,1920x1080@60,0x0,1.5")
				require.NoError(t, os.WriteFile(configFile3, newContent, 0o600))

				select {
				case event := <-events:
					assert.NotNil(t, event, "should receive an event for a file update")
				case <-time.After(500 * time.Millisecond):
					t.Fatalf("no event after waiting")
				}

				return nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := setupFilewatcherTest(t)

			service := filewatcher.NewService(cfg, utils.BoolPtr(false))

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

			errCh := make(chan error, 1)
			go func() {
				errCh <- service.Run(ctx)
			}()

			time.Sleep(50 * time.Millisecond)

			events := service.Listen()

			require.NoError(t, tt.setupChange(cfg))

			if tt.expectEvent {
				select {
				case event := <-events:
					assert.NotNil(t, event, "should receive event for %s", tt.changeDesc)
				case <-time.After(500 * time.Millisecond):
					t.Fatalf("timeout waiting for event after %s", tt.changeDesc)
				}
			}

			if tt.validateAfter != nil {
				require.NoError(t, tt.validateAfter(cfg, service, events))
			}

			cancel()

			select {
			case err := <-errCh:
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "context canceled")
			case <-time.After(1 * time.Second):
				t.Fatal("timeout waiting for service to shut down")
			}
		})
	}
}

func drainEvents(t *testing.T, events <-chan interface{}) {
	for {
		t.Log("draining events")
		select {
		case event, ok := <-events:
			require.True(t, ok, "events channel closed")
			t.Log("event received")
			assert.NotNil(t, event, "should receive event")
		case <-time.After(250 * time.Millisecond):
			t.Log("no event, exiting")
			return
		}
	}
}

func TestFilewatcher_DisabledHotReload(t *testing.T) {
	cfg := testutils.NewTestConfig(t).Get()
	service := filewatcher.NewService(cfg, utils.BoolPtr(true))

	err := service.Update()
	assert.NoError(t, err, "Update should succeed when hot reload is disabled")
}

func TestFilewatcher_UpdateError(t *testing.T) {
	cfg := testutils.NewTestConfig(t).Get()
	service := filewatcher.NewService(cfg, utils.BoolPtr(false))

	err := service.Update()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no watcher assigned")
}
