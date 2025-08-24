package service

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/fiffeek/hyprdynamicmonitors/internal/config"
	"github.com/fiffeek/hyprdynamicmonitors/internal/detectors"
	"github.com/fiffeek/hyprdynamicmonitors/internal/generators"
	"github.com/fiffeek/hyprdynamicmonitors/internal/hypr"
	"github.com/fiffeek/hyprdynamicmonitors/internal/matchers"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

type Service struct {
	config          *config.Config
	monitorDetector *detectors.MonitorDetector
	powerDetector   *detectors.PowerDetector
	matcher         *matchers.Matcher
	cfg             *Config
	generator       *generators.ConfigGenerator

	stateMu          sync.RWMutex
	cachedMonitors   []*hypr.MonitorSpec
	cachedPowerState detectors.PowerState
	debounceTimer    *time.Timer
	debounceMutex    sync.Mutex
}

type Config struct {
	DryRun bool
}

func NewService(cfg *config.Config, monitorDetector *detectors.MonitorDetector, powerDetector *detectors.PowerDetector, svcCfg *Config, matcher *matchers.Matcher, generator *generators.ConfigGenerator) *Service {
	return &Service{
		config:           cfg,
		monitorDetector:  monitorDetector,
		powerDetector:    powerDetector,
		cfg:              svcCfg,
		matcher:          matcher,
		generator:        generator,
		cachedPowerState: detectors.Battery,
		debounceTimer:    time.NewTimer(0),
	}
}

func (s *Service) Run(ctx context.Context) error {
	if err := s.RunOnce(); err != nil {
		return fmt.Errorf("unable to update configuration on start: %v", err)
	}

	monitorEventsChannel := s.monitorDetector.Listen()

	powerEventsChannel := s.powerDetector.Listen()
	logrus.Info("Listening for monitor and power events...")

	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		return s.updateProcessor(ctx)
	})

	g.Go(func() error {
		for {
			select {
			case monitors, ok := <-monitorEventsChannel:
				if !ok {
					return fmt.Errorf("monitor events channel closed")
				}
				logrus.WithField("monitor_count", len(monitors)).Debug("Monitor event received")
				s.stateMu.Lock()
				s.cachedMonitors = monitors
				s.stateMu.Unlock()
				s.triggerUpdate()

			case powerEvent, ok := <-powerEventsChannel:
				if !ok {
					return fmt.Errorf("power event channel closed")
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

	return g.Wait()
}

func (s *Service) RunOnce() error {
	monitors := s.monitorDetector.GetConnected()
	powerState, err := s.powerDetector.GetCurrentState()
	if err != nil {
		return fmt.Errorf("unable to fetch power state: %v", err)
	}

	s.stateMu.Lock()
	s.cachedMonitors = monitors
	s.cachedPowerState = powerState
	s.stateMu.Unlock()

	if err := s.UpdateOnce(); err != nil {
		return fmt.Errorf("unable to update configuration: %v", err)
	}

	return nil
}

func (s *Service) triggerUpdate() {
	s.debounceMutex.Lock()
	defer s.debounceMutex.Unlock()

	s.debounceTimer.Stop()
	s.debounceTimer.Reset(1500 * time.Millisecond)

	logrus.Debug("Update scheduled (debounced 1500ms)")
}

func (s *Service) updateProcessor(ctx context.Context) error {
	s.debounceTimer.Stop()

	for {
		select {
		case <-s.debounceTimer.C:
			logrus.Debug("Debounce timer expired, performing update")
			if err := s.UpdateOnce(); err != nil {
				return fmt.Errorf("configuration update failed: %v", err)
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
		"dry_run":       s.cfg.DryRun,
	}).Debug("Updating configuration")

	profile, err := s.matcher.Match(monitors, powerState)
	if err != nil {
		return fmt.Errorf("failed to match a profile %v", err)
	}

	if profile == nil {
		logrus.Debug("No matching profile found")
		return nil
	}

	profileFields := logrus.Fields{
		"profile_name": profile.Name,
		"config_file":  profile.ConfigFile,
		"config_type":  profile.ConfigType.Value(),
	}

	if s.cfg.DryRun {
		logrus.WithFields(profileFields).Info("[DRY RUN] Would use profile")
		return nil
	}

	logrus.WithFields(profileFields).Info("Using profile")

	destination := *s.config.General.Destination
	if err := s.generator.GenerateConfig(profile, monitors, powerState, destination); err != nil {
		return fmt.Errorf("failed to generate config: %w", err)
	}

	return nil
}
