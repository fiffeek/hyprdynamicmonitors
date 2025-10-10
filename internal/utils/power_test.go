package utils

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetPowerLine(t *testing.T) {
	tests := []struct {
		name          string
		upowerOutput  string
		upowerErr     error
		expectedPath  *string
		expectError   bool
		errorContains string
	}{
		{
			name: "finds preferred AC adapter",
			upowerOutput: `/org/freedesktop/UPower/devices/battery_BAT0
/org/freedesktop/UPower/devices/line_power_ACAD
/org/freedesktop/UPower/devices/DisplayDevice`,
			expectedPath: stringPtr("/org/freedesktop/UPower/devices/line_power_ACAD"),
		},
		{
			name: "finds first matching preferred suffix",
			upowerOutput: `/org/freedesktop/UPower/devices/battery_BAT0
/org/freedesktop/UPower/devices/line_power_AC
/org/freedesktop/UPower/devices/line_power_ACAD
/org/freedesktop/UPower/devices/DisplayDevice`,
			expectedPath: stringPtr("/org/freedesktop/UPower/devices/line_power_AC"),
		},
		{
			name: "finds ADP1 adapter",
			upowerOutput: `/org/freedesktop/UPower/devices/battery_BAT0
/org/freedesktop/UPower/devices/line_power_ADP1`,
			expectedPath: stringPtr("/org/freedesktop/UPower/devices/line_power_ADP1"),
		},
		{
			name: "finds Mains adapter",
			upowerOutput: `/org/freedesktop/UPower/devices/battery_BAT0
/org/freedesktop/UPower/devices/line_power_Mains`,
			expectedPath: stringPtr("/org/freedesktop/UPower/devices/line_power_Mains"),
		},
		{
			name: "finds USB power",
			upowerOutput: `/org/freedesktop/UPower/devices/battery_BAT0
/org/freedesktop/UPower/devices/line_power_USB`,
			expectedPath: stringPtr("/org/freedesktop/UPower/devices/line_power_USB"),
		},
		{
			name: "finds wireless charger",
			upowerOutput: `/org/freedesktop/UPower/devices/battery_BAT0
/org/freedesktop/UPower/devices/line_power_wireless`,
			expectedPath: stringPtr("/org/freedesktop/UPower/devices/line_power_wireless"),
		},
		{
			name: "no matching power line",
			upowerOutput: `/org/freedesktop/UPower/devices/battery_BAT0
/org/freedesktop/UPower/devices/DisplayDevice`,
			expectError:   true,
			errorContains: "cant find the current power line",
		},
		{
			name:          "upower command fails",
			upowerErr:     errors.New("upower not found"),
			expectError:   true,
			errorContains: "cant get the current power line",
		},
		{
			name:          "empty upower output",
			upowerOutput:  "",
			expectError:   true,
			errorContains: "cant find the current power line",
		},
		{
			name: "whitespace only lines are ignored",
			upowerOutput: `

/org/freedesktop/UPower/devices/line_power_ACAD

`,
			expectedPath: stringPtr("/org/freedesktop/UPower/devices/line_power_ACAD"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			origRunCmd := runCmd
			defer func() { runCmd = origRunCmd }()

			runCmd = func(ctx context.Context, name string, args ...string) (string, error) {
				require.Equal(t, "upower", name)
				require.Equal(t, []string{"-e"}, args)
				return tt.upowerOutput, tt.upowerErr
			}

			result, err := GetPowerLine()

			if tt.expectError {
				require.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
				assert.Equal(t, *tt.expectedPath, *result)
			}
		})
	}
}

func TestMatchesPreferred(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		{
			name:     "matches line_power_AC",
			path:     "/org/freedesktop/UPower/devices/line_power_AC",
			expected: true,
		},
		{
			name:     "matches line_power_ACAD",
			path:     "/org/freedesktop/UPower/devices/line_power_ACAD",
			expected: true,
		},
		{
			name:     "matches line_power_ADP1",
			path:     "/org/freedesktop/UPower/devices/line_power_ADP1",
			expected: true,
		},
		{
			name:     "matches line_power_wireless",
			path:     "/org/freedesktop/UPower/devices/line_power_wireless",
			expected: true,
		},
		{
			name:     "does not match battery",
			path:     "/org/freedesktop/UPower/devices/battery_BAT0",
			expected: false,
		},
		{
			name:     "does not match DisplayDevice",
			path:     "/org/freedesktop/UPower/devices/DisplayDevice",
			expected: false,
		},
		{
			name:     "does not match empty string",
			path:     "",
			expected: false,
		},
		{
			name:     "does not match random device",
			path:     "/org/freedesktop/UPower/devices/line_power_UNKNOWN",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := matchesPreferred(tt.path)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func stringPtr(s string) *string {
	return &s
}
