package service

import (
	"fmt"
	"log"

	"github.com/fiffeek/hyprdynamicmonitors/internal/config"
	"github.com/fiffeek/hyprdynamicmonitors/internal/detectors"
	"github.com/fiffeek/hyprdynamicmonitors/internal/generators"
	"github.com/fiffeek/hyprdynamicmonitors/internal/hypr"
	"github.com/fiffeek/hyprdynamicmonitors/internal/matchers"
)

type Service struct {
	config          *config.Config
	monitorDetector *detectors.MonitorDetector
	matcher         *matchers.Matcher
	cfg             *Config
	generator       *generators.ConfigGenerator
}

type Config struct {
	DryRun  bool
	Verbose bool
}

func NewService(cfg *config.Config, monitorDetector *detectors.MonitorDetector, svcCfg *Config, matcher *matchers.Matcher, generator *generators.ConfigGenerator) *Service {
	return &Service{
		config:          cfg,
		monitorDetector: monitorDetector,
		cfg:             svcCfg,
		matcher:         matcher,
		generator:       generator,
	}
}

func (s *Service) Run() error {
	connectedMonitors := s.monitorDetector.GetConnected()
	if err := s.UpdateOnce(connectedMonitors); err != nil {
		log.Printf("Initial configuration update failed: %v", err)
	}

	connectedMonitorsChannel, err := s.monitorDetector.Listen()
	if err != nil {
		return fmt.Errorf("failed to start event listener: %w", err)
	}

	log.Printf("Listening for monitor events...")

	for monitors := range connectedMonitorsChannel {
		if err := s.UpdateOnce(monitors); err != nil {
			log.Printf("Configuration update failed: %v", err)
		}
	}

	return fmt.Errorf("event listener stopped unexpectedly")
}

func (s *Service) UpdateOnce(connectedMonitors []*hypr.MonitorSpec) error {
	if s.cfg.Verbose || s.cfg.DryRun {
		log.Printf("Connected monitors: %v,", connectedMonitors)
	}

	profile, err := s.matcher.Match(connectedMonitors)
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
	if err := s.generator.GenerateConfig(profile, connectedMonitors, destination); err != nil {
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
