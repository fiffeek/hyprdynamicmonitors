package service

import (
	"fmt"
	"log"
	"os"

	"github.com/fiffeek/hyprdynamicmonitors/internal/config"
	"github.com/fiffeek/hyprdynamicmonitors/internal/detectors"
	"github.com/fiffeek/hyprdynamicmonitors/internal/hypr"
	"github.com/fiffeek/hyprdynamicmonitors/internal/matchers"
)

type Service struct {
	config          *config.Config
	monitorDetector *detectors.MonitorDetector
	matcher         *matchers.Matcher
	cfg             *Config
}

type Config struct {
	DryRun  bool
	Verbose bool
}

func NewService(cfg *config.Config, monitorDetector *detectors.MonitorDetector, svcCfg *Config, matcher *matchers.Matcher) *Service {
	return &Service{
		config:          cfg,
		monitorDetector: monitorDetector,
		cfg:             svcCfg,
		matcher:         matcher,
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

	if err := s.linkConfigFile(profile); err != nil {
		return fmt.Errorf("failed to link config file: %w", err)
	}

	log.Printf("Successfully linked configuration: %s", profile.ConfigFile)
	return nil
}

func (s *Service) linkConfigFile(profile *config.Profile) error {
	source := profile.ConfigFile
	dest := *s.config.General.Destination
	if _, err := os.Lstat(dest); err == nil {
		if err := os.Remove(dest); err != nil {
			return fmt.Errorf("failed to remove existing config: %w", err)
		}
	}
	if err := os.Symlink(source, dest); err != nil {
		return fmt.Errorf("failed to create symlink from %s to %s: %w", source, dest, err)
	}

	return nil
}
