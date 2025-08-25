package matchers

import (
	"testing"

	"github.com/fiffeek/hyprdynamicmonitors/internal/config"
	"github.com/fiffeek/hyprdynamicmonitors/internal/detectors"
	"github.com/fiffeek/hyprdynamicmonitors/internal/hypr"
	"github.com/fiffeek/hyprdynamicmonitors/internal/utils"
)

func TestMatcher_Match(t *testing.T) {
	tests := []struct {
		name              string
		config            *config.Config
		connectedMonitors []*hypr.MonitorSpec
		powerState        detectors.PowerState
		expectedProfile   string // profile name or empty string for no match
		description       string
	}{
		{
			name: "exact_name_match_wins",
			config: createTestConfig(map[string]*config.Profile{
				"laptop_only": {
					Name: "laptop_only",
					Conditions: config.ProfileCondition{
						RequiredMonitors: []*config.RequiredMonitor{
							{Name: utils.StringPtr("eDP-1")},
						},
					},
				},
				"dual_monitors": {
					Name: "dual_monitors",
					Conditions: config.ProfileCondition{
						RequiredMonitors: []*config.RequiredMonitor{
							{Name: utils.StringPtr("eDP-1")},
							{Name: utils.StringPtr("DP-1")},
						},
					},
				},
			}),
			connectedMonitors: []*hypr.MonitorSpec{
				{Name: "eDP-1", ID: utils.IntPtr(0), Description: "Built-in Display"},
				{Name: "DP-1", ID: utils.IntPtr(1), Description: "External Monitor"},
			},
			powerState:      detectors.ACPower,
			expectedProfile: "dual_monitors",
			description:     "Higher scoring profile (2 name matches) should win over lower scoring (1 name match)",
		},
		{
			name: "description_match_works",
			config: createTestConfig(map[string]*config.Profile{
				"external_only": {
					Name: "external_only",
					Conditions: config.ProfileCondition{
						RequiredMonitors: []*config.RequiredMonitor{
							{Description: utils.StringPtr("External Monitor")},
						},
					},
				},
			}),
			connectedMonitors: []*hypr.MonitorSpec{
				{Name: "DP-1", ID: utils.IntPtr(0), Description: "External Monitor"},
			},
			powerState:      detectors.ACPower,
			expectedProfile: "external_only",
			description:     "Profile should match based on description",
		},
		{
			name: "power_state_scoring",
			config: createTestConfig(map[string]*config.Profile{
				"battery_profile": {
					Name: "battery_profile",
					Conditions: config.ProfileCondition{
						PowerState: powerStateTypePtr(config.BAT),
						RequiredMonitors: []*config.RequiredMonitor{
							{Name: utils.StringPtr("eDP-1")},
						},
					},
				},
				"ac_profile": {
					Name: "ac_profile",
					Conditions: config.ProfileCondition{
						PowerState: powerStateTypePtr(config.AC),
						RequiredMonitors: []*config.RequiredMonitor{
							{Name: utils.StringPtr("eDP-1")},
						},
					},
				},
			}),
			connectedMonitors: []*hypr.MonitorSpec{
				{Name: "eDP-1", ID: utils.IntPtr(0), Description: "Built-in Display"},
			},
			powerState:      detectors.Battery,
			expectedProfile: "battery_profile",
			description:     "Profile with matching power state should win (name match + power state match vs just name match)",
		},
		{
			name: "partial_match_discarded",
			config: createTestConfig(map[string]*config.Profile{
				"partial_match": {
					Name: "partial_match",
					Conditions: config.ProfileCondition{
						RequiredMonitors: []*config.RequiredMonitor{
							{Name: utils.StringPtr("eDP-1")},
							{Name: utils.StringPtr("HDMI-1")}, // This won't match
						},
					},
				},
			}),
			connectedMonitors: []*hypr.MonitorSpec{
				{Name: "eDP-1", ID: utils.IntPtr(0), Description: "Built-in Display"},
			},
			powerState:      detectors.ACPower,
			expectedProfile: "",
			description:     "Profile with partial match should be discarded (only 1 of 2 required monitors match)",
		},
		{
			name: "no_matching_profiles",
			config: createTestConfig(map[string]*config.Profile{
				"hdmi_only": {
					Name: "hdmi_only",
					Conditions: config.ProfileCondition{
						RequiredMonitors: []*config.RequiredMonitor{
							{Name: utils.StringPtr("HDMI-1")},
						},
					},
				},
			}),
			connectedMonitors: []*hypr.MonitorSpec{
				{Name: "eDP-1", ID: utils.IntPtr(0), Description: "Built-in Display"},
			},
			powerState:      detectors.ACPower,
			expectedProfile: "",
			description:     "No profile should match when required monitors are not connected",
		},
		{
			name: "mixed_name_and_description_scoring",
			config: createTestConfig(map[string]*config.Profile{
				"mixed_high_score": {
					Name: "mixed_high_score",
					Conditions: config.ProfileCondition{
						RequiredMonitors: []*config.RequiredMonitor{
							{Name: utils.StringPtr("eDP-1")},                   // Name match: 10 points
							{Description: utils.StringPtr("External Monitor")}, // Description match: 5 points
						},
					},
				},
				"name_only": {
					Name: "name_only",
					Conditions: config.ProfileCondition{
						RequiredMonitors: []*config.RequiredMonitor{
							{Name: utils.StringPtr("eDP-1")}, // Name match: 10 points only
						},
					},
				},
			}),
			connectedMonitors: []*hypr.MonitorSpec{
				{Name: "eDP-1", ID: utils.IntPtr(0), Description: "Built-in Display"},
				{Name: "DP-1", ID: utils.IntPtr(1), Description: "External Monitor"},
			},
			powerState:      detectors.ACPower,
			expectedProfile: "mixed_high_score",
			description:     "Profile with mixed name+description scoring (15 points) should win over name-only (10 points)",
		},
		{
			name: "power_state_mismatch_discards_profile",
			config: createTestConfig(map[string]*config.Profile{
				"ac_only": {
					Name: "ac_only",
					Conditions: config.ProfileCondition{
						PowerState: powerStateTypePtr(config.AC),
						RequiredMonitors: []*config.RequiredMonitor{
							{Name: utils.StringPtr("eDP-1")},
						},
					},
				},
			}),
			connectedMonitors: []*hypr.MonitorSpec{
				{Name: "eDP-1", ID: utils.IntPtr(0), Description: "Built-in Display"},
			},
			powerState:      detectors.Battery,
			expectedProfile: "",
			description:     "Profile with power state requirement should be discarded when power state doesn't match",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matcher := NewMatcher(tt.config)

			result, err := matcher.Match(tt.connectedMonitors, tt.powerState)
			if err != nil {
				t.Fatalf("Match returned unexpected error: %v", err)
			}

			if tt.expectedProfile == "" {
				if result != nil {
					t.Errorf("Expected no profile match, but got profile: %s", result.Name)
				}
			} else {
				if result == nil {
					t.Errorf("Expected profile %s, but got no match", tt.expectedProfile)
				} else if result.Name != tt.expectedProfile {
					t.Errorf("Expected profile %s, but got %s", tt.expectedProfile, result.Name)
				}
			}

			t.Logf("Test case: %s - %s", tt.name, tt.description)
		})
	}
}

// Helper function to create test config with default scoring
func createTestConfig(profiles map[string]*config.Profile) *config.Config {
	return &config.Config{
		Profiles: profiles,
		Scoring: &config.ScoringSection{
			NameMatch:        utils.IntPtr(10),
			DescriptionMatch: utils.IntPtr(5),
			PowerStateMatch:  utils.IntPtr(3),
		},
	}
}

// Helper function to create PowerStateType pointer
func powerStateTypePtr(p config.PowerStateType) *config.PowerStateType {
	return &p
}
