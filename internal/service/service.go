package service

import (
	"fmt"
	"log"
	"sync"

	"github.com/fiffeek/hyprdynamicmonitors/internal/config"
	"github.com/fiffeek/hyprdynamicmonitors/internal/detectors"
	"github.com/fiffeek/hyprdynamicmonitors/internal/generators"
	"github.com/fiffeek/hyprdynamicmonitors/internal/hypr"
	"github.com/fiffeek/hyprdynamicmonitors/internal/matchers"
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
	updateChan       chan struct{}
}

type Config struct {
	DryRun  bool
	Verbose bool
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
		updateChan:       make(chan struct{}, 1),
	}
}

func (s *Service) Run() error {
	s.stateMu.Lock()
	s.cachedMonitors = s.monitorDetector.GetConnected()
	if powerState, err := s.powerDetector.GetCurrentState(); err == nil {
		s.cachedPowerState = powerState
	}
	s.stateMu.Unlock()

	if err := s.UpdateOnce(); err != nil {
		log.Printf("Initial configuration update failed: %v", err)
	}

	monitorEventsChannel, err := s.monitorDetector.Listen()
	if err != nil {
		return fmt.Errorf("failed to start monitor event listener: %w", err)
	}

	powerEventsChannel, err := s.powerDetector.Listen()
	if err != nil {
		return fmt.Errorf("failed to start power event listener: %w", err)
	}
	log.Printf("Listening for monitor and power events...")

	go s.updateProcessor()

	for {
		select {
		case monitors, ok := <-monitorEventsChannel:
			if !ok {
				return fmt.Errorf("monitor event channel closed unexpectedly")
			}
			s.stateMu.Lock()
			s.cachedMonitors = monitors
			s.stateMu.Unlock()
			s.triggerUpdate()

		case powerEvent, ok := <-powerEventsChannel:
			if !ok {
				return fmt.Errorf("power event channel closed unexpectedly")
			}
			s.stateMu.Lock()
			s.cachedPowerState = powerEvent.State
			s.stateMu.Unlock()
			s.triggerUpdate()
		}
	}
}

func (s *Service) triggerUpdate() {
	s.updateChan <- struct{}{}
}

func (s *Service) updateProcessor() {
	for range s.updateChan {
		if err := s.UpdateOnce(); err != nil {
			log.Printf("Configuration update failed: %v", err)
		}
	}
}

func (s *Service) UpdateOnce() error {
	s.stateMu.RLock()
	monitors := s.cachedMonitors
	powerState := s.cachedPowerState
	s.stateMu.RUnlock()

	if s.cfg.Verbose || s.cfg.DryRun {
		log.Printf("Current state: %d monitors, power: %s", len(monitors), powerState)
	}

	profile, err := s.matcher.Match(monitors, powerState)
	if err != nil {
		return fmt.Errorf("failed to match a profile %v", err)
	}

	if profile == nil {
		if s.cfg.Verbose {
			log.Printf("No matching profile found")
		}
		return nil
	}

	if s.cfg.DryRun {
		log.Printf("[DRY RUN] Would use profile: %s -> %s", profile.Name, profile.ConfigFile)
		return nil
	}

	log.Printf("Using profile: %s -> %s", profile.Name, profile.ConfigFile)

	destination := *s.config.General.Destination
	if err := s.generator.GenerateConfig(profile, monitors, powerState, destination); err != nil {
		return fmt.Errorf("failed to generate config: %w", err)
	}

	switch *profile.ConfigType {
	case config.Static:
		log.Printf("Successfully linked configuration: %s", profile.ConfigFile)
	case config.Template:
		log.Printf("Successfully rendered template configuration: %s", profile.ConfigFile)
	}

	return nil
}
