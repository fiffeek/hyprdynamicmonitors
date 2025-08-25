package config_test

import (
	"path/filepath"
	"testing"

	"github.com/fiffeek/hyprdynamicmonitors/internal/config"
	"github.com/fiffeek/hyprdynamicmonitors/internal/utils"
	"github.com/stretchr/testify/assert"
)

func TestLoad(t *testing.T) {
	tests := []struct {
		name          string
		configFile    string
		expectError   bool
		errorContains string
		validate      func(*testing.T, *config.Config)
	}{
		{
			name:       "valid basic config",
			configFile: "valid_basic.toml",
			validate: func(t *testing.T, c *config.Config) {
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
					if *laptop.ConfigType != config.Static {
						t.Errorf("expected config type Static, got %v", *laptop.ConfigType)
					}
					if len(laptop.Conditions.RequiredMonitors) != 1 {
						t.Errorf("expected 1 required monitor, got %d",
							len(laptop.Conditions.RequiredMonitors))
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
				} else if acProfile.Conditions.PowerState == nil ||
					*acProfile.Conditions.PowerState != config.AC {
					t.Errorf("expected power state AC, got %v", acProfile.Conditions.PowerState)
				}

				if c.PowerEvents == nil {
					t.Error("power events section should not be nil")
				} else {
					if len(c.PowerEvents.DbusSignalMatchRules) < 2 {
						t.Errorf("expected at least 2 custom dbus match rules, got %d",
							len(c.PowerEvents.DbusSignalMatchRules))
					}

					if len(c.PowerEvents.DbusSignalReceiveFilters) < 2 {
						t.Errorf("expected at least 2 custom dbus receive filters, got %d",
							len(c.PowerEvents.DbusSignalReceiveFilters))
					}
				}
			},
		},
		{
			name:       "valid minimal config",
			configFile: "valid_minimal.toml",
			validate: func(t *testing.T, c *config.Config) {
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
				if profile.ConfigType == nil || *profile.ConfigType != config.Static {
					t.Error("config_file_type should default to static")
				}

				if c.PowerEvents == nil {
					t.Error("power events section should not be nil after validation")
				} else {
					if len(c.PowerEvents.DbusSignalMatchRules) != 3 {
						t.Errorf("expected 3 default dbus match rules, got %d",
							len(c.PowerEvents.DbusSignalMatchRules))
					}

					if len(c.PowerEvents.DbusSignalReceiveFilters) != 3 {
						t.Errorf("expected 3 default dbus receive filters, got %d",
							len(c.PowerEvents.DbusSignalReceiveFilters))
					}

					expectedRules := map[string]bool{
						"DeviceAdded":       false,
						"DeviceRemoved":     false,
						"PropertiesChanged": false,
					}
					for _, rule := range c.PowerEvents.DbusSignalMatchRules {
						if rule.Member != nil {
							expectedRules[*rule.Member] = true
						}
					}
					for member, found := range expectedRules {
						if !found {
							t.Errorf("expected default rule for %s not found", member)
						}
					}
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
			name:       "valid custom upower query",
			configFile: "valid_custom_upower_query.toml",
			validate: func(t *testing.T, c *config.Config) {
				if c.PowerEvents == nil {
					t.Error("power events section should not be nil")
					return
				}

				if c.PowerEvents.DbusQueryObject == nil {
					t.Error("dbus query object should not be nil")
					return
				}

				expectedQuery := &config.DbusQueryObject{
					Destination: "org.freedesktop.UPower",
					Path:        "/org/freedesktop/UPower",
					Method:      "org.freedesktop.DBus.Properties.Get",
					Args: []config.DbusQueryObjectArg{
						{Arg: "org.freedesktop.UPower"},
						{Arg: "LidIsPresent"},
					},
				}

				assert.Equal(t, expectedQuery, c.PowerEvents.DbusQueryObject,
					"DbusQueryObject should match expected")

				expectedCollectedArgs := []interface{}{"org.freedesktop.UPower", "LidIsPresent"}
				collectedArgs := c.PowerEvents.DbusQueryObject.CollectArgs()
				assert.Equal(t, expectedCollectedArgs, collectedArgs, "collected args should match")
			},
		},
		{
			name:          "invalid - empty upower destination",
			configFile:    "invalid_upower_empty_destination.toml",
			expectError:   true,
			errorContains: "destination cant be empty",
		},
		{
			name:          "invalid - empty upower args",
			configFile:    "invalid_upower_empty_args.toml",
			expectError:   true,
			errorContains: "arg cant be empty",
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

			config, err := config.Load(configPath)

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
		general  *config.GeneralSection
		expected string
	}{
		{
			name:     "nil destination gets default",
			general:  &config.GeneralSection{},
			expected: "$HOME/.config/hypr/monitors.conf",
		},
		{
			name: "existing destination is preserved",
			general: &config.GeneralSection{
				Destination: utils.StringPtr("/custom/path.conf"),
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
		scoring     *config.ScoringSection
		expectError bool
	}{
		{
			name:    "nil values get defaults",
			scoring: &config.ScoringSection{},
		},
		{
			name: "existing values preserved",
			scoring: &config.ScoringSection{
				NameMatch:        utils.IntPtr(5),
				DescriptionMatch: utils.IntPtr(10),
				PowerStateMatch:  utils.IntPtr(3),
			},
		},
		{
			name: "zero value causes error",
			scoring: &config.ScoringSection{
				NameMatch:        utils.IntPtr(0),
				DescriptionMatch: utils.IntPtr(1),
				PowerStateMatch:  utils.IntPtr(1),
			},
			expectError: true,
		},
		{
			name: "negative value causes error",
			scoring: &config.ScoringSection{
				NameMatch:        utils.IntPtr(-1),
				DescriptionMatch: utils.IntPtr(1),
				PowerStateMatch:  utils.IntPtr(1),
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
		monitor     *config.RequiredMonitor
		expectError bool
	}{
		{
			name: "name only is valid",
			monitor: &config.RequiredMonitor{
				Name: utils.StringPtr("eDP-1"),
			},
		},
		{
			name: "description only is valid",
			monitor: &config.RequiredMonitor{
				Description: utils.StringPtr("BOE Screen"),
			},
		},
		{
			name: "both name and description is valid",
			monitor: &config.RequiredMonitor{
				Name:        utils.StringPtr("eDP-1"),
				Description: utils.StringPtr("BOE Screen"),
			},
		},
		{
			name: "monitor tag with name is valid",
			monitor: &config.RequiredMonitor{
				Name:       utils.StringPtr("eDP-1"),
				MonitorTag: utils.StringPtr("LaptopScreen"),
			},
		},
		{
			name: "only monitor tag is invalid",
			monitor: &config.RequiredMonitor{
				MonitorTag: utils.StringPtr("LaptopScreen"),
			},
			expectError: true,
		},
		{
			name:        "empty monitor is invalid",
			monitor:     &config.RequiredMonitor{},
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
			expected    config.ConfigFileType
			expectError bool
		}{
			{
				name:     "static",
				value:    "static",
				expected: config.Static,
			},
			{
				name:     "template",
				value:    "template",
				expected: config.Template,
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
				var cft config.ConfigFileType
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
			expected    config.PowerStateType
			expectError bool
		}{
			{
				name:     "AC",
				value:    "AC",
				expected: config.AC,
			},
			{
				name:     "BAT",
				value:    "BAT",
				expected: config.BAT,
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
				var pst config.PowerStateType
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

func TestPowerSectionValidate(t *testing.T) {
	tests := []struct {
		name        string
		powerEvents *config.PowerSection
		expectError bool
	}{
		{
			name:        "nil power section gets defaults",
			powerEvents: &config.PowerSection{},
		},
		{
			name: "existing rules preserved",
			powerEvents: &config.PowerSection{
				DbusSignalMatchRules: []*config.DbusSignalMatchRule{
					{
						Sender:    utils.StringPtr("custom.sender"),
						Interface: utils.StringPtr("custom.interface"),
					},
				},
				DbusSignalReceiveFilters: []*config.DbusSignalReceiveFilter{
					{Name: utils.StringPtr("custom.signal")},
				},
			},
		},
		{
			name: "invalid match rule - all nil",
			powerEvents: &config.PowerSection{
				DbusSignalMatchRules: []*config.DbusSignalMatchRule{
					{}, // empty rule should fail validation
				},
			},
			expectError: true,
		},
		{
			name: "invalid receive filter - nil name",
			powerEvents: &config.PowerSection{
				DbusSignalReceiveFilters: []*config.DbusSignalReceiveFilter{
					{}, // empty filter should fail validation
				},
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.powerEvents.Validate()

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

			if tt.name == "nil power section gets defaults" {
				if len(tt.powerEvents.DbusSignalMatchRules) != 3 {
					t.Errorf("expected 3 default match rules, got %d",
						len(tt.powerEvents.DbusSignalMatchRules))
				}
				if len(tt.powerEvents.DbusSignalReceiveFilters) != 3 {
					t.Errorf("expected 3 default receive filters, got %d",
						len(tt.powerEvents.DbusSignalReceiveFilters))
				}

				expectedSignals := []string{
					"org.freedesktop.DBus.Properties.PropertiesChanged",
					"org.freedesktop.UPower.DeviceAdded",
					"org.freedesktop.UPower.DeviceRemoved",
				}
				for i, filter := range tt.powerEvents.DbusSignalReceiveFilters {
					if filter.Name == nil {
						t.Errorf("filter %d name should not be nil", i)
					} else {
						found := false
						for _, expected := range expectedSignals {
							if *filter.Name == expected {
								found = true
								break
							}
						}
						if !found {
							t.Errorf("unexpected default filter name: %s", *filter.Name)
						}
					}
				}
			}
		})
	}
}

func TestDbusQueryObjectValidate(t *testing.T) {
	tests := []struct {
		name          string
		queryObject   *config.DbusQueryObject
		expectError   bool
		errorContains string
	}{
		{
			name: "valid upower query object",
			queryObject: &config.DbusQueryObject{
				Destination: "org.freedesktop.UPower",
				Path:        "/org/freedesktop/UPower",
				Method:      "org.freedesktop.DBus.Properties.Get",
				Args: []config.DbusQueryObjectArg{
					{Arg: "org.freedesktop.UPower"},
					{Arg: "OnBattery"},
				},
			},
		},
		{
			name: "custom upower query object with different property",
			queryObject: &config.DbusQueryObject{
				Destination: "org.freedesktop.UPower",
				Path:        "/org/freedesktop/UPower",
				Method:      "org.freedesktop.DBus.Properties.Get",
				Args: []config.DbusQueryObjectArg{
					{Arg: "org.freedesktop.UPower"},
					{Arg: "LidIsPresent"},
				},
			},
		},
		{
			name: "custom destination and path",
			queryObject: &config.DbusQueryObject{
				Destination: "org.custom.PowerManager",
				Path:        "/org/custom/PowerManager",
				Method:      "org.freedesktop.DBus.Properties.Get",
				Args: []config.DbusQueryObjectArg{
					{Arg: "org.custom.PowerManager"},
					{Arg: "PowerState"},
				},
			},
		},
		{
			name: "empty destination",
			queryObject: &config.DbusQueryObject{
				Destination: "",
				Path:        "/org/freedesktop/UPower",
				Method:      "org.freedesktop.DBus.Properties.Get",
				Args: []config.DbusQueryObjectArg{
					{Arg: "org.freedesktop.UPower"},
					{Arg: "OnBattery"},
				},
			},
			expectError:   true,
			errorContains: "destination cant be empty",
		},
		{
			name: "empty path",
			queryObject: &config.DbusQueryObject{
				Destination: "org.freedesktop.UPower",
				Path:        "",
				Method:      "org.freedesktop.DBus.Properties.Get",
				Args: []config.DbusQueryObjectArg{
					{Arg: "org.freedesktop.UPower"},
					{Arg: "OnBattery"},
				},
			},
			expectError:   true,
			errorContains: "path cant be empty",
		},
		{
			name: "empty method",
			queryObject: &config.DbusQueryObject{
				Destination: "org.freedesktop.UPower",
				Path:        "/org/freedesktop/UPower",
				Method:      "",
				Args: []config.DbusQueryObjectArg{
					{Arg: "org.freedesktop.UPower"},
					{Arg: "OnBattery"},
				},
			},
			expectError:   true,
			errorContains: "method cant be empty",
		},
		{
			name: "empty arg",
			queryObject: &config.DbusQueryObject{
				Destination: "org.freedesktop.UPower",
				Path:        "/org/freedesktop/UPower",
				Method:      "org.freedesktop.DBus.Properties.Get",
				Args: []config.DbusQueryObjectArg{
					{Arg: "org.freedesktop.UPower"},
					{Arg: ""},
				},
			},
			expectError:   true,
			errorContains: "arg cant be empty",
		},
		{
			name: "no args",
			queryObject: &config.DbusQueryObject{
				Destination: "org.freedesktop.UPower",
				Path:        "/org/freedesktop/UPower",
				Method:      "org.freedesktop.DBus.Properties.GetAll",
				Args:        []config.DbusQueryObjectArg{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.queryObject.Validate()

			if tt.expectError {
				if err == nil {
					t.Error("expected error but got none")
					return
				}
				if tt.errorContains != "" && !containsString(err.Error(), tt.errorContains) {
					t.Errorf("expected error to contain '%s', got '%s'", tt.errorContains, err.Error())
				}
			} else if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestDbusQueryObjectCollectArgs(t *testing.T) {
	tests := []struct {
		name        string
		queryObject *config.DbusQueryObject
		expected    []interface{}
	}{
		{
			name: "standard upower query args",
			queryObject: &config.DbusQueryObject{
				Args: []config.DbusQueryObjectArg{
					{Arg: "org.freedesktop.UPower"},
					{Arg: "OnBattery"},
				},
			},
			expected: []interface{}{"org.freedesktop.UPower", "OnBattery"},
		},
		{
			name: "single arg",
			queryObject: &config.DbusQueryObject{
				Args: []config.DbusQueryObjectArg{
					{Arg: "org.freedesktop.UPower"},
				},
			},
			expected: []interface{}{"org.freedesktop.UPower"},
		},
		{
			name: "no args",
			queryObject: &config.DbusQueryObject{
				Args: []config.DbusQueryObjectArg{},
			},
			expected: []interface{}{},
		},
		{
			name: "multiple args",
			queryObject: &config.DbusQueryObject{
				Args: []config.DbusQueryObjectArg{
					{Arg: "interface"},
					{Arg: "property"},
					{Arg: "value"},
				},
			},
			expected: []interface{}{"interface", "property", "value"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.queryObject.CollectArgs()
			assert.Equal(t, tt.expected, result, "collected args should match expected")
		})
	}
}

func TestPowerSectionDbusQueryObjectDefaults(t *testing.T) {
	powerSection := &config.PowerSection{}

	err := powerSection.Validate()
	assert.NoError(t, err, "power section validation should succeed")
	assert.NotNil(t, powerSection.DbusQueryObject, "default DbusQueryObject should be set")

	expected := &config.DbusQueryObject{
		Destination: "org.freedesktop.UPower",
		Path:        "/org/freedesktop/UPower",
		Method:      "org.freedesktop.DBus.Properties.Get",
		Args: []config.DbusQueryObjectArg{
			{Arg: "org.freedesktop.UPower"},
			{Arg: "OnBattery"},
		},
	}

	assert.Equal(t, expected, powerSection.DbusQueryObject,
		"default DbusQueryObject should match expected")

	expectedCollectedArgs := []interface{}{"org.freedesktop.UPower", "OnBattery"}
	collectedArgs := powerSection.DbusQueryObject.CollectArgs()
	assert.Equal(t, expectedCollectedArgs, collectedArgs, "collected args should match expected")
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
