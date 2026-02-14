package app

import (
	"github.com/fiffeek/hyprdynamicmonitors/internal/config"
	"github.com/sirupsen/logrus"
)

func forceFlags(disablePowerEventsChanged, disablePowerEvents bool, cfg *config.Config,
	enableLidEventsChanged, enableLidEvents bool,
) (bool, bool) {
	if !disablePowerEventsChanged && disablePowerEvents && cfg.Get().ReliesOnPowerEvents() {
		logrus.Info("config relies on power events, forcing enable. If this is not what you want pass --disable-power-events=false directly.")
		disablePowerEvents = false
	}

	if !enableLidEventsChanged && !enableLidEvents && cfg.Get().ReliesOnLidEvents() {
		logrus.Info("config relies on lid events, forcing enable. If this is not what you want pass --enable-lid-events=false directly.")
		enableLidEvents = true
	}
	return disablePowerEvents, enableLidEvents
}
