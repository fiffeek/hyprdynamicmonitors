// Package service provides the main service coordination and event handling.
package service

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/fiffeek/hyprdynamicmonitors/internal/config"
	"github.com/fiffeek/hyprdynamicmonitors/internal/detectors"
	"github.com/fiffeek/hyprdynamicmonitors/internal/generators"
	"github.com/fiffeek/hyprdynamicmonitors/internal/hypr"
	"github.com/fiffeek/hyprdynamicmonitors/internal/matchers"
	"github.com/fiffeek/hyprdynamicmonitors/internal/notifications"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

type IPowerDetector interface {
	GetCurrentState(context.Context) (detectors.PowerState, error)
	Listen() <-chan detectors.PowerEvent
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
	cachedPowerState detectors.PowerState
	debounceTimer    *time.Timer
	debounceMutex    sync.Mutex
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
		cachedPowerState:     detectors.Battery,
		debounceTimer:        time.NewTimer(0),
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
		return s.updateProcessor(ctx)
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
				s.triggerUpdate()

			case powerEvent, ok := <-powerEventsChannel:
				if !ok {
					return errors.New("power event channel closed")
				}
				logrus.WithField("power_state", powerEvent.State.String()).Debug("Power event received")
				s.stateMu.Lock()
				s.cachedPowerState = powerEvent.State
				s.stateMu.Unlock()
				s.triggerUpdate()

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

func (s *Service) triggerUpdate() {
	s.debounceMutex.Lock()
	defer s.debounceMutex.Unlock()

	s.debounceTimer.Stop()
	s.debounceTimer.Reset(time.Duration(*s.config.General.DebounceTimeMs) * time.Millisecond)

	logrus.WithField("debounce", *s.config.General.DebounceTimeMs).Debug("Update scheduled")
}

func (s *Service) updateProcessor(ctx context.Context) error {
	s.debounceTimer.Stop()

	for {
		select {
		case <-s.debounceTimer.C:
			logrus.Debug("Debounce timer expired, performing update")
			if err := s.UpdateOnce(); err != nil {
				return fmt.Errorf("configuration update failed: %w", err)
			}
		case <-ctx.Done():
			logrus.Debug("Update processor context cancelled, shutting down")
			return nil
		}
	}
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

	destination := *s.config.General.Destination
	if err := s.generator.GenerateConfig(profile, monitors, powerState, destination); err != nil {
		return fmt.Errorf("failed to generate config: %w", err)
	}

	if err := s.notificationsService.NotifyProfileApplied(profile); err != nil {
		logrus.WithFields(profileFields).WithError(err).Error("swallowing notification error")
	}

	return nil
}
