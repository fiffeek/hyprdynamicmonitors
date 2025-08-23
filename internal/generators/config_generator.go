package generators

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"text/template"

	"github.com/fiffeek/hyprdynamicmonitors/internal/config"
	"github.com/fiffeek/hyprdynamicmonitors/internal/hypr"
)

type ConfigGenerator struct {
	verbose bool
}

type GeneratorConfig struct {
	Verbose bool
}

func NewConfigGenerator(cfg *GeneratorConfig) *ConfigGenerator {
	return &ConfigGenerator{
		verbose: cfg.Verbose,
	}
}

func (g *ConfigGenerator) GenerateConfig(profile *config.Profile, connectedMonitors []*hypr.MonitorSpec, destination string) error {
	switch *profile.ConfigType {
	case config.Static:
		return g.linkConfigFile(profile, destination)
	case config.Template:
		return g.renderTemplateFile(profile, connectedMonitors, destination)
	default:
		return fmt.Errorf("unsupported config type: %v", *profile.ConfigType)
	}
}

func (g *ConfigGenerator) renderTemplateFile(profile *config.Profile, connectedMonitors []*hypr.MonitorSpec, destination string) error {
	templatePath := profile.ConfigFile

	templateContent, err := os.ReadFile(templatePath)
	if err != nil {
		return fmt.Errorf("failed to read template file %s: %w", templatePath, err)
	}

	tmpl, err := template.New("config").Parse(string(templateContent))
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	templateData := g.createTemplateData(profile, connectedMonitors)

	var rendered bytes.Buffer
	if err := tmpl.Execute(&rendered, templateData); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	renderedContent := rendered.Bytes()
	if existingContent, err := os.ReadFile(destination); err == nil {
		if bytes.Equal(existingContent, renderedContent) {
			if g.verbose {
				log.Printf("Template content unchanged, skipping write: %s", destination)
			}
			return nil
		}
	}

	tempFile := destination + ".tmp"
	if err := os.WriteFile(tempFile, renderedContent, 0644); err != nil {
		return fmt.Errorf("failed to write temp config to %s: %w", tempFile, err)
	}

	if err := os.Rename(tempFile, destination); err != nil {
		return fmt.Errorf("failed to rename temp config %s to %s: %w", tempFile, destination, err)
	}

	return nil
}

func (g *ConfigGenerator) createTemplateData(profile *config.Profile, connectedMonitors []*hypr.MonitorSpec) map[string]any {
	data := make(map[string]any)
	data["Monitors"] = connectedMonitors

	monitorsByTag := make(map[string]*hypr.MonitorSpec)

	for _, requiredMonitor := range profile.Conditions.RequiredMonitors {
		if requiredMonitor.MonitorTag == nil {
			continue
		}

		for _, connectedMonitor := range connectedMonitors {
			if g.monitorMatches(requiredMonitor, connectedMonitor) {
				monitorsByTag[*requiredMonitor.MonitorTag] = connectedMonitor
				if g.verbose {
					log.Printf("Mapped monitor tag '%s' to monitor: Name=%s, Description=%s",
						*requiredMonitor.MonitorTag, connectedMonitor.Name, connectedMonitor.Description)
				}
				break
			}
		}
	}

	if g.verbose {
		log.Printf("Template data: %d monitors, %d tagged monitors", len(connectedMonitors), len(monitorsByTag))
		for tag, monitor := range monitorsByTag {
			log.Printf("  Tag '%s': %s (%s)", tag, monitor.Name, monitor.Description)
		}
	}

	data["MonitorsByTag"] = monitorsByTag
	return data
}

func (g *ConfigGenerator) monitorMatches(required *config.RequiredMonitor, connected *hypr.MonitorSpec) bool {
	if required.Name != nil && *required.Name != connected.Name {
		return false
	}
	if required.Description != nil && *required.Description != connected.Description {
		return false
	}
	return true
}

func (g *ConfigGenerator) linkConfigFile(profile *config.Profile, destination string) error {
	source := profile.ConfigFile
	if _, err := os.Lstat(destination); err == nil {
		if err := os.Remove(destination); err != nil {
			return fmt.Errorf("failed to remove existing config: %w", err)
		}
	}
	if err := os.Symlink(source, destination); err != nil {
		return fmt.Errorf("failed to create symlink from %s to %s: %w", source, destination, err)
	}

	return nil
}
