// Package userconfigupdater provides the main service coordination and event handling.
package userconfigupdater

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/fiffeek/hyprdynamicmonitors/internal/config"
	"github.com/fiffeek/hyprdynamicmonitors/internal/generators"
	"github.com/fiffeek/hyprdynamicmonitors/internal/hypr"
	"github.com/fiffeek/hyprdynamicmonitors/internal/matchers"
	"github.com/fiffeek/hyprdynamicmonitors/internal/notifications"
	"github.com/fiffeek/hyprdynamicmonitors/internal/power"
	"github.com/fiffeek/hyprdynamicmonitors/internal/utils"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

type IPowerDetector interface {
	GetCurrentState(context.Context) (power.PowerState, error)
	Listen() <-chan power.PowerEvent
}

type IMonitorDetector interface {
	Listen() <-chan hypr.MonitorSpecs
	QueryConnectedMonitors(context.Context) (hypr.MonitorSpecs, error)
}

type Service struct {
	config               *config.Config
	monitorDetector      IMonitorDetector
	powerDetector        IPowerDetector
	matcher              *matchers.Matcher
	serviceConfig        *Config
	generator            *generators.ConfigGenerator
	notificationsService *notifications.Service

	stateMu          sync.RWMutex
	cachedMonitors   []*hypr.MonitorSpec
	cachedPowerState power.PowerState
	debouncer        *utils.Debouncer
}

type Config struct {
	DryRun bool
}

func NewService(cfg *config.Config, monitorDetector IMonitorDetector,
	powerDetector IPowerDetector, svcCfg *Config, matcher *matchers.Matcher, generator *generators.ConfigGenerator,
	notifications *notifications.Service,
) *Service {
	return &Service{
		config:               cfg,
		monitorDetector:      monitorDetector,
		powerDetector:        powerDetector,
		serviceConfig:        svcCfg,
		matcher:              matcher,
		generator:            generator,
		cachedPowerState:     power.BatteryPowerState,
		debouncer:            utils.NewDebouncer(),
		notificationsService: notifications,
	}
}

func (s *Service) Run(ctx context.Context) error {
	if err := s.RunOnce(ctx); err != nil {
		return fmt.Errorf("unable to update configuration on start: %w", err)
	}

	monitorEventsChannel := s.monitorDetector.Listen()
	powerEventsChannel := s.powerDetector.Listen()
	logrus.Info("Listening for monitor and power events...")

	eg, ctx := errgroup.WithContext(ctx)

	eg.Go(func() error {
		<-ctx.Done()
		s.debouncer.Cancel()
		logrus.Debug("Context cancelled for service, shutting down")
		return ctx.Err()
	})

	eg.Go(func() error {
		logrus.Debug("Running debouncer for userconfigupdater")
		if err := s.debouncer.Run(ctx); err != nil {
			return fmt.Errorf("debouncer failed: %w", err)
		}
		return nil
	})

	eg.Go(func() error {
		for {
			select {
			case monitors, ok := <-monitorEventsChannel:
				if !ok {
					return errors.New("monitor events channel closed")
				}
				logrus.WithField("monitor_count", len(monitors)).Debug("Monitor event received")
				s.stateMu.Lock()
				s.cachedMonitors = monitors
				s.stateMu.Unlock()
				s.debouncer.Do(ctx, time.Duration(*s.config.Get().General.DebounceTimeMs)*time.Millisecond, s.debounceUpdate)

			case powerEvent, ok := <-powerEventsChannel:
				if !ok {
					return errors.New("power event channel closed")
				}
				logrus.WithField("power_state", powerEvent.State.String()).Debug("Power event received")
				s.stateMu.Lock()
				s.cachedPowerState = powerEvent.State
				s.stateMu.Unlock()
				s.debouncer.Do(ctx, time.Duration(*s.config.Get().General.DebounceTimeMs)*time.Millisecond, s.debounceUpdate)

			case <-ctx.Done():
				logrus.Debug("Event processor context cancelled, shutting down")
				return ctx.Err()
			}
		}
	})

	if err := eg.Wait(); err != nil {
		return fmt.Errorf("goroutines for service failed %w", err)
	}
	return nil
}

func (s *Service) RunOnce(ctx context.Context) error {
	monitors, err := s.monitorDetector.QueryConnectedMonitors(ctx)
	if err != nil {
		return fmt.Errorf("unable to query current monitors: %w", err)
	}
	powerState, err := s.powerDetector.GetCurrentState(ctx)
	if err != nil {
		return fmt.Errorf("unable to fetch power state: %w", err)
	}

	s.stateMu.Lock()
	s.cachedMonitors = monitors
	s.cachedPowerState = powerState
	s.stateMu.Unlock()

	if err := s.UpdateOnce(); err != nil {
		return fmt.Errorf("unable to update configuration: %w", err)
	}

	return nil
}

func (s *Service) debounceUpdate(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		return s.UpdateOnce()
	}
}

func (s *Service) Handle(ctx context.Context) error {
	return s.UpdateOnce()
}

func (s *Service) UpdateOnce() error {
	s.stateMu.RLock()
	monitors := s.cachedMonitors
	powerState := s.cachedPowerState
	s.stateMu.RUnlock()

	logrus.WithFields(logrus.Fields{
		"monitor_count": len(monitors),
		"power_state":   powerState.String(),
		"dry_run":       s.serviceConfig.DryRun,
	}).Debug("Updating configuration")

	found, profile, err := s.matcher.Match(monitors, powerState)
	if err != nil {
		return fmt.Errorf("failed to match a profile %w", err)
	}

	if !found {
		logrus.Debug("No matching profile found")
		return nil
	}

	profileFields := logrus.Fields{
		"profile_name": profile.Name,
		"config_file":  profile.ConfigFile,
		"config_type":  profile.ConfigType.Value(),
	}

	if s.serviceConfig.DryRun {
		logrus.WithFields(profileFields).Info("[DRY RUN] Would use profile")
		return nil
	}

	logrus.WithFields(profileFields).Info("Using profile")

	destination := *s.config.Get().General.Destination
	changed, err := s.generator.GenerateConfig(profile, monitors, powerState, destination)
	if err != nil {
		return fmt.Errorf("failed to generate config: %w", err)
	}

	if !changed {
		logrus.Info("Not sending notifications since the config has not been changed")
		return nil
	}

	if err := s.notificationsService.NotifyProfileApplied(profile); err != nil {
		logrus.WithFields(profileFields).WithError(err).Error("swallowing notification error")
	}

	return nil
}
