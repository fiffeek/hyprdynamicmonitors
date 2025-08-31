// Package matchers provides profile matching logic based on monitor conditions.
package matchers

import (
	"github.com/fiffeek/hyprdynamicmonitors/internal/config"
	"github.com/fiffeek/hyprdynamicmonitors/internal/hypr"
	"github.com/fiffeek/hyprdynamicmonitors/internal/power"
)

type Matcher struct{}

func NewMatcher() *Matcher {
	return &Matcher{}
}

func (m *Matcher) Match(cfg *config.RawConfig, connectedMonitors []*hypr.MonitorSpec,
	powerState power.PowerState,
) (bool, *config.Profile, error) {
	score := make(map[string]int)
	profiles := cfg.Profiles
	ok, fallbackProfile := m.returnNoneOrFallback(cfg)
	for name := range profiles {
		score[name] = 0
	}

	for name, profile := range profiles {
		conditions := profile.Conditions
		fullMatchScore := m.calcFullProfileScore(cfg, conditions)
		score[name] = m.scoreProfile(cfg, conditions, powerState, connectedMonitors)

		// if there is a partial match discard the config
		if fullMatchScore != score[name] {
			delete(score, name)
		}
	}

	bestScore := 0
	for _, value := range score {
		bestScore = max(bestScore, value)
	}

	// when nothing scored > 0 then no config matches
	if bestScore == 0 {
		return ok, fallbackProfile, nil
	}

	for name, profile := range profiles {
		if score[name] == bestScore {
			return true, profile, nil
		}
	}
	return ok, fallbackProfile, nil
}

func (m *Matcher) returnNoneOrFallback(cfg *config.RawConfig) (bool, *config.Profile) {
	if cfg.FallbackProfile != nil {
		return true, cfg.FallbackProfile
	}
	return false, nil
}

func (m *Matcher) scoreProfile(cfg *config.RawConfig, conditions *config.ProfileCondition,
	powerState power.PowerState, connectedMonitors []*hypr.MonitorSpec,
) int {
	profileScore := 0
	if conditions.PowerState != nil && conditions.PowerState.Value() == powerState.String() {
		profileScore += *cfg.Scoring.PowerStateMatch
	}

	for _, condition := range conditions.RequiredMonitors {
		for _, connectedMonitor := range connectedMonitors {
			if condition.Name != nil && *condition.Name == connectedMonitor.Name {
				profileScore += *cfg.Scoring.NameMatch
			}
			if condition.Description != nil && *condition.Description == connectedMonitor.Description {
				profileScore += *cfg.Scoring.DescriptionMatch
			}
		}
	}

	return profileScore
}

func (m *Matcher) calcFullProfileScore(cfg *config.RawConfig, conditions *config.ProfileCondition) int {
	fullMatchScore := 0
	if conditions.PowerState != nil {
		fullMatchScore += *cfg.Scoring.PowerStateMatch
	}

	for _, condition := range conditions.RequiredMonitors {
		if condition.Name != nil {
			fullMatchScore += *cfg.Scoring.NameMatch
		}
		if condition.Description != nil {
			fullMatchScore += *cfg.Scoring.DescriptionMatch
		}
	}

	return fullMatchScore
}
