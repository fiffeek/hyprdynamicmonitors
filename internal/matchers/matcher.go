// Package matchers provides profile matching logic based on monitor conditions.
package matchers

import (
	"slices"

	"github.com/fiffeek/hyprdynamicmonitors/internal/config"
	"github.com/fiffeek/hyprdynamicmonitors/internal/hypr"
	"github.com/fiffeek/hyprdynamicmonitors/internal/power"
	"github.com/sirupsen/logrus"
)

type Matcher struct{}

func NewMatcher() *Matcher {
	return &Matcher{}
}

func (m *Matcher) Match(cfg *config.RawConfig, connectedMonitors []*hypr.MonitorSpec,
	powerState power.PowerState, lidState power.LidState,
) (bool, *MatchedProfile, error) {
	score := make(map[string]int)
	profileRules := make(map[string]map[int]*config.RequiredMonitor)
	profiles := cfg.Profiles
	ok, fallbackProfile := m.returnNoneOrFallback(cfg)
	for name := range profiles {
		score[name] = 0
	}

	for name, profile := range profiles {
		conditions := profile.Conditions
		fullMatchScore := m.calcFullProfileScore(cfg, conditions)
		profileScore, profileRule := m.scoreProfile(cfg, conditions, powerState, lidState, connectedMonitors)
		score[name] = profileScore
		profileRules[name] = profileRule
		logrus.Debugf("Profile %s score %d, full match %d", name, score[name], fullMatchScore)

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
		return ok, NewFallbackProfile(fallbackProfile), nil
	}

	// match from the last entry in the toml config
	ascProfiles := cfg.OrderedProfileKeys()
	slices.Reverse(ascProfiles)
	for _, name := range ascProfiles {
		profile := cfg.Profiles[name]
		if score[name] == bestScore {
			return true, NewMatchedProfile(profile, profileRules[name]), nil
		}
	}
	return ok, NewFallbackProfile(fallbackProfile), nil
}

func (m *Matcher) returnNoneOrFallback(cfg *config.RawConfig) (bool, *config.Profile) {
	if cfg.FallbackProfile != nil {
		return true, cfg.FallbackProfile
	}
	return false, nil
}

func (m *Matcher) scoreProfile(cfg *config.RawConfig, conditions *config.ProfileCondition,
	powerState power.PowerState, lidState power.LidState, connectedMonitors []*hypr.MonitorSpec,
) (int, map[int]*config.RequiredMonitor) {
	monitorToRule := make(map[int]*config.RequiredMonitor)
	profileScore := 0
	if conditions.PowerState != nil && conditions.PowerState.Value() == powerState.String() {
		profileScore += *cfg.Scoring.PowerStateMatch
	}

	if conditions.LidState != nil && conditions.LidState.Value() == lidState.String() {
		profileScore += *cfg.Scoring.LidStateMatch
	}

	usedMonitors := map[int]bool{}
	for _, connectedMonitor := range connectedMonitors {
		usedMonitors[*connectedMonitor.ID] = false
	}

	// one rule will match at best with one monitor
	for _, condition := range conditions.RequiredMonitors {
		// find best monitor match in terms of scoring
		bestScore, bestMonitor := 0, -1

		// iterate over all the monitors, excluding the already matched
		for _, connectedMonitor := range connectedMonitors {
			// if the monitor was matched by any other rule, skip
			if usedMonitors[*connectedMonitor.ID] {
				logrus.WithFields(logrus.Fields{"monitor_id": *connectedMonitor.ID}).Debug(
					"Monitor already matched to another rule")
				continue
			}

			// if both are defined but there is a mismatch on either then skip
			if condition.HasName() && condition.HasDescription() &&
				(!condition.MatchName(connectedMonitor.Name) ||
					!condition.MatchDescription(connectedMonitor.Description)) {
				logrus.WithFields(logrus.Fields{"monitor_id": *connectedMonitor.ID}).Debug(
					"Monitor mismatches on description or name but both are required")
				continue
			}

			logrus.WithFields(logrus.Fields{"monitor_id": *connectedMonitor.ID}).Debug("Matching monitor with rule")

			score := 0
			if condition.HasName() && condition.MatchName(connectedMonitor.Name) {
				score += *cfg.Scoring.NameMatch
				logrus.WithFields(logrus.Fields{
					"monitor_id":   *connectedMonitor.ID,
					"monitor_name": connectedMonitor.Name, "rule": *condition.Name,
				}).Debug("Name matches")
			}
			if condition.HasDescription() && condition.MatchDescription(connectedMonitor.Description) {
				score += *cfg.Scoring.DescriptionMatch
				logrus.WithFields(logrus.Fields{
					"monitor_id":   *connectedMonitor.ID,
					"monitor_name": connectedMonitor.Name, "rule": *condition.Description,
				}).Debug("Description matches")
			}

			if score > bestScore {
				bestScore = score
				bestMonitor = *connectedMonitor.ID
			}
		}

		if bestScore > 0 {
			profileScore += bestScore
			usedMonitors[bestMonitor] = true
			monitorToRule[bestMonitor] = condition
		}
	}

	return profileScore, monitorToRule
}

func (m *Matcher) calcFullProfileScore(cfg *config.RawConfig, conditions *config.ProfileCondition) int {
	fullMatchScore := 0
	if conditions.PowerState != nil {
		fullMatchScore += *cfg.Scoring.PowerStateMatch
	}

	if conditions.LidState != nil {
		fullMatchScore += *cfg.Scoring.LidStateMatch
	}

	for _, condition := range conditions.RequiredMonitors {
		if condition.HasName() {
			fullMatchScore += *cfg.Scoring.NameMatch
		}
		if condition.HasDescription() {
			fullMatchScore += *cfg.Scoring.DescriptionMatch
		}
	}

	return fullMatchScore
}
