package service

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"text/template"

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

	switch *profile.ConfigType {
	case config.Static:
		if err := s.linkConfigFile(profile); err != nil {
			return fmt.Errorf("failed to link config file: %w", err)
		}
		log.Printf("Successfully linked configuration: %s", profile.ConfigFile)
	case config.Template:
		if err := s.renderTemplateFile(profile, connectedMonitors); err != nil {
			return fmt.Errorf("failed to render template file: %w", err)
		}
		log.Printf("Successfully rendered template configuration: %s", profile.ConfigFile)
	}

	return nil
}

func (s *Service) renderTemplateFile(profile *config.Profile, connectedMonitors []*hypr.MonitorSpec) error {
	templatePath := profile.ConfigFile
	dest := *s.config.General.Destination

	templateContent, err := os.ReadFile(templatePath)
	if err != nil {
		return fmt.Errorf("failed to read template file %s: %w", templatePath, err)
	}

	tmpl, err := template.New("config").Parse(string(templateContent))
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	templateData := s.createTemplateData(profile, connectedMonitors)

	var rendered bytes.Buffer
	if err := tmpl.Execute(&rendered, templateData); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	renderedContent := rendered.Bytes()
	if existingContent, err := os.ReadFile(dest); err == nil {
		if bytes.Equal(existingContent, renderedContent) {
			if s.cfg.Verbose {
				log.Printf("Template content unchanged, skipping write: %s", dest)
			}
			return nil
		}
	}

	tempFile := dest + ".tmp"
	if err := os.WriteFile(tempFile, renderedContent, 0644); err != nil {
		return fmt.Errorf("failed to write temp config to %s: %w", tempFile, err)
	}

	if err := os.Rename(tempFile, dest); err != nil {
		return fmt.Errorf("failed to rename temp config %s to %s: %w", tempFile, dest, err)
	}

	return nil
}

func (s *Service) createTemplateData(profile *config.Profile, connectedMonitors []*hypr.MonitorSpec) map[string]any {
	data := make(map[string]any)
	data["Monitors"] = connectedMonitors

	monitorsByTag := make(map[string]*hypr.MonitorSpec)

	for _, requiredMonitor := range profile.Conditions.RequiredMonitors {
		if requiredMonitor.MonitorTag == nil {
			continue
		}

		for _, connectedMonitor := range connectedMonitors {
			if s.monitorMatches(requiredMonitor, connectedMonitor) {
				monitorsByTag[*requiredMonitor.MonitorTag] = connectedMonitor
				if s.cfg.Verbose {
					log.Printf("Mapped monitor tag '%s' to monitor: Name=%s, Description=%s",
						*requiredMonitor.MonitorTag, connectedMonitor.Name, connectedMonitor.Description)
				}
				break
			}
		}
	}

	if s.cfg.Verbose {
		log.Printf("Template data: %d monitors, %d tagged monitors", len(connectedMonitors), len(monitorsByTag))
		for tag, monitor := range monitorsByTag {
			log.Printf("  Tag '%s': %s (%s)", tag, monitor.Name, monitor.Description)
		}
	}

	data["MonitorsByTag"] = monitorsByTag
	return data
}

func (s *Service) monitorMatches(required *config.RequiredMonitor, connected *hypr.MonitorSpec) bool {
	if required.Name != nil && *required.Name != connected.Name {
		return false
	}
	if required.Description != nil && *required.Description != connected.Description {
		return false
	}
	return true
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
