package utils

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExecCommand(t *testing.T) {
	tests := []struct {
		name          string
		command       string
		args          []string
		expectError   bool
		errorContains string
	}{
		{
			name:    "successful echo command",
			command: "echo",
			args:    []string{"hello", "world"},
		},
		{
			name:    "successful printf command",
			command: "printf",
			args:    []string{"test output"},
		},
		{
			name:          "nonexistent command",
			command:       "this_command_does_not_exist_12345",
			args:          []string{},
			expectError:   true,
			errorContains: "failed",
		},
		{
			name:    "command with no output",
			command: "true",
			args:    []string{},
		},
		{
			name:          "command that fails",
			command:       "false",
			args:          []string{},
			expectError:   true,
			errorContains: "failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			output, err := execCommand(ctx, tt.command, tt.args...)

			if tt.expectError {
				require.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				require.NoError(t, err)
				assert.NotContains(t, output, "\n\n")
			}
		})
	}
}

func TestExecCommandContext(t *testing.T) {
	t.Run("respects context cancellation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		_, err := execCommand(ctx, "sleep", "10")
		require.Error(t, err)
	})
}

func TestRunCmdMocking(t *testing.T) {
	t.Run("can mock runCmd for testing", func(t *testing.T) {
		origRunCmd := runCmd
		defer func() { runCmd = origRunCmd }()

		called := false
		runCmd = func(ctx context.Context, name string, args ...string) (string, error) {
			called = true
			assert.Equal(t, "test-command", name)
			assert.Equal(t, []string{"arg1", "arg2"}, args)
			return "mocked output", nil
		}

		output, err := runCmd(context.Background(), "test-command", "arg1", "arg2")

		require.NoError(t, err)
		assert.Equal(t, "mocked output", output)
		assert.True(t, called, "mock should have been called")
	})

	t.Run("can mock runCmd to return error", func(t *testing.T) {
		origRunCmd := runCmd
		defer func() { runCmd = origRunCmd }()

		runCmd = func(ctx context.Context, name string, args ...string) (string, error) {
			return "", errors.New("mocked error")
		}

		_, err := runCmd(context.Background(), "any-command")

		require.Error(t, err)
		assert.Contains(t, err.Error(), "mocked error")
	})
}
