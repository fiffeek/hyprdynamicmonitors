package config

import (
	"path/filepath"
	"testing"
)

func TestLoad(t *testing.T) {
	tests := []struct {
		name          string
		configFile    string
		expectError   bool
		errorContains string
		validate      func(*testing.T, *Config)
	}{
		{
			name:       "valid basic config",
			configFile: "valid_basic.toml",
			validate: func(t *testing.T, c *Config) {
				if len(c.Profiles) != 3 {
					t.Errorf("expected 3 profiles, got %d", len(c.Profiles))
				}

				if c.General == nil || c.General.Destination == nil {
					t.Error("general.destination should not be nil")
				}
				if *c.General.Destination != "/tmp/test-monitors.conf" {
					t.Errorf("expected destination '/tmp/test-monitors.conf', got '%s'", *c.General.Destination)
				}

				if c.Scoring == nil {
					t.Error("scoring section should not be nil")
				}
				if *c.Scoring.NameMatch != 2 {
					t.Errorf("expected name_match 2, got %d", *c.Scoring.NameMatch)
				}
				if *c.Scoring.DescriptionMatch != 3 {
					t.Errorf("expected description_match 3, got %d", *c.Scoring.DescriptionMatch)
				}
				if *c.Scoring.PowerStateMatch != 1 {
					t.Errorf("expected power_state_match 1, got %d", *c.Scoring.PowerStateMatch)
				}

				laptop, exists := c.Profiles["laptop_only"]
				if !exists {
					t.Error("laptop_only profile should exist")
				} else {
					if laptop.Name != "laptop_only" {
						t.Errorf("expected profile name 'laptop_only', got '%s'", laptop.Name)
					}
					if *laptop.ConfigType != Static {
						t.Errorf("expected config type Static, got %v", *laptop.ConfigType)
					}
					if len(laptop.Conditions.RequiredMonitors) != 1 {
						t.Errorf("expected 1 required monitor, got %d", len(laptop.Conditions.RequiredMonitors))
					} else {
						monitor := laptop.Conditions.RequiredMonitors[0]
						if monitor.Name == nil || *monitor.Name != "eDP-1" {
							t.Errorf("expected monitor name 'eDP-1', got %v", monitor.Name)
						}
					}
				}

				acProfile, exists := c.Profiles["ac_power_profile"]
				if !exists {
					t.Error("ac_power_profile should exist")
				} else {
					if acProfile.Conditions.PowerState == nil || *acProfile.Conditions.PowerState != AC {
						t.Errorf("expected power state AC, got %v", acProfile.Conditions.PowerState)
					}
				}
			},
		},
		{
			name:       "valid minimal config",
			configFile: "valid_minimal.toml",
			validate: func(t *testing.T, c *Config) {
				if len(c.Profiles) != 1 {
					t.Errorf("expected 1 profile, got %d", len(c.Profiles))
				}

				if c.General.Destination == nil {
					t.Error("destination should have default value")
				}
				if c.Scoring.NameMatch == nil || *c.Scoring.NameMatch != 1 {
					t.Error("name_match should have default value of 1")
				}

				profile := c.Profiles["minimal"]
				if profile.ConfigType == nil || *profile.ConfigType != Static {
					t.Error("config_file_type should default to static")
				}
			},
		},
		{
			name:          "invalid - no profiles",
			configFile:    "invalid_no_profiles.toml",
			expectError:   true,
			errorContains: "no profiles defined",
		},
		{
			name:          "invalid - missing config file",
			configFile:    "invalid_missing_config_file.toml",
			expectError:   true,
			errorContains: "config_file is required",
		},
		{
			name:          "invalid - no required monitors",
			configFile:    "invalid_no_monitors.toml",
			expectError:   true,
			errorContains: "at least one required_monitor must be specified",
		},
		{
			name:          "invalid - monitor without name or description",
			configFile:    "invalid_monitor_no_name_desc.toml",
			expectError:   true,
			errorContains: "at least one of name, or description must be specified",
		},
		{
			name:          "invalid - scoring value zero",
			configFile:    "invalid_scoring_zero.toml",
			expectError:   true,
			errorContains: "score needs to be > 1",
		},
		{
			name:          "invalid - bad power state",
			configFile:    "invalid_power_state.toml",
			expectError:   true,
			errorContains: "invalid enum value",
		},
		{
			name:          "invalid - bad config file type",
			configFile:    "invalid_config_type.toml",
			expectError:   true,
			errorContains: "invalid enum value",
		},
		{
			name:          "file not found",
			configFile:    "nonexistent.toml",
			expectError:   true,
			errorContains: "not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configPath := filepath.Join("testdata", tt.configFile)

			config, err := Load(configPath)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
					return
				}
				if tt.errorContains != "" && !containsString(err.Error(), tt.errorContains) {
					t.Errorf("expected error to contain '%s', got '%s'", tt.errorContains, err.Error())
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if config == nil {
				t.Error("expected config but got nil")
				return
			}

			if tt.validate != nil {
				tt.validate(t, config)
			}
		})
	}
}

func TestGeneralSectionValidate(t *testing.T) {
	tests := []struct {
		name     string
		general  *GeneralSection
		expected string
	}{
		{
			name:     "nil destination gets default",
			general:  &GeneralSection{},
			expected: "$HOME/.config/hypr/monitors.conf",
		},
		{
			name: "existing destination is preserved",
			general: &GeneralSection{
				Destination: stringPtr("/custom/path.conf"),
			},
			expected: "/custom/path.conf",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.general.Validate()
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if tt.general.Destination == nil {
				t.Error("destination should not be nil after validation")
				return
			}

			if tt.name == "nil destination gets default" {
				if !containsString(*tt.general.Destination, ".config/hypr/monitors.conf") {
					t.Errorf("expected destination to contain default path, got '%s'", *tt.general.Destination)
				}
			} else {
				if *tt.general.Destination != tt.expected {
					t.Errorf("expected destination '%s', got '%s'", tt.expected, *tt.general.Destination)
				}
			}
		})
	}
}

func TestScoringSectionValidate(t *testing.T) {
	tests := []struct {
		name        string
		scoring     *ScoringSection
		expectError bool
	}{
		{
			name:    "nil values get defaults",
			scoring: &ScoringSection{},
		},
		{
			name: "existing values preserved",
			scoring: &ScoringSection{
				NameMatch:        intPtr(5),
				DescriptionMatch: intPtr(10),
				PowerStateMatch:  intPtr(3),
			},
		},
		{
			name: "zero value causes error",
			scoring: &ScoringSection{
				NameMatch:        intPtr(0),
				DescriptionMatch: intPtr(1),
				PowerStateMatch:  intPtr(1),
			},
			expectError: true,
		},
		{
			name: "negative value causes error",
			scoring: &ScoringSection{
				NameMatch:        intPtr(-1),
				DescriptionMatch: intPtr(1),
				PowerStateMatch:  intPtr(1),
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.scoring.Validate()

			if tt.expectError {
				if err == nil {
					t.Error("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if tt.scoring.NameMatch == nil || *tt.scoring.NameMatch < 1 {
				t.Error("name_match should be >= 1")
			}
			if tt.scoring.DescriptionMatch == nil || *tt.scoring.DescriptionMatch < 1 {
				t.Error("description_match should be >= 1")
			}
			if tt.scoring.PowerStateMatch == nil || *tt.scoring.PowerStateMatch < 1 {
				t.Error("power_state_match should be >= 1")
			}
		})
	}
}

func TestRequiredMonitorValidate(t *testing.T) {
	tests := []struct {
		name        string
		monitor     *RequiredMonitor
		expectError bool
	}{
		{
			name: "name only is valid",
			monitor: &RequiredMonitor{
				Name: stringPtr("eDP-1"),
			},
		},
		{
			name: "description only is valid",
			monitor: &RequiredMonitor{
				Description: stringPtr("BOE Screen"),
			},
		},
		{
			name: "both name and description is valid",
			monitor: &RequiredMonitor{
				Name:        stringPtr("eDP-1"),
				Description: stringPtr("BOE Screen"),
			},
		},
		{
			name: "monitor tag with name is valid",
			monitor: &RequiredMonitor{
				Name:       stringPtr("eDP-1"),
				MonitorTag: stringPtr("LaptopScreen"),
			},
		},
		{
			name: "only monitor tag is invalid",
			monitor: &RequiredMonitor{
				MonitorTag: stringPtr("LaptopScreen"),
			},
			expectError: true,
		},
		{
			name:        "empty monitor is invalid",
			monitor:     &RequiredMonitor{},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.monitor.Validate()

			if tt.expectError {
				if err == nil {
					t.Error("expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

func TestEnumUnmarshalTOML(t *testing.T) {
	t.Run("ConfigFileType", func(t *testing.T) {
		tests := []struct {
			name        string
			value       interface{}
			expected    ConfigFileType
			expectError bool
		}{
			{
				name:     "static",
				value:    "static",
				expected: Static,
			},
			{
				name:     "template",
				value:    "template",
				expected: Template,
			},
			{
				name:        "invalid string",
				value:       "invalid",
				expectError: true,
			},
			{
				name:        "non-string value",
				value:       123,
				expectError: true,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				var cft ConfigFileType
				err := cft.UnmarshalTOML(tt.value)

				if tt.expectError {
					if err == nil {
						t.Error("expected error but got none")
					}
				} else {
					if err != nil {
						t.Errorf("unexpected error: %v", err)
					}
					if cft != tt.expected {
						t.Errorf("expected %v, got %v", tt.expected, cft)
					}
				}
			})
		}
	})

	t.Run("PowerStateType", func(t *testing.T) {
		tests := []struct {
			name        string
			value       interface{}
			expected    PowerStateType
			expectError bool
		}{
			{
				name:     "AC",
				value:    "AC",
				expected: AC,
			},
			{
				name:     "BAT",
				value:    "BAT",
				expected: BAT,
			},
			{
				name:        "invalid string",
				value:       "INVALID",
				expectError: true,
			},
			{
				name:        "non-string value",
				value:       123,
				expectError: true,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				var pst PowerStateType
				err := pst.UnmarshalTOML(tt.value)

				if tt.expectError {
					if err == nil {
						t.Error("expected error but got none")
					}
				} else {
					if err != nil {
						t.Errorf("unexpected error: %v", err)
					}
					if pst != tt.expected {
						t.Errorf("expected %v, got %v", tt.expected, pst)
					}
				}
			})
		}
	})
}

func stringPtr(s string) *string {
	return &s
}

func intPtr(i int) *int {
	return &i
}

func containsString(haystack, needle string) bool {
	return len(haystack) >= len(needle) &&
		(haystack == needle ||
			haystack[:len(needle)] == needle ||
			haystack[len(haystack)-len(needle):] == needle ||
			containsSubstring(haystack, needle))
}

func containsSubstring(haystack, needle string) bool {
	for i := 0; i <= len(haystack)-len(needle); i++ {
		if haystack[i:i+len(needle)] == needle {
			return true
		}
	}
	return false
}
