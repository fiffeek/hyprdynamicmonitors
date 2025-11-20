// Package userconfigupdater provides the main service coordination and event handling.
package userconfigupdater

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
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
	GetCurrentState() power.PowerState
	Listen() <-chan power.PowerEvent
}

type IMonitorDetector interface {
	Listen() <-chan hypr.MonitorSpecs
	GetConnectedMonitors() hypr.MonitorSpecs
}

type ILidDetector interface {
	Listen() <-chan power.LidEvent
	GetCurrentState() power.LidState
}

type Service struct {
	config               *config.Config
	monitorDetector      IMonitorDetector
	powerDetector        IPowerDetector
	lidDetector          ILidDetector
	matcher              *matchers.Matcher
	serviceConfig        *Config
	generator            *generators.ConfigGenerator
	notificationsService *notifications.Service

	stateMu          sync.RWMutex
	cachedMonitors   []*hypr.MonitorSpec
	cachedPowerState power.PowerState
	cachedLidState   power.LidState
	debouncer        *utils.Debouncer
}

type Config struct {
	DryRun bool
}

func NewService(cfg *config.Config, monitorDetector IMonitorDetector,
	powerDetector IPowerDetector, svcCfg *Config, matcher *matchers.Matcher, generator *generators.ConfigGenerator,
	notifications *notifications.Service, lidDetector ILidDetector,
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
		lidDetector:          lidDetector,
	}
}

func (s *Service) Run(ctx context.Context) error {
	if err := s.RunOnce(ctx); err != nil {
		return fmt.Errorf("unable to update configuration on start: %w", err)
	}

	monitorEventsChannel := s.monitorDetector.Listen()
	powerEventsChannel := s.powerDetector.Listen()
	lidEventsChannel := s.lidDetector.Listen()
	logrus.Info("Listening for monitor and power events...")

	eg, ctx := errgroup.WithContext(ctx)

	eg.Go(func() error {
		<-ctx.Done()
		s.debouncer.Cancel()
		logrus.Debug("Context cancelled for service, shutting down")
		return context.Cause(ctx)
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
			case lidEvent, ok := <-lidEventsChannel:
				if !ok {
					return errors.New("lid events channel closed")
				}
				s.stateMu.Lock()
				s.cachedLidState = lidEvent.State
				s.stateMu.Unlock()
				s.debouncer.Do(ctx, time.Duration(*s.config.Get().General.DebounceTimeMs)*time.Millisecond, s.debounceUpdate)
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
				return context.Cause(ctx)
			}
		}
	})

	if err := eg.Wait(); err != nil {
		return fmt.Errorf("goroutines for service failed %w", err)
	}
	return nil
}

func (s *Service) RunOnce(ctx context.Context) error {
	monitors := s.monitorDetector.GetConnectedMonitors()
	powerState := s.powerDetector.GetCurrentState()
	lidState := s.lidDetector.GetCurrentState()

	s.stateMu.Lock()
	s.cachedMonitors = monitors
	s.cachedPowerState = powerState
	s.cachedLidState = lidState
	s.stateMu.Unlock()

	if err := s.UpdateOnce(ctx); err != nil {
		return fmt.Errorf("unable to update configuration: %w", err)
	}

	return nil
}

func (s *Service) debounceUpdate(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return context.Cause(ctx)
	default:
		return s.UpdateOnce(ctx)
	}
}

func (s *Service) Handle(ctx context.Context) error {
	return s.UpdateOnce(ctx)
}

func (s *Service) UpdateOnce(ctx context.Context) error {
	s.stateMu.RLock()
	monitors := s.cachedMonitors
	powerState := s.cachedPowerState
	lidState := s.cachedLidState
	s.stateMu.RUnlock()

	// grab latest config and pass along for the same world-view
	cfg := s.config.Get()

	logrus.WithFields(logrus.Fields{
		"monitor_count": len(monitors),
		"power_state":   powerState.String(),
		"lid_state":     lidState.String(),
		"dry_run":       s.serviceConfig.DryRun,
	}).Debug("Updating configuration")

	found, matchedProfile, err := s.matcher.Match(cfg, monitors, powerState, lidState)
	if err != nil {
		return fmt.Errorf("failed to match a profile %w", err)
	}

	if !found {
		logrus.Info("No matching profile found")
		return nil
	}

	profileFields := logrus.Fields{
		"profile_name": matchedProfile.Profile.Name,
		"config_file":  matchedProfile.Profile.ConfigFile,
		"config_type":  matchedProfile.Profile.ConfigType.Value(),
	}

	if s.serviceConfig.DryRun {
		logrus.WithFields(utils.NewLogrusCustomFields(profileFields).WithLogID(utils.DryRunLogID)).
			Info("[DRY RUN] Using profile")
	}

	logrus.WithFields(profileFields).Info("Using profile")

	s.tryExec(ctx, matchedProfile.Profile.PreApplyExec, cfg.General.PreApplyExec, utils.PreExecLogID)

	destination := *cfg.General.Destination
	changed, err := s.generator.GenerateConfig(cfg, matchedProfile, monitors, powerState,
		lidState, destination, s.serviceConfig.DryRun)
	if err != nil {
		return fmt.Errorf("failed to generate config: %w", err)
	}

	// if not changed and not running in dry run then exit early
	if !changed && !s.serviceConfig.DryRun {
		logrus.Info("Not sending notifications since the config has not been changed")
		return nil
	}

	s.tryExec(ctx, matchedProfile.Profile.PostApplyExec, cfg.General.PostApplyExec, utils.PostExecLogID)

	if err := s.notificationsService.NotifyProfileApplied(matchedProfile.Profile, s.serviceConfig.DryRun); err != nil {
		logrus.WithFields(profileFields).WithError(err).Error("swallowing notification error")
	}

	return nil
}

func (s *Service) tryExec(ctx context.Context, command, fallbackCommand *string, logID utils.LogID) {
	// fallback on a default command when it's not provided for a profile
	if command == nil || *command == "" {
		command = fallbackCommand
	}
	// if it's still empty then nothing to be done
	if command == nil || *command == "" {
		return
	}
	// if running with dry run then just output the commands
	if s.serviceConfig.DryRun {
		logrus.WithFields(utils.NewLogrusCustomFields(map[string]interface{}{
			"command": *command,
			"order":   logID,
		}).WithLogID(utils.DryRunExedLogID)).
			Info("[DRY RUN] Would run command")
		return
	}

	logrus.WithFields(
		utils.NewLogrusCustomFields(logrus.Fields{"command": *command}).WithLogID(logID)).Info("Executing user callback")
	// nolint:gosec
	out, err := exec.CommandContext(ctx, "bash", "-c", *command).CombinedOutput()
	if err != nil {
		logrus.Errorf("error %v while executing %s output %s, continuing as normal", err, *command, string(out))
	}
}
