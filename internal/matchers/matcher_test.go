package matchers_test

import (
	"testing"

	"github.com/fiffeek/hyprdynamicmonitors/internal/config"
	"github.com/fiffeek/hyprdynamicmonitors/internal/hypr"
	"github.com/fiffeek/hyprdynamicmonitors/internal/matchers"
	"github.com/fiffeek/hyprdynamicmonitors/internal/power"
	"github.com/fiffeek/hyprdynamicmonitors/internal/testutils"
	"github.com/fiffeek/hyprdynamicmonitors/internal/utils"
	"github.com/stretchr/testify/assert"
)

func TestMatcher_Match(t *testing.T) {
	tests := []struct {
		name              string
		config            *config.Config
		connectedMonitors []*hypr.MonitorSpec
		powerState        power.PowerState
		lidState          power.LidState
		expectedProfile   string
		// profile name or empty string for no match
		expectedMonitorToRule map[int]*config.RequiredMonitor
		// optional: validate monitor to rule mapping
		description string
	}{
		{
			name: "regex description",
			config: createTestConfig(t, map[string]*config.Profile{
				"regex": {
					Name: "regex",
					Conditions: &config.ProfileCondition{
						RequiredMonitors: []*config.RequiredMonitor{
							{Description: utils.StringPtr("Dell.*"), MatchDescriptionUsingRegex: utils.JustPtr(true)},
						},
					},
				},
			}).Get(),
			connectedMonitors: []*hypr.MonitorSpec{
				{Name: "eDP-1", ID: utils.IntPtr(0), Description: "Dell Inc. DELL P2422H 8LFQ514"},
			},
			powerState:      power.ACPowerState,
			expectedProfile: "regex",
			expectedMonitorToRule: map[int]*config.RequiredMonitor{
				0: {Description: utils.StringPtr("Dell.*"), MatchDescriptionUsingRegex: utils.JustPtr(
					true), MatchNameUsingRegex: utils.JustPtr(false)},
			},
			description: "Description matches by regex",
		},
		{
			name: "no regex description",
			config: createTestConfig(t, map[string]*config.Profile{
				"noregex": {
					Name: "noregex",
					Conditions: &config.ProfileCondition{
						RequiredMonitors: []*config.RequiredMonitor{
							{Description: utils.StringPtr("Dell.*")},
						},
					},
				},
			}).Get(),
			connectedMonitors: []*hypr.MonitorSpec{
				{Name: "eDP-1", ID: utils.IntPtr(0), Description: "Dell Inc. DELL P2422H 8LFQ514"},
			},
			powerState:      power.ACPowerState,
			expectedProfile: "",
			description:     "Description does not match if regex flag is not provided",
		},
		{
			name: "do not use same monitor twice",
			config: createTestConfig(t, map[string]*config.Profile{
				"dual_monitors": {
					Name: "dual_monitors",
					Conditions: &config.ProfileCondition{
						RequiredMonitors: []*config.RequiredMonitor{
							{Name: utils.StringPtr("eDP-1")},
							{Name: utils.StringPtr("eDP-1")},
						},
					},
				},
			}).Get(),
			connectedMonitors: []*hypr.MonitorSpec{
				{Name: "eDP-1", ID: utils.IntPtr(0), Description: "Built-in Display"},
			},
			powerState:            power.ACPowerState,
			expectedProfile:       "",
			expectedMonitorToRule: nil, // No match, so no mapping expected
			description:           "Required two of the 'same' monitors but only one is present",
		},
		{
			name: "exact_name_match_wins",
			config: createTestConfig(t, map[string]*config.Profile{
				"laptop_only": {
					Name: "laptop_only",
					Conditions: &config.ProfileCondition{
						RequiredMonitors: []*config.RequiredMonitor{
							{Name: utils.StringPtr("eDP-1")},
						},
					},
				},
				"dual_monitors": {
					Name: "dual_monitors",
					Conditions: &config.ProfileCondition{
						RequiredMonitors: []*config.RequiredMonitor{
							{Name: utils.StringPtr("eDP-1")},
							{Name: utils.StringPtr("DP-1")},
						},
					},
				},
			}).Get(),
			connectedMonitors: []*hypr.MonitorSpec{
				{Name: "eDP-1", ID: utils.IntPtr(0), Description: "Built-in Display"},
				{Name: "DP-1", ID: utils.IntPtr(1), Description: "External Monitor"},
			},
			powerState:      power.ACPowerState,
			expectedProfile: "dual_monitors",
			expectedMonitorToRule: map[int]*config.RequiredMonitor{
				0: {
					Name: utils.StringPtr("eDP-1"), MatchNameUsingRegex: utils.JustPtr(false),
					MatchDescriptionUsingRegex: utils.JustPtr(false),
				},
				1: {
					Name: utils.StringPtr("DP-1"), MatchNameUsingRegex: utils.JustPtr(false),
					MatchDescriptionUsingRegex: utils.JustPtr(false),
				},
			},
			description: "Higher scoring profile (2 name matches) should win over lower scoring (1 name match)",
		},
		{
			name: "description_match_works",
			config: createTestConfig(t, map[string]*config.Profile{
				"external_only": {
					Name: "external_only",
					Conditions: &config.ProfileCondition{
						RequiredMonitors: []*config.RequiredMonitor{
							{Description: utils.StringPtr("External Monitor")},
						},
					},
				},
			}).Get(),
			connectedMonitors: []*hypr.MonitorSpec{
				{Name: "DP-1", ID: utils.IntPtr(0), Description: "External Monitor"},
			},
			powerState:      power.ACPowerState,
			expectedProfile: "external_only",
			expectedMonitorToRule: map[int]*config.RequiredMonitor{
				0: {
					Description:         utils.StringPtr("External Monitor"),
					MatchNameUsingRegex: utils.JustPtr(false), MatchDescriptionUsingRegex: utils.JustPtr(false),
				},
			},
			description: "Profile should match based on description",
		},
		{
			name: "lid_state_scoring",
			config: createTestConfig(t, map[string]*config.Profile{
				"whatever": {
					Name: "whatever",
					Conditions: &config.ProfileCondition{
						RequiredMonitors: []*config.RequiredMonitor{
							{Name: utils.StringPtr("eDP-1")},
						},
					},
				},
				"lid_closed": {
					Name: "lid_closed",
					Conditions: &config.ProfileCondition{
						LidState: utils.JustPtr(config.ClosedLidStateType),
						RequiredMonitors: []*config.RequiredMonitor{
							{Name: utils.StringPtr("eDP-1")},
						},
					},
				},
				"lid_opened": {
					Name: "lid_opened",
					Conditions: &config.ProfileCondition{
						LidState: utils.JustPtr(config.OpenedLidStateType),
						RequiredMonitors: []*config.RequiredMonitor{
							{Name: utils.StringPtr("eDP-1")},
						},
					},
				},
			}).Get(),
			connectedMonitors: []*hypr.MonitorSpec{
				{Name: "eDP-1", ID: utils.IntPtr(0), Description: "Built-in Display"},
			},
			lidState:        power.ClosedLidState,
			expectedProfile: "lid_closed",
			expectedMonitorToRule: map[int]*config.RequiredMonitor{
				0: {
					Name: utils.StringPtr("eDP-1"), MatchNameUsingRegex: utils.JustPtr(false),
					MatchDescriptionUsingRegex: utils.JustPtr(false),
				},
			},
			description: "Profile with matching lid state should win (name match + power state match vs just name match)",
		},
		{
			name: "power_state_scoring",
			config: createTestConfig(t, map[string]*config.Profile{
				"battery_profile": {
					Name: "battery_profile",
					Conditions: &config.ProfileCondition{
						PowerState: utils.JustPtr(config.BAT),
						RequiredMonitors: []*config.RequiredMonitor{
							{Name: utils.StringPtr("eDP-1")},
						},
					},
				},
				"ac_profile": {
					Name: "ac_profile",
					Conditions: &config.ProfileCondition{
						PowerState: utils.JustPtr(config.AC),
						RequiredMonitors: []*config.RequiredMonitor{
							{Name: utils.StringPtr("eDP-1")},
						},
					},
				},
			}).Get(),
			connectedMonitors: []*hypr.MonitorSpec{
				{Name: "eDP-1", ID: utils.IntPtr(0), Description: "Built-in Display"},
			},
			powerState:      power.BatteryPowerState,
			expectedProfile: "battery_profile",
			expectedMonitorToRule: map[int]*config.RequiredMonitor{
				0: {
					Name: utils.StringPtr("eDP-1"), MatchNameUsingRegex: utils.JustPtr(false),
					MatchDescriptionUsingRegex: utils.JustPtr(false),
				},
			},
			description: "Profile with matching power state should win (name match + power state match vs just name match)",
		},
		{
			name: "partial_match_discarded",
			config: createTestConfig(t, map[string]*config.Profile{
				"partial_match": {
					Name: "partial_match",
					Conditions: &config.ProfileCondition{
						RequiredMonitors: []*config.RequiredMonitor{
							{Name: utils.StringPtr("eDP-1")},
							{Name: utils.StringPtr("HDMI-1")}, // This won't match
						},
					},
				},
			}).Get(),
			connectedMonitors: []*hypr.MonitorSpec{
				{Name: "eDP-1", ID: utils.IntPtr(0), Description: "Built-in Display"},
			},
			powerState:      power.ACPowerState,
			expectedProfile: "",
			description:     "Profile with partial match should be discarded (only 1 of 2 required monitors match)",
		},
		{
			name: "no_matching_profiles",
			config: createTestConfig(t, map[string]*config.Profile{
				"hdmi_only": {
					Name: "hdmi_only",
					Conditions: &config.ProfileCondition{
						RequiredMonitors: []*config.RequiredMonitor{
							{Name: utils.StringPtr("HDMI-1")},
						},
					},
				},
			}).Get(),
			connectedMonitors: []*hypr.MonitorSpec{
				{Name: "eDP-1", ID: utils.IntPtr(0), Description: "Built-in Display"},
			},
			powerState:      power.ACPowerState,
			expectedProfile: "",
			description:     "No profile should match when required monitors are not connected",
		},
		{
			name: "mixed_name_and_description_scoring",
			config: createTestConfig(t, map[string]*config.Profile{
				"mixed_high_score": {
					Name: "mixed_high_score",
					Conditions: &config.ProfileCondition{
						RequiredMonitors: []*config.RequiredMonitor{
							{Name: utils.StringPtr("eDP-1")},                   // Name match: 10 points
							{Description: utils.StringPtr("External Monitor")}, // Description match: 5 points
						},
					},
				},
				"name_only": {
					Name: "name_only",
					Conditions: &config.ProfileCondition{
						RequiredMonitors: []*config.RequiredMonitor{
							{Name: utils.StringPtr("eDP-1")}, // Name match: 10 points only
						},
					},
				},
			}).Get(),
			connectedMonitors: []*hypr.MonitorSpec{
				{Name: "eDP-1", ID: utils.IntPtr(0), Description: "Built-in Display"},
				{Name: "DP-1", ID: utils.IntPtr(1), Description: "External Monitor"},
			},
			powerState:      power.ACPowerState,
			expectedProfile: "mixed_high_score",
			expectedMonitorToRule: map[int]*config.RequiredMonitor{
				0: {
					Name: utils.StringPtr("eDP-1"), MatchNameUsingRegex: utils.JustPtr(false),
					MatchDescriptionUsingRegex: utils.JustPtr(false),
				},
				1: {
					Description:         utils.StringPtr("External Monitor"),
					MatchNameUsingRegex: utils.JustPtr(false), MatchDescriptionUsingRegex: utils.JustPtr(false),
				},
			},
			description: "Profile with mixed name+description scoring (15 points) should win over name-only (10 points)",
		},
		{
			name: "power_state_mismatch_discards_profile",
			config: createTestConfig(t, map[string]*config.Profile{
				"ac_only": {
					Name: "ac_only",
					Conditions: &config.ProfileCondition{
						PowerState: utils.JustPtr(config.AC),
						RequiredMonitors: []*config.RequiredMonitor{
							{Name: utils.StringPtr("eDP-1")},
						},
					},
				},
			}).Get(),
			connectedMonitors: []*hypr.MonitorSpec{
				{Name: "eDP-1", ID: utils.IntPtr(0), Description: "Built-in Display"},
			},
			powerState:      power.BatteryPowerState,
			expectedProfile: "",
			description:     "Profile with power state requirement should be discarded when power state doesn't match",
		},
		{
			name: "fallback_profile_when_no_match",
			config: createTestConfig(t, map[string]*config.Profile{
				"hdmi_only": {
					Name: "hdmi_only",
					Conditions: &config.ProfileCondition{
						RequiredMonitors: []*config.RequiredMonitor{
							{Name: utils.StringPtr("HDMI-1")},
						},
					},
				},
			}).WithFallbackProfile(
				&config.Profile{
					Name:              "fallback",
					IsFallbackProfile: true,
					ConfigFile:        "/tmp/fallback.conf",
					ConfigType:        &[]config.ConfigFileType{config.Static}[0],
				},
			).Get(),
			connectedMonitors: []*hypr.MonitorSpec{
				{Name: "eDP-1", ID: utils.IntPtr(0), Description: "Built-in Display"},
			},
			powerState:            power.ACPowerState,
			expectedProfile:       "fallback",
			expectedMonitorToRule: map[int]*config.RequiredMonitor{}, // Fallback profile has no monitor requirements
			description:           "Fallback profile should be used when no regular profile matches",
		},
		{
			name: "regular_profile_preferred_over_fallback",
			config: createTestConfig(t, map[string]*config.Profile{
				"laptop_only": {
					Name: "laptop_only",
					Conditions: &config.ProfileCondition{
						RequiredMonitors: []*config.RequiredMonitor{
							{Name: utils.StringPtr("eDP-1")},
						},
					},
				},
			}).WithFallbackProfile(
				&config.Profile{
					Name:              "fallback",
					IsFallbackProfile: true,
					ConfigFile:        "/tmp/fallback.conf",
					ConfigType:        &[]config.ConfigFileType{config.Static}[0],
				},
			).Get(),
			connectedMonitors: []*hypr.MonitorSpec{
				{Name: "eDP-1", ID: utils.IntPtr(0), Description: "Built-in Display"},
			},
			powerState:      power.ACPowerState,
			expectedProfile: "laptop_only",
			expectedMonitorToRule: map[int]*config.RequiredMonitor{
				0: {
					Name: utils.StringPtr("eDP-1"), MatchNameUsingRegex: utils.JustPtr(false),
					MatchDescriptionUsingRegex: utils.JustPtr(false),
				},
			},
			description: "Regular matching profile should be preferred over fallback profile",
		},
		{
			name: "no_match_no_fallback_returns_false",
			config: createTestConfig(t, map[string]*config.Profile{
				"hdmi_only": {
					Name: "hdmi_only",
					Conditions: &config.ProfileCondition{
						RequiredMonitors: []*config.RequiredMonitor{
							{Name: utils.StringPtr("HDMI-1")},
						},
					},
				},
			}).Get(),
			connectedMonitors: []*hypr.MonitorSpec{
				{Name: "eDP-1", ID: utils.IntPtr(0), Description: "Built-in Display"},
			},
			powerState:      power.ACPowerState,
			expectedProfile: "",
			description:     "Should return no match when no profile matches and no fallback is configured",
		},
		{
			name: "both_name_and_description_must_match_same_monitor",
			config: createTestConfig(t, map[string]*config.Profile{
				"strict_match": {
					Name: "strict_match",
					Conditions: &config.ProfileCondition{
						RequiredMonitors: []*config.RequiredMonitor{
							{
								Name:        utils.StringPtr("eDP-1"),
								Description: utils.StringPtr("External Monitor"),
							},
						},
					},
				},
			}).Get(),
			connectedMonitors: []*hypr.MonitorSpec{
				{Name: "eDP-1", ID: utils.IntPtr(0), Description: "Built-in Display"},
				{Name: "DP-1", ID: utils.IntPtr(1), Description: "External Monitor"},
			},
			powerState:      power.ACPowerState,
			expectedProfile: "",
			description:     "Profile should not match when name comes from one monitor and description from another",
		},
		{
			name: "last_profile_wins_when_scores_equal",
			config: createTestConfig(t, map[string]*config.Profile{
				"first_profile": {
					Name: "first_profile",
					Conditions: &config.ProfileCondition{
						RequiredMonitors: []*config.RequiredMonitor{
							{Name: utils.StringPtr("eDP-1")},
						},
					},
				},
				"second_profile": {
					Name: "second_profile",
					Conditions: &config.ProfileCondition{
						RequiredMonitors: []*config.RequiredMonitor{
							{Name: utils.StringPtr("eDP-1")},
						},
					},
				},
			}).Get(),
			connectedMonitors: []*hypr.MonitorSpec{
				{Name: "eDP-1", ID: utils.IntPtr(0), Description: "Built-in Display"},
			},
			powerState:      power.ACPowerState,
			expectedProfile: "second_profile",
			expectedMonitorToRule: map[int]*config.RequiredMonitor{
				0: {
					Name: utils.StringPtr("eDP-1"), MatchNameUsingRegex: utils.JustPtr(false),
					MatchDescriptionUsingRegex: utils.JustPtr(false),
				},
			},
			description: "When two profiles have equal scores, the last one in TOML order should win",
		},
		{
			name: "both_name_and_description_match_same_monitor",
			config: createTestConfig(t, map[string]*config.Profile{
				"full_match": {
					Name: "full_match",
					Conditions: &config.ProfileCondition{
						RequiredMonitors: []*config.RequiredMonitor{
							{
								Name:        utils.StringPtr("eDP-1"),
								Description: utils.StringPtr("Built-in Display"),
							},
						},
					},
				},
				"name_only": {
					Name: "name_only",
					Conditions: &config.ProfileCondition{
						RequiredMonitors: []*config.RequiredMonitor{
							{Name: utils.StringPtr("eDP-1")},
						},
					},
				},
			}).Get(),
			connectedMonitors: []*hypr.MonitorSpec{
				{Name: "eDP-1", ID: utils.IntPtr(0), Description: "Built-in Display"},
			},
			powerState:      power.ACPowerState,
			expectedProfile: "full_match",
			expectedMonitorToRule: map[int]*config.RequiredMonitor{
				0: {
					Name: utils.StringPtr("eDP-1"), Description: utils.StringPtr("Built-in Display"),
					MatchNameUsingRegex: utils.JustPtr(false), MatchDescriptionUsingRegex: utils.JustPtr(false),
				},
			},
			description: "Profile matching both name and description on same monitor should score higher (15 points) than name-only (10 points)",
		},
		{
			name: "best_monitor_selected_for_rule",
			config: createTestConfig(t, map[string]*config.Profile{
				"best_match": {
					Name: "best_match",
					Conditions: &config.ProfileCondition{
						RequiredMonitors: []*config.RequiredMonitor{
							// This rule could match either monitor, but should pick the one with both name and desc
							{
								Name:        utils.StringPtr("DP-1"),
								Description: utils.StringPtr("Dell Monitor"),
							},
						},
					},
				},
			}).Get(),
			connectedMonitors: []*hypr.MonitorSpec{
				{Name: "DP-1", ID: utils.IntPtr(0), Description: "Generic Display"},
				{Name: "DP-1", ID: utils.IntPtr(1), Description: "Dell Monitor"},
			},
			powerState:      power.ACPowerState,
			expectedProfile: "best_match",
			expectedMonitorToRule: map[int]*config.RequiredMonitor{
				1: {
					Name:                       utils.StringPtr("DP-1"),
					Description:                utils.StringPtr("Dell Monitor"),
					MatchDescriptionUsingRegex: utils.JustPtr(false),
					MatchNameUsingRegex:        utils.JustPtr(false),
				},
			},
			description: "When multiple monitors could match, should select the one with highest score (name+desc match over just name match)",
		},
		{
			name: "lid_state_mismatch_discards_profile",
			config: createTestConfig(t, map[string]*config.Profile{
				"lid_closed_only": {
					Name: "lid_closed_only",
					Conditions: &config.ProfileCondition{
						LidState: utils.JustPtr(config.ClosedLidStateType),
						RequiredMonitors: []*config.RequiredMonitor{
							{Name: utils.StringPtr("eDP-1")},
						},
					},
				},
			}).Get(),
			connectedMonitors: []*hypr.MonitorSpec{
				{Name: "eDP-1", ID: utils.IntPtr(0), Description: "Built-in Display"},
			},
			lidState:        power.OpenedLidState,
			expectedProfile: "",
			description:     "Profile with lid state requirement should be discarded when lid state doesn't match",
		},
		{
			name: "complex_multi_monitor_matching",
			config: createTestConfig(t, map[string]*config.Profile{
				"triple_monitor": {
					Name: "triple_monitor",
					Conditions: &config.ProfileCondition{
						RequiredMonitors: []*config.RequiredMonitor{
							{Name: utils.StringPtr("eDP-1")},
							{Description: utils.StringPtr("Dell.*"), MatchDescriptionUsingRegex: utils.JustPtr(true)},
							{Description: utils.StringPtr("LG.*"), MatchDescriptionUsingRegex: utils.JustPtr(true)},
						},
					},
				},
			}).Get(),
			connectedMonitors: []*hypr.MonitorSpec{
				{Name: "eDP-1", ID: utils.IntPtr(0), Description: "Built-in Display"},
				{Name: "DP-1", ID: utils.IntPtr(1), Description: "Dell Monitor 27inch"},
				{Name: "DP-2", ID: utils.IntPtr(2), Description: "LG UltraWide"},
			},
			powerState:      power.ACPowerState,
			expectedProfile: "triple_monitor",
			expectedMonitorToRule: map[int]*config.RequiredMonitor{
				0: {
					Name: utils.StringPtr("eDP-1"), MatchNameUsingRegex: utils.JustPtr(false),
					MatchDescriptionUsingRegex: utils.JustPtr(false),
				},
				1: {Description: utils.StringPtr("Dell.*"), MatchDescriptionUsingRegex: utils.JustPtr(
					true), MatchNameUsingRegex: utils.JustPtr(false)},
				2: {
					Description: utils.StringPtr("LG.*"), MatchDescriptionUsingRegex: utils.JustPtr(true),
					MatchNameUsingRegex: utils.JustPtr(false),
				},
			},
			description: "Should match profile requiring three different monitors with various matching criteria",
		},
		{
			name: "monitor_reuse_prevented_in_scoring_0",
			config: createTestConfig(t, map[string]*config.Profile{
				"requires_two_dell": {
					Name: "requires_two_dell",
					Conditions: &config.ProfileCondition{
						RequiredMonitors: []*config.RequiredMonitor{
							{Description: utils.StringPtr("Dell.*"), MatchDescriptionUsingRegex: utils.JustPtr(true)},
							{Description: utils.StringPtr("Dell.*"), MatchDescriptionUsingRegex: utils.JustPtr(true)},
						},
					},
				},
			}).Get(),
			connectedMonitors: []*hypr.MonitorSpec{
				{Name: "DP-1", ID: utils.IntPtr(0), Description: "Dell Monitor A"},
			},
			powerState:      power.ACPowerState,
			expectedProfile: "",
			description:     "Should not match when one monitor matches the same regex pattern",
		},
		{
			name: "monitor_reuse_prevented_in_scoring",
			config: createTestConfig(t, map[string]*config.Profile{
				"requires_two_dell": {
					Name: "requires_two_dell",
					Conditions: &config.ProfileCondition{
						RequiredMonitors: []*config.RequiredMonitor{
							{Description: utils.StringPtr("Dell.*"), MatchDescriptionUsingRegex: utils.JustPtr(true)},
							{Description: utils.StringPtr("Dell.*"), MatchDescriptionUsingRegex: utils.JustPtr(true)},
						},
					},
				},
			}).Get(),
			connectedMonitors: []*hypr.MonitorSpec{
				{Name: "DP-1", ID: utils.IntPtr(0), Description: "Dell Monitor A"},
				{Name: "DP-2", ID: utils.IntPtr(1), Description: "Dell Monitor B"},
			},
			powerState:      power.ACPowerState,
			expectedProfile: "requires_two_dell",
			expectedMonitorToRule: map[int]*config.RequiredMonitor{
				0: {Description: utils.StringPtr("Dell.*"), MatchDescriptionUsingRegex: utils.JustPtr(
					true), MatchNameUsingRegex: utils.JustPtr(false)},
				1: {Description: utils.StringPtr("Dell.*"), MatchDescriptionUsingRegex: utils.JustPtr(
					true), MatchNameUsingRegex: utils.JustPtr(false)},
			},
			description: "Should match when two different monitors both match the same regex pattern",
		},
		{
			name: "monitor_reuse_causes_partial_match",
			config: createTestConfig(t, map[string]*config.Profile{
				"requires_two_dell": {
					Name: "requires_two_dell",
					Conditions: &config.ProfileCondition{
						RequiredMonitors: []*config.RequiredMonitor{
							{Description: utils.StringPtr("Dell.*"), MatchDescriptionUsingRegex: utils.JustPtr(true)},
							{Description: utils.StringPtr("Dell.*"), MatchDescriptionUsingRegex: utils.JustPtr(true)},
						},
					},
				},
			}).Get(),
			connectedMonitors: []*hypr.MonitorSpec{
				{Name: "DP-1", ID: utils.IntPtr(0), Description: "Dell Monitor A"},
				{Name: "DP-2", ID: utils.IntPtr(1), Description: "LG Monitor"},
			},
			powerState:      power.ACPowerState,
			expectedProfile: "",
			description:     "Should not match when only one monitor matches a rule that requires two matches",
		},
		{
			name: "combined_power_and_lid_state_matching",
			config: createTestConfig(t, map[string]*config.Profile{
				"battery_lid_closed": {
					Name: "battery_lid_closed",
					Conditions: &config.ProfileCondition{
						PowerState: utils.JustPtr(config.BAT),
						LidState:   utils.JustPtr(config.ClosedLidStateType),
						RequiredMonitors: []*config.RequiredMonitor{
							{Name: utils.StringPtr("DP-1")},
						},
					},
				},
				"just_external": {
					Name: "just_external",
					Conditions: &config.ProfileCondition{
						RequiredMonitors: []*config.RequiredMonitor{
							{Name: utils.StringPtr("DP-1")},
						},
					},
				},
			}).Get(),
			connectedMonitors: []*hypr.MonitorSpec{
				{Name: "DP-1", ID: utils.IntPtr(0), Description: "External Monitor"},
			},
			powerState:      power.BatteryPowerState,
			lidState:        power.ClosedLidState,
			expectedProfile: "battery_lid_closed",
			expectedMonitorToRule: map[int]*config.RequiredMonitor{
				0: {
					Name: utils.StringPtr("DP-1"), MatchNameUsingRegex: utils.JustPtr(false),
					MatchDescriptionUsingRegex: utils.JustPtr(false),
				},
			},
			description: "Profile matching both power state and lid state should score higher than just monitor match",
		},
		{
			name: "regex_description_multiple_matches_picks_best",
			config: createTestConfig(t, map[string]*config.Profile{
				"dell_preference": {
					Name: "dell_preference",
					Conditions: &config.ProfileCondition{
						RequiredMonitors: []*config.RequiredMonitor{
							{
								Name:                       utils.StringPtr("DP-1"),
								Description:                utils.StringPtr("Dell.*"),
								MatchDescriptionUsingRegex: utils.JustPtr(true),
							},
						},
					},
				},
			}).Get(),
			connectedMonitors: []*hypr.MonitorSpec{
				{Name: "DP-1", ID: utils.IntPtr(0), Description: "Samsung Monitor"},
				{Name: "DP-1", ID: utils.IntPtr(1), Description: "Dell Monitor"},
			},
			powerState:      power.ACPowerState,
			expectedProfile: "dell_preference",
			description:     "When multiple monitors have same name, should pick the one that also matches description regex",
		},
		{
			name: "empty_monitors_no_match",
			config: createTestConfig(t, map[string]*config.Profile{
				"any_monitor": {
					Name: "any_monitor",
					Conditions: &config.ProfileCondition{
						RequiredMonitors: []*config.RequiredMonitor{
							{Name: utils.StringPtr("eDP-1")},
						},
					},
				},
			}).Get(),
			connectedMonitors: []*hypr.MonitorSpec{},
			powerState:        power.ACPowerState,
			expectedProfile:   "",
			description:       "Should not match any profile when no monitors are connected",
		},
		{
			name: "regex_name",
			config: createTestConfig(t, map[string]*config.Profile{
				"regex": {
					Name: "regex",
					Conditions: &config.ProfileCondition{
						RequiredMonitors: []*config.RequiredMonitor{
							{Name: utils.StringPtr("eDP-.*"), MatchNameUsingRegex: utils.JustPtr(true)},
						},
					},
				},
			}).Get(),
			connectedMonitors: []*hypr.MonitorSpec{
				{Name: "eDP-1", ID: utils.IntPtr(0), Description: "Built-in Display"},
			},
			powerState:      power.ACPowerState,
			expectedProfile: "regex",
			description:     "Name matches by regex",
		},
		{
			name: "no_regex_name",
			config: createTestConfig(t, map[string]*config.Profile{
				"noregex": {
					Name: "noregex",
					Conditions: &config.ProfileCondition{
						RequiredMonitors: []*config.RequiredMonitor{
							{Name: utils.StringPtr("eDP-.*")},
						},
					},
				},
			}).Get(),
			connectedMonitors: []*hypr.MonitorSpec{
				{Name: "eDP-1", ID: utils.IntPtr(0), Description: "Built-in Display"},
			},
			powerState:      power.ACPowerState,
			expectedProfile: "",
			description:     "Name does not match if regex flag is not provided",
		},
		{
			name: "regex_name_and_description_combined",
			config: createTestConfig(t, map[string]*config.Profile{
				"regex_both": {
					Name: "regex_both",
					Conditions: &config.ProfileCondition{
						RequiredMonitors: []*config.RequiredMonitor{
							{
								Name:                       utils.StringPtr("DP-[0-9]+"),
								Description:                utils.StringPtr("Dell.*"),
								MatchNameUsingRegex:        utils.JustPtr(true),
								MatchDescriptionUsingRegex: utils.JustPtr(true),
							},
						},
					},
				},
			}).Get(),
			connectedMonitors: []*hypr.MonitorSpec{
				{Name: "DP-1", ID: utils.IntPtr(0), Description: "Dell Monitor 27inch"},
			},
			powerState:      power.ACPowerState,
			expectedProfile: "regex_both",
			description:     "Both name and description match by regex on same monitor",
		},
		{
			name: "regex_name_multiple_matches",
			config: createTestConfig(t, map[string]*config.Profile{
				"multi_dp": {
					Name: "multi_dp",
					Conditions: &config.ProfileCondition{
						RequiredMonitors: []*config.RequiredMonitor{
							{Name: utils.StringPtr("DP-[0-9]+"), MatchNameUsingRegex: utils.JustPtr(true)},
							{Name: utils.StringPtr("DP-[0-9]+"), MatchNameUsingRegex: utils.JustPtr(true)},
						},
					},
				},
			}).Get(),
			connectedMonitors: []*hypr.MonitorSpec{
				{Name: "DP-1", ID: utils.IntPtr(0), Description: "Monitor A"},
				{Name: "DP-2", ID: utils.IntPtr(1), Description: "Monitor B"},
			},
			powerState:      power.ACPowerState,
			expectedProfile: "multi_dp",
			expectedMonitorToRule: map[int]*config.RequiredMonitor{
				0: {
					Name:                       utils.StringPtr("DP-[0-9]+"),
					MatchNameUsingRegex:        utils.JustPtr(true),
					MatchDescriptionUsingRegex: utils.JustPtr(false),
				},
				1: {
					Name:                       utils.StringPtr("DP-[0-9]+"),
					MatchNameUsingRegex:        utils.JustPtr(true),
					MatchDescriptionUsingRegex: utils.JustPtr(false),
				},
			},
			description: "Should match when two different monitors both match the same name regex pattern",
		},
		{
			name: "regex_name_partial_match_only_one",
			config: createTestConfig(t, map[string]*config.Profile{
				"multi_dp": {
					Name: "multi_dp",
					Conditions: &config.ProfileCondition{
						RequiredMonitors: []*config.RequiredMonitor{
							{Name: utils.StringPtr("DP-[0-9]+"), MatchNameUsingRegex: utils.JustPtr(true)},
							{Name: utils.StringPtr("DP-[0-9]+"), MatchNameUsingRegex: utils.JustPtr(true)},
						},
					},
				},
			}).Get(),
			connectedMonitors: []*hypr.MonitorSpec{
				{Name: "DP-1", ID: utils.IntPtr(0), Description: "Monitor A"},
				{Name: "HDMI-1", ID: utils.IntPtr(1), Description: "Monitor B"},
			},
			powerState:      power.ACPowerState,
			expectedProfile: "",
			description:     "Should not match when only one monitor matches a name regex that requires two matches",
		},
		{
			name: "regex_name_picks_best_match",
			config: createTestConfig(t, map[string]*config.Profile{
				"best_dp": {
					Name: "best_dp",
					Conditions: &config.ProfileCondition{
						RequiredMonitors: []*config.RequiredMonitor{
							{
								Name:                utils.StringPtr("DP-[0-9]+"),
								Description:         utils.StringPtr("Dell Monitor"),
								MatchNameUsingRegex: utils.JustPtr(true),
							},
						},
					},
				},
			}).Get(),
			connectedMonitors: []*hypr.MonitorSpec{
				{Name: "DP-1", ID: utils.IntPtr(0), Description: "Generic Monitor"},
				{Name: "DP-2", ID: utils.IntPtr(1), Description: "Dell Monitor"},
			},
			powerState:      power.ACPowerState,
			expectedProfile: "best_dp",
			expectedMonitorToRule: map[int]*config.RequiredMonitor{
				1: {
					Name:                       utils.StringPtr("DP-[0-9]+"),
					Description:                utils.StringPtr("Dell Monitor"),
					MatchNameUsingRegex:        utils.JustPtr(true),
					MatchDescriptionUsingRegex: utils.JustPtr(false),
				},
			},
			description: "When multiple monitors match name regex, should pick the one that also matches description",
		},
		{
			name: "mixed_regex_and_exact_name_matching",
			config: createTestConfig(t, map[string]*config.Profile{
				"mixed": {
					Name: "mixed",
					Conditions: &config.ProfileCondition{
						RequiredMonitors: []*config.RequiredMonitor{
							{Name: utils.StringPtr("eDP-1")},                                               // Exact match
							{Name: utils.StringPtr("DP-[0-9]+"), MatchNameUsingRegex: utils.JustPtr(true)}, // Regex match
						},
					},
				},
			}).Get(),
			connectedMonitors: []*hypr.MonitorSpec{
				{Name: "eDP-1", ID: utils.IntPtr(0), Description: "Built-in Display"},
				{Name: "DP-1", ID: utils.IntPtr(1), Description: "External Monitor"},
			},
			powerState:      power.ACPowerState,
			expectedProfile: "mixed",
			description:     "Should match profile with mix of exact and regex name matching",
		},
		{
			name: "regex_name_with_exact_description",
			config: createTestConfig(t, map[string]*config.Profile{
				"hybrid": {
					Name: "hybrid",
					Conditions: &config.ProfileCondition{
						RequiredMonitors: []*config.RequiredMonitor{
							{
								Name:                utils.StringPtr("DP-[0-9]+"),
								Description:         utils.StringPtr("Dell Monitor"),
								MatchNameUsingRegex: utils.JustPtr(true),
							},
						},
					},
				},
			}).Get(),
			connectedMonitors: []*hypr.MonitorSpec{
				{Name: "DP-1", ID: utils.IntPtr(0), Description: "Dell Monitor"},
			},
			powerState:      power.ACPowerState,
			expectedProfile: "hybrid",
			description:     "Should match when name is regex but description is exact match",
		},
		{
			name: "complex_single_monitor_profiles_most_specific_wins",
			config: createTestConfig(t, map[string]*config.Profile{
				"name_only": {
					Name: "name_only",
					Conditions: &config.ProfileCondition{
						RequiredMonitors: []*config.RequiredMonitor{
							{Name: utils.StringPtr("eDP-1")},
						},
					},
				},
				"desc_only": {
					Name: "desc_only",
					Conditions: &config.ProfileCondition{
						RequiredMonitors: []*config.RequiredMonitor{
							{Description: utils.StringPtr("Built-in Display")},
						},
					},
				},
				"name_and_desc": {
					Name: "name_and_desc",
					Conditions: &config.ProfileCondition{
						RequiredMonitors: []*config.RequiredMonitor{
							{
								Name:        utils.StringPtr("eDP-1"),
								Description: utils.StringPtr("Built-in Display"),
							},
						},
					},
				},
			}).Get(),
			connectedMonitors: []*hypr.MonitorSpec{
				{Name: "eDP-1", ID: utils.IntPtr(0), Description: "Built-in Display"},
				{Name: "DP-1", ID: utils.IntPtr(1), Description: "External Monitor"},
			},
			powerState:      power.ACPowerState,
			expectedProfile: "name_and_desc",
			description:     "With two monitors, most specific single-monitor profile (name+desc=15pts) should win over less specific (name=10pts, desc=5pts)",
		},
		{
			name: "complex_mixed_single_and_dual_monitor_profiles",
			config: createTestConfig(t, map[string]*config.Profile{
				"single_name": {
					Name: "single_name",
					Conditions: &config.ProfileCondition{
						RequiredMonitors: []*config.RequiredMonitor{
							{Name: utils.StringPtr("eDP-1")},
						},
					},
				},
				"single_name_desc": {
					Name: "single_name_desc",
					Conditions: &config.ProfileCondition{
						RequiredMonitors: []*config.RequiredMonitor{
							{
								Name:        utils.StringPtr("eDP-1"),
								Description: utils.StringPtr("Built-in Display"),
							},
						},
					},
				},
				"dual_names": {
					Name: "dual_names",
					Conditions: &config.ProfileCondition{
						RequiredMonitors: []*config.RequiredMonitor{
							{Name: utils.StringPtr("eDP-1")},
							{Name: utils.StringPtr("DP-1")},
						},
					},
				},
				"dual_names_descs": {
					Name: "dual_names_descs",
					Conditions: &config.ProfileCondition{
						RequiredMonitors: []*config.RequiredMonitor{
							{
								Name:        utils.StringPtr("eDP-1"),
								Description: utils.StringPtr("Built-in Display"),
								MonitorTag:  utils.JustPtr("edp"),
							},
							{
								Name:        utils.StringPtr("DP-1"),
								Description: utils.StringPtr("External Monitor"),
								MonitorTag:  utils.JustPtr("dp"),
							},
						},
					},
				},
			}).Get(),
			connectedMonitors: []*hypr.MonitorSpec{
				{Name: "eDP-1", ID: utils.IntPtr(0), Description: "Built-in Display"},
				{Name: "DP-1", ID: utils.IntPtr(1), Description: "External Monitor"},
			},
			powerState:      power.ACPowerState,
			expectedProfile: "dual_names_descs",
			expectedMonitorToRule: map[int]*config.RequiredMonitor{
				0: {
					Name:                       utils.StringPtr("eDP-1"),
					Description:                utils.StringPtr("Built-in Display"),
					MatchDescriptionUsingRegex: utils.JustPtr(false),
					MatchNameUsingRegex:        utils.JustPtr(false),
					MonitorTag:                 utils.JustPtr("edp"),
				},
				1: {
					Name:                       utils.StringPtr("DP-1"),
					Description:                utils.StringPtr("External Monitor"),
					MatchDescriptionUsingRegex: utils.JustPtr(false),
					MatchNameUsingRegex:        utils.JustPtr(false),
					MonitorTag:                 utils.JustPtr("dp"),
				},
			},
			description: "Dual monitor profile with names+descs (30pts) should win over single monitor profiles (10-15pts) and dual names only (20pts)",
		},
		{
			name: "complex_single_vs_dual_single_wins_when_one_monitor",
			config: createTestConfig(t, map[string]*config.Profile{
				"single_full": {
					Name: "single_full",
					Conditions: &config.ProfileCondition{
						RequiredMonitors: []*config.RequiredMonitor{
							{
								Name:        utils.StringPtr("eDP-1"),
								Description: utils.StringPtr("Built-in Display"),
							},
						},
					},
				},
				"dual_monitors": {
					Name: "dual_monitors",
					Conditions: &config.ProfileCondition{
						RequiredMonitors: []*config.RequiredMonitor{
							{Name: utils.StringPtr("eDP-1")},
							{Name: utils.StringPtr("DP-1")},
						},
					},
				},
			}).Get(),
			connectedMonitors: []*hypr.MonitorSpec{
				{Name: "eDP-1", ID: utils.IntPtr(0), Description: "Built-in Display"},
			},
			powerState:      power.ACPowerState,
			expectedProfile: "single_full",
			description:     "Single monitor profile should match when only one monitor connected, dual monitor profile gets partial match and is discarded",
		},
		{
			name: "complex_profiles_with_regex_vs_exact",
			config: createTestConfig(t, map[string]*config.Profile{
				"exact_match": {
					Name: "exact_match",
					Conditions: &config.ProfileCondition{
						RequiredMonitors: []*config.RequiredMonitor{
							{Name: utils.StringPtr("DP-1")},
							{Name: utils.StringPtr("DP-2")},
						},
					},
				},
				"regex_match": {
					Name: "regex_match",
					Conditions: &config.ProfileCondition{
						RequiredMonitors: []*config.RequiredMonitor{
							{Name: utils.StringPtr("DP-[0-9]+"), MatchNameUsingRegex: utils.JustPtr(true)},
							{Name: utils.StringPtr("DP-[0-9]+"), MatchNameUsingRegex: utils.JustPtr(true)},
						},
					},
				},
				"regex_with_desc": {
					Name: "regex_with_desc",
					Conditions: &config.ProfileCondition{
						RequiredMonitors: []*config.RequiredMonitor{
							{
								Name:                utils.StringPtr("DP-[0-9]+"),
								Description:         utils.StringPtr("Monitor A"),
								MatchNameUsingRegex: utils.JustPtr(true),
							},
							{
								Name:                utils.StringPtr("DP-[0-9]+"),
								Description:         utils.StringPtr("Monitor B"),
								MatchNameUsingRegex: utils.JustPtr(true),
							},
						},
					},
				},
			}).Get(),
			connectedMonitors: []*hypr.MonitorSpec{
				{Name: "DP-1", ID: utils.IntPtr(0), Description: "Monitor A"},
				{Name: "DP-2", ID: utils.IntPtr(1), Description: "Monitor B"},
			},
			powerState:      power.ACPowerState,
			expectedProfile: "regex_with_desc",
			expectedMonitorToRule: map[int]*config.RequiredMonitor{
				0: {
					Name:                       utils.StringPtr("DP-[0-9]+"),
					Description:                utils.StringPtr("Monitor A"),
					MatchNameUsingRegex:        utils.JustPtr(true),
					MatchDescriptionUsingRegex: utils.JustPtr(false),
				},
				1: {
					Name:                       utils.StringPtr("DP-[0-9]+"),
					Description:                utils.StringPtr("Monitor B"),
					MatchNameUsingRegex:        utils.JustPtr(true),
					MatchDescriptionUsingRegex: utils.JustPtr(false),
				},
			},
			description: "Regex profile with descriptions (30pts) should win over exact match (20pts) and regex without desc (20pts)",
		},
		{
			name: "complex_with_power_state_tiebreaker",
			config: createTestConfig(t, map[string]*config.Profile{
				"dual_no_power": {
					Name: "dual_no_power",
					Conditions: &config.ProfileCondition{
						RequiredMonitors: []*config.RequiredMonitor{
							{Name: utils.StringPtr("eDP-1")},
							{Name: utils.StringPtr("DP-1")},
						},
					},
				},
				"dual_with_power": {
					Name: "dual_with_power",
					Conditions: &config.ProfileCondition{
						PowerState: utils.JustPtr(config.BAT),
						RequiredMonitors: []*config.RequiredMonitor{
							{Name: utils.StringPtr("eDP-1")},
							{Name: utils.StringPtr("DP-1")},
						},
					},
				},
			}).Get(),
			connectedMonitors: []*hypr.MonitorSpec{
				{Name: "eDP-1", ID: utils.IntPtr(0), Description: "Built-in Display"},
				{Name: "DP-1", ID: utils.IntPtr(1), Description: "External Monitor"},
			},
			powerState:      power.BatteryPowerState,
			expectedProfile: "dual_with_power",
			description:     "Profile matching power state (23pts) should win over same monitors without power state (20pts)",
		},
		{
			name: "complex_three_monitors_partial_matches",
			config: createTestConfig(t, map[string]*config.Profile{
				"all_three": {
					Name: "all_three",
					Conditions: &config.ProfileCondition{
						RequiredMonitors: []*config.RequiredMonitor{
							{Name: utils.StringPtr("eDP-1")},
							{Name: utils.StringPtr("DP-1")},
							{Name: utils.StringPtr("DP-2")},
						},
					},
				},
				"first_two": {
					Name: "first_two",
					Conditions: &config.ProfileCondition{
						RequiredMonitors: []*config.RequiredMonitor{
							{Name: utils.StringPtr("eDP-1")},
							{Name: utils.StringPtr("DP-1")},
						},
					},
				},
				"last_two": {
					Name: "last_two",
					Conditions: &config.ProfileCondition{
						RequiredMonitors: []*config.RequiredMonitor{
							{Name: utils.StringPtr("DP-1")},
							{Name: utils.StringPtr("DP-2")},
						},
					},
				},
			}).Get(),
			connectedMonitors: []*hypr.MonitorSpec{
				{Name: "eDP-1", ID: utils.IntPtr(0), Description: "Built-in"},
				{Name: "DP-1", ID: utils.IntPtr(1), Description: "Monitor 1"},
				{Name: "DP-2", ID: utils.IntPtr(2), Description: "Monitor 2"},
			},
			powerState:      power.ACPowerState,
			expectedProfile: "all_three",
			description:     "Profile matching all three monitors (30pts) should win over profiles matching only two (20pts each)",
		},
		{
			name: "complex_overlapping_regex_patterns",
			config: createTestConfig(t, map[string]*config.Profile{
				"broad_regex": {
					Name: "broad_regex",
					Conditions: &config.ProfileCondition{
						RequiredMonitors: []*config.RequiredMonitor{
							{Name: utils.StringPtr(".*"), MatchNameUsingRegex: utils.JustPtr(true)},
							{Name: utils.StringPtr(".*"), MatchNameUsingRegex: utils.JustPtr(true)},
						},
					},
				},
				"specific_regex": {
					Name: "specific_regex",
					Conditions: &config.ProfileCondition{
						RequiredMonitors: []*config.RequiredMonitor{
							{Name: utils.StringPtr("eDP-.*"), MatchNameUsingRegex: utils.JustPtr(true)},
							{Name: utils.StringPtr("DP-.*"), MatchNameUsingRegex: utils.JustPtr(true)},
						},
					},
				},
				"specific_with_desc": {
					Name: "specific_with_desc",
					Conditions: &config.ProfileCondition{
						RequiredMonitors: []*config.RequiredMonitor{
							{
								Name:                utils.StringPtr("eDP-.*"),
								Description:         utils.StringPtr("Built-in Display"),
								MatchNameUsingRegex: utils.JustPtr(true),
							},
							{Name: utils.StringPtr("DP-.*"), MatchNameUsingRegex: utils.JustPtr(true)},
						},
					},
				},
			}).Get(),
			connectedMonitors: []*hypr.MonitorSpec{
				{Name: "eDP-1", ID: utils.IntPtr(0), Description: "Built-in Display"},
				{Name: "DP-1", ID: utils.IntPtr(1), Description: "External"},
			},
			powerState:      power.ACPowerState,
			expectedProfile: "specific_with_desc",
			description:     "Most specific regex with description (25pts) should win over less specific regex patterns (20pts each)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matcher := matchers.NewMatcher()

			found, result, err := matcher.Match(tt.config.Get(), tt.connectedMonitors, tt.powerState, tt.lidState)
			if err != nil {
				t.Fatalf("Match returned unexpected error: %v", err)
			}

			if tt.expectedProfile == "" {
				if result != nil || found {
					t.Errorf("Expected no profile match, but got profile: %s", result.Profile.Name)
				}
			} else {
				if result == nil || !found {
					t.Errorf("Expected profile %s, but got no match", tt.expectedProfile)
				} else if result.Profile.Name != tt.expectedProfile {
					t.Errorf("Expected profile %s, but got %s",
						tt.expectedProfile, result.Profile.Name)
				}

				// Validate MonitorToRule mapping if expected values are provided
				if tt.expectedMonitorToRule != nil {
					assert.Equal(t, normalizeMonitorToRule(tt.expectedMonitorToRule),
						normalizeMonitorToRule(result.MonitorToRule),
						"MonitorToRule mapping should match expected")
				}
			}

			t.Logf("Test case: %s - %s", tt.name, tt.description)
		})
	}
}

// Helper function to create test config with default scoring
func createTestConfig(t *testing.T, profiles map[string]*config.Profile) *testutils.TestConfig {
	return testutils.NewTestConfig(t).WithProfiles(profiles).WithScoring(&config.ScoringSection{
		NameMatch:        utils.IntPtr(10),
		DescriptionMatch: utils.IntPtr(5),
		PowerStateMatch:  utils.IntPtr(3),
		LidStateMatch:    utils.IntPtr(3),
	})
}

// Helper to normalize RequiredMonitor for comparison (strips compiled regex)
func normalizeRequiredMonitor(rm *config.RequiredMonitor) *config.RequiredMonitor {
	if rm == nil {
		return nil
	}
	return &config.RequiredMonitor{
		Name:                       rm.Name,
		Description:                rm.Description,
		MonitorTag:                 rm.MonitorTag,
		MatchDescriptionUsingRegex: rm.MatchDescriptionUsingRegex,
		MatchNameUsingRegex:        rm.MatchNameUsingRegex,
		// Intentionally omit DescriptionRegex and NameRegex
	}
}

// Helper to normalize MonitorToRule map for comparison
func normalizeMonitorToRule(m map[int]*config.RequiredMonitor) map[int]*config.RequiredMonitor {
	if m == nil {
		return nil
	}
	normalized := make(map[int]*config.RequiredMonitor)
	for k, v := range m {
		normalized[k] = normalizeRequiredMonitor(v)
	}
	return normalized
}
