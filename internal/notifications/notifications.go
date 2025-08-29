// Package notifications provides notifications through dbus
package notifications

import (
	"fmt"

	"github.com/TheCreeper/go-notify"
	"github.com/fiffeek/hyprdynamicmonitors/internal/config"
	"github.com/sirupsen/logrus"
)

type Service struct {
	config *config.Config
	hints  map[string]interface{}
}

func NewService(cfg *config.Config) *Service {
	return &Service{
		config: cfg,
		hints: map[string]interface{}{
			"synchronous":       "hyprdynamicmonitors",
			"x-dunst-stack-tag": "hyprdynamicmonitors",
		},
	}
}

func (s *Service) NotifyProfileApplied(profile *config.Profile) error {
	if *s.config.Get().Notifications.Disabled {
		logrus.Debug("notifications are not enabled, not sending")
		return nil
	}

	summary := "Monitor profile `" + profile.Name + "` applied"
	body := "Updated " + *s.config.Get().General.Destination
	ntf := notify.NewNotification(summary, body)
	ntf.Timeout = *s.config.Get().Notifications.TimeoutMs
	ntf.Hints = s.hints

	if _, err := ntf.Show(); err != nil {
		return fmt.Errorf("cant send notificaiton for %s: %w", profile.Name, err)
	}
	return nil
}
