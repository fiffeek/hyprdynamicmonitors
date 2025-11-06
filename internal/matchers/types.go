package matchers

import "github.com/fiffeek/hyprdynamicmonitors/internal/config"

type MatchedProfile struct {
	Profile       *config.Profile
	MonitorToRule map[int]*config.RequiredMonitor
}

func NewMatchedProfile(profile *config.Profile, monitorToRule map[int]*config.RequiredMonitor) *MatchedProfile {
	if profile == nil {
		return nil
	}
	return &MatchedProfile{
		profile,
		monitorToRule,
	}
}

func NewFallbackProfile(profile *config.Profile) *MatchedProfile {
	if profile == nil {
		return nil
	}
	return &MatchedProfile{
		profile,
		make(map[int]*config.RequiredMonitor),
	}
}
