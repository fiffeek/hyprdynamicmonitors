package matchers

import (
	"github.com/fiffeek/hyprdynamicmonitors/internal/config"
	"github.com/fiffeek/hyprdynamicmonitors/internal/detectors"
	"github.com/fiffeek/hyprdynamicmonitors/internal/hypr"
)

type Matcher struct {
	cfg     *config.Config
	verbose bool
}

func NewMatcher(cfg *config.Config, verbose bool) *Matcher {
	return &Matcher{
		cfg,
		verbose,
	}
}

func (m *Matcher) Match(connectedMonitors []*hypr.MonitorSpec, powerState detectors.PowerState) (*config.Profile, error) {
	score := make(map[string]int)
	for name := range m.cfg.Profiles {
		score[name] = 0
	}

	for name, profile := range m.cfg.Profiles {
		conditions := profile.Conditions
		fullMatchScore := 0
		if conditions.PowerState != nil {
			fullMatchScore += *m.cfg.Scoring.PowerStateMatch
		}
		if conditions.PowerState != nil && conditions.PowerState.Value() == powerState.String() {
			score[name] += *m.cfg.Scoring.PowerStateMatch
		}
		for _, condition := range conditions.RequiredMonitors {
			if condition.Name != nil {
				fullMatchScore += *m.cfg.Scoring.NameMatch
			}
			if condition.Description != nil {
				fullMatchScore += *m.cfg.Scoring.DescriptionMatch
			}

			for _, connectedMonitor := range connectedMonitors {
				if condition.Name != nil && *condition.Name == connectedMonitor.Name {
					score[name] += *m.cfg.Scoring.NameMatch
				}
				if condition.Description != nil && *condition.Description == connectedMonitor.Description {
					score[name] += *m.cfg.Scoring.DescriptionMatch
				}
			}
		}

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
		return nil, nil
	}

	for name, profile := range m.cfg.Profiles {
		if score[name] == bestScore {
			return profile, nil
		}
	}
	return nil, nil
}
