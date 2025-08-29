// Package matchers provides profile matching logic based on monitor conditions.
package matchers

import (
	"github.com/fiffeek/hyprdynamicmonitors/internal/config"
	"github.com/fiffeek/hyprdynamicmonitors/internal/hypr"
	"github.com/fiffeek/hyprdynamicmonitors/internal/power"
)

type Matcher struct {
	cfg *config.Config
}

func NewMatcher(cfg *config.Config) *Matcher {
	return &Matcher{
		cfg: cfg,
	}
}

func (m *Matcher) Match(connectedMonitors []*hypr.MonitorSpec, powerState power.PowerState) (bool, *config.Profile, error) {
	score := make(map[string]int)
	for name := range m.cfg.Get().Profiles {
		score[name] = 0
	}

	for name, profile := range m.cfg.Get().Profiles {
		conditions := profile.Conditions
		fullMatchScore := m.calcFullProfileScore(conditions)
		score[name] = m.scoreProfile(conditions, powerState, connectedMonitors)

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
		return false, nil, nil
	}

	for name, profile := range m.cfg.Get().Profiles {
		if score[name] == bestScore {
			return true, profile, nil
		}
	}
	return false, nil, nil
}

func (m *Matcher) scoreProfile(conditions config.ProfileCondition, powerState power.PowerState, connectedMonitors []*hypr.MonitorSpec) int {
	profileScore := 0
	if conditions.PowerState != nil && conditions.PowerState.Value() == powerState.String() {
		profileScore += *m.cfg.Get().Scoring.PowerStateMatch
	}

	for _, condition := range conditions.RequiredMonitors {
		for _, connectedMonitor := range connectedMonitors {
			if condition.Name != nil && *condition.Name == connectedMonitor.Name {
				profileScore += *m.cfg.Get().Scoring.NameMatch
			}
			if condition.Description != nil && *condition.Description == connectedMonitor.Description {
				profileScore += *m.cfg.Get().Scoring.DescriptionMatch
			}
		}
	}

	return profileScore
}

func (m *Matcher) calcFullProfileScore(conditions config.ProfileCondition) int {
	fullMatchScore := 0
	if conditions.PowerState != nil {
		fullMatchScore += *m.cfg.Get().Scoring.PowerStateMatch
	}

	for _, condition := range conditions.RequiredMonitors {
		if condition.Name != nil {
			fullMatchScore += *m.cfg.Get().Scoring.NameMatch
		}
		if condition.Description != nil {
			fullMatchScore += *m.cfg.Get().Scoring.DescriptionMatch
		}
	}

	return fullMatchScore
}
