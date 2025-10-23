package config_test

import (
	"context"
	"errors"
	"flag"
	"os"
	"path/filepath"
	"testing"

	"github.com/fiffeek/hyprdynamicmonitors/internal/config"
	"github.com/fiffeek/hyprdynamicmonitors/internal/testutils"
	"github.com/fiffeek/hyprdynamicmonitors/internal/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var regenerate = flag.Bool("regenerate", false, "regenerate fixtures instead of comparing")

func TestCreateDefaultConfig(t *testing.T) {
	tests := []struct {
		name          string
		upowerOutput  *string
		upowerErr     error
		expectError   bool
		errorContains string
		fixture       string
	}{
		{
			name:         "creates config with power line",
			upowerOutput: utils.StringPtr("/org/freedesktop/UPower/devices/line_power_ACAD"),
			fixture:      "testdata/fixtures/default_config_acad.toml",
		},
		{
			name:         "creates config with custom power line",
			upowerOutput: utils.StringPtr("/org/freedesktop/UPower/devices/line_power_AC"),
			fixture:      "testdata/fixtures/default_config_ac.toml",
		},
		{
			name:      "creates config when upower fails",
			upowerErr: errors.New("upower not available"),
			fixture:   "testdata/fixtures/default_config_acad.toml",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			configPath := filepath.Join(tmpDir, "test_config.toml")

			origRunCmd := utils.GetRunCmd()
			defer utils.SetRunCmd(origRunCmd)

			utils.SetRunCmd(func(ctx context.Context, name string, args ...string) (string, error) {
				if tt.upowerErr != nil {
					return "", tt.upowerErr
				}
				if tt.upowerOutput != nil {
					return *tt.upowerOutput, nil
				}
				return "/org/freedesktop/UPower/devices/line_power_ACAD", nil
			})

			err := config.CreateDefaultConfig(configPath)

			if tt.expectError {
				require.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
				return
			}

			require.NoError(t, err)
			_, err = os.Stat(configPath)
			require.NoError(t, err, "config file should exist")

			testutils.AssertFixture(t, configPath, tt.fixture, *regenerate)
		})
	}
}

func TestCreateDefaultConfigTemplate(t *testing.T) {
	t.Run("template renders correctly", func(t *testing.T) {
		tmpDir := t.TempDir()
		tmpDir = filepath.Join(tmpDir, "nonexistentdir")
		configPath := filepath.Join(tmpDir, "config.toml")

		origRunCmd := utils.GetRunCmd()
		defer utils.SetRunCmd(origRunCmd)

		powerLine := "/org/freedesktop/UPower/devices/line_power_ADP1"
		utils.SetRunCmd(func(ctx context.Context, name string, args ...string) (string, error) {
			return powerLine, nil
		})

		err := config.CreateDefaultConfig(configPath)
		require.NoError(t, err)

		cfg, err := config.Load(configPath)
		require.NoError(t, err)
		require.NotNil(t, cfg)

		assert.NotNil(t, cfg.General)
		assert.NotNil(t, cfg.PowerEvents)
		assert.NotNil(t, cfg.PowerEvents.DbusQueryObject)
		assert.Equal(t, powerLine, cfg.PowerEvents.DbusQueryObject.Path)
		require.Len(t, cfg.PowerEvents.DbusSignalMatchRules, 1)
		assert.Equal(t, powerLine, *cfg.PowerEvents.DbusSignalMatchRules[0].ObjectPath)
	})
}

func TestLoadCreatesDefaultConfigIfMissing(t *testing.T) {
	t.Run("creates default config when file doesn't exist", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "nonexistent.toml")

		origRunCmd := utils.GetRunCmd()
		defer utils.SetRunCmd(origRunCmd)

		utils.SetRunCmd(func(ctx context.Context, name string, args ...string) (string, error) {
			return "/org/freedesktop/UPower/devices/line_power_ACAD", nil
		})

		cfg, err := config.Load(configPath)
		require.NoError(t, err)
		require.NotNil(t, cfg)

		_, err = os.Stat(configPath)
		require.NoError(t, err, "config file should have been created")

		assert.NotNil(t, cfg.General)
		assert.NotNil(t, cfg.PowerEvents)
	})

	t.Run("uses existing config if present", func(t *testing.T) {
		cfg, err := config.Load(filepath.Join("testdata", "valid_basic.toml"))
		require.NoError(t, err)
		require.NotNil(t, cfg)
		require.Len(t, cfg.Profiles, 3)
	})
}

func TestCondFunction(t *testing.T) {
	tests := []struct {
		name     string
		cond     *string
		a        *string
		b        string
		expected string
	}{
		{
			name:     "condition is nil, returns b",
			cond:     nil,
			a:        utils.StringPtr("value_a"),
			b:        "value_b",
			expected: "value_b",
		},
		{
			name:     "condition is not nil, returns a",
			cond:     utils.StringPtr("something"),
			a:        utils.StringPtr("value_a"),
			b:        "value_b",
			expected: "value_a",
		},
		{
			name:     "condition is empty string, returns a",
			cond:     utils.StringPtr(""),
			a:        utils.StringPtr("value_a"),
			b:        "value_b",
			expected: "value_a",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := config.Cond(tt.cond, tt.a, tt.b)
			assert.Equal(t, tt.expected, result)
		})
	}
}
