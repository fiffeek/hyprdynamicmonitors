package test

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

var examples = filepath.Join(basepath, "examples")

func Test_Validate_Examples(t *testing.T) {
	files, err := find(examples, ".toml")
	require.NoError(t, err, "didnt find all example configs")
	for _, file := range files {
		t.Run(file, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
			defer cancel()

			done := make(chan any, 1)
			defer close(done)

			go func() {
				out, err := runBinary(t, ctx, []string{"--config", file, "validate"})
				require.NoError(t, err, "binary failed %s", string(out))
				done <- true
			}()

			select {
			case <-ctx.Done():
				require.NoError(t, ctx.Err(), "timeout")
			case <-done:
			}
		})
	}
}

func Test_Validate_InvalidConfigs(t *testing.T) {
	tests := []struct {
		name               string
		configFile         string
		doesNotExpectError bool
		expectedError      string // substring that should appear in error message
	}{
		{
			name:               "missing profiles",
			configFile:         "missing_profiles.toml",
			doesNotExpectError: true,
		},
		{
			name:          "invalid config file type",
			configFile:    "invalid_config_file_type.toml",
			expectedError: "invalid enum value",
		},
		{
			name:          "missing config file",
			configFile:    "missing_config_file.toml",
			expectedError: "config_file is required",
		},
		{
			name:          "nonexistent config file",
			configFile:    "nonexistent_config_file.toml",
			expectedError: "config file /nonexistent/path/to/config.conf not found",
		},
		{
			name:          "no required monitors",
			configFile:    "no_required_monitors.toml",
			expectedError: "at least one required_monitors must be specified",
		},
		{
			name:          "invalid monitor spec",
			configFile:    "invalid_monitor_spec.toml",
			expectedError: "at least one of name, or description must be specified",
		},
		{
			name:          "invalid power state",
			configFile:    "invalid_power_state.toml",
			expectedError: "invalid enum value",
		},
		{
			name:          "reserved template variable",
			configFile:    "reserved_template_variable.toml",
			expectedError: "PowerState cant be used since it is a reserved keyword",
		},
		{
			name:          "fallback with conditions",
			configFile:    "fallback_with_conditions.toml",
			expectedError: "fallback profile cant define any conditions",
		},
		{
			name:          "negative scoring",
			configFile:    "negative_scoring.toml",
			expectedError: "score needs to be > 1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configsDir := filepath.Join(basepath, "test", "testdata", "app", "configs")
			if tt.doesNotExpectError {
				configsDir = filepath.Join(configsDir, "valid")
			} else {
				configsDir = filepath.Join(configsDir, "invalid")
			}
			configPath := filepath.Join(configsDir, tt.configFile)

			ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
			defer cancel()

			done := make(chan struct{})
			var out []byte
			var err error

			go func() {
				defer close(done)
				out, err = runBinary(t, ctx, []string{"--config", configPath, "validate"})
			}()

			select {
			case <-ctx.Done():
				require.NoError(t, ctx.Err(), "timeout while running validation")
			case <-done:
				if tt.doesNotExpectError {
					require.NoError(t, err, "expected validation to succeed but it failed. Output: %s", string(out))
					return
				}
				require.Error(t, err, "expected validation to fail but it succeeded. Output: %s", string(out))
				require.Contains(t, string(out), tt.expectedError,
					"error message should contain expected substring. Got: %s", string(out))
			}
		})
	}
}
