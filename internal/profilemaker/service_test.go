package profilemaker_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/fiffeek/hyprdynamicmonitors/internal/config"
	"github.com/fiffeek/hyprdynamicmonitors/internal/hypr"
	"github.com/fiffeek/hyprdynamicmonitors/internal/profilemaker"
	"github.com/fiffeek/hyprdynamicmonitors/internal/testutils"
	"github.com/fiffeek/hyprdynamicmonitors/internal/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestService_EditExisting(t *testing.T) {
	// Sample monitors data
	monitors := []*hypr.MonitorSpec{
		{
			ID:          utils.IntPtr(1),
			Description: "New Monitor A",
			Width:       2560,
			Height:      1440,
			RefreshRate: 120.0,
			X:           0,
			Y:           0,
			Scale:       1.5,
			Transform:   0,
		},
		{
			ID:          utils.IntPtr(2),
			Description: "New Monitor B",
			Width:       1920,
			Height:      1080,
			RefreshRate: 60.0,
			X:           2560,
			Y:           0,
			Scale:       1.0,
			Transform:   0,
			Mirror:      "eDP-1",
		},
	}

	testCases := []struct {
		name          string
		inputFile     string
		expectedFile  string
		profileName   string
		expectError   bool
		errorContains string
	}{
		{
			name:         "Replace content between existing markers",
			inputFile:    "testdata/existing_config_with_markers.conf",
			expectedFile: "testdata/expected_replace_markers.conf",
			profileName:  "test-profile",
		},
		{
			name:         "Append content when no markers exist",
			inputFile:    "testdata/existing_config_no_markers.conf",
			expectedFile: "testdata/expected_append_no_markers.conf",
			profileName:  "test-profile",
		},
		{
			name:         "Append content when markers are broken",
			inputFile:    "testdata/config_with_broken_markers.conf",
			expectedFile: "testdata/expected_append_broken_markers.conf",
			profileName:  "test-profile",
		},
		{
			name:         "Handle empty config file",
			inputFile:    "testdata/empty_config.conf",
			expectedFile: "testdata/expected_empty_config.conf",
			profileName:  "test-profile",
		},
		{
			name:         "Handle non-existent file",
			inputFile:    "testdata/non_existent.conf",
			expectedFile: "testdata/expected_non_existent.conf",
			profileName:  "test-profile",
		},
		{
			name:          "Profile not found",
			inputFile:     "testdata/empty_config.conf",
			profileName:   "non-existent-profile",
			expectError:   true,
			errorContains: "profile not found",
		},
		{
			name:         "Repeated calls should not accumulate newlines",
			inputFile:    "testdata/existing_config_with_markers.conf",
			expectedFile: "testdata/expected_replace_markers.conf",
			profileName:  "test-profile",
		},
		{
			name:         "Repeated calls with markers only should not accumulate newlines",
			inputFile:    "testdata/markers_only.conf",
			expectedFile: "testdata/expected_markers_only.conf",
			profileName:  "test-profile",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tempDir, err := os.MkdirTemp("", "profilemaker_test_")
			require.NoError(t, err)
			defer os.RemoveAll(tempDir)

			configFile := filepath.Join(tempDir, "test_config.conf")

			var cfg *config.Config
			if tc.expectError && tc.errorContains == "profile not found" {
				cfg = testutils.NewTestConfig(t).Get()
			} else {
				profile := &config.Profile{
					Name:       tc.profileName,
					ConfigFile: configFile,
					ConfigType: utils.JustPtr(config.Template),
					Conditions: &config.ProfileCondition{
						RequiredMonitors: []*config.RequiredMonitor{
							{Name: utils.StringPtr("eDP-1")},
						},
					},
				}

				cfg = testutils.NewTestConfig(t).
					WithProfiles(map[string]*config.Profile{
						tc.profileName: profile,
					}).
					Get()
			}

			if tc.inputFile != "testdata/non_existent.conf" {
				inputData, err := os.ReadFile(tc.inputFile)
				require.NoError(t, err)
				err = os.WriteFile(configFile, inputData, 0o644)
				require.NoError(t, err)
			}

			service := profilemaker.NewService(cfg, nil)

			// For repeated calls test, call EditExisting multiple times
			if strings.Contains(tc.name, "Repeated calls") {
				for i := 0; i < 3; i++ {
					err = service.EditExisting(tc.profileName, monitors)
					if err != nil {
						break
					}
				}
			} else {
				err = service.EditExisting(tc.profileName, monitors)
			}

			if tc.expectError {
				assert.Error(t, err)
				if tc.errorContains != "" {
					assert.Contains(t, err.Error(), tc.errorContains)
				}
				return
			}

			assert.NoError(t, err)

			resultContent, err := os.ReadFile(configFile)
			require.NoError(t, err)

			expectedContent, err := os.ReadFile(tc.expectedFile)
			require.NoError(t, err)

			actualOutput := strings.TrimSpace(string(resultContent))
			expectedOutput := strings.TrimSpace(string(expectedContent))

			assert.Equal(t, expectedOutput, actualOutput)

			if !tc.expectError {
				assert.Contains(t, string(resultContent), "# <<<<< TUI AUTO START")
				assert.Contains(t, string(resultContent), "# <<<<< TUI AUTO END")
			}
		})
	}
}
