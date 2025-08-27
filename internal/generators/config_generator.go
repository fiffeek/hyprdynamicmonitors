// Package generators provides configuration file generation functionality.
package generators

import (
	"bytes"
	"fmt"
	"os"
	"text/template"

	"github.com/fiffeek/hyprdynamicmonitors/internal/config"
	"github.com/fiffeek/hyprdynamicmonitors/internal/detectors"
	"github.com/fiffeek/hyprdynamicmonitors/internal/hypr"
	"github.com/fiffeek/hyprdynamicmonitors/internal/utils"
	"github.com/sirupsen/logrus"
)

type ConfigGenerator struct{}

func NewConfigGenerator() *ConfigGenerator {
	return &ConfigGenerator{}
}

func (g *ConfigGenerator) GenerateConfig(profile *config.Profile,
	connectedMonitors []*hypr.MonitorSpec, powerState detectors.PowerState, destination string,
) (bool, error) {
	switch *profile.ConfigType {
	case config.Static:
		return g.linkConfigFile(profile, destination)
	case config.Template:
		return g.renderTemplateFile(profile, connectedMonitors, powerState, destination)
	default:
		return false, fmt.Errorf("unsupported config type: %v", *profile.ConfigType)
	}
}

func (g *ConfigGenerator) renderTemplateFile(profile *config.Profile,
	connectedMonitors []*hypr.MonitorSpec, powerState detectors.PowerState, destination string,
) (bool, error) {
	templatePath := profile.ConfigFile

	//nolint:gosec
	templateContent, err := os.ReadFile(templatePath)
	if err != nil {
		return false, fmt.Errorf("failed to read template file %s: %w", templatePath, err)
	}

	tmpl, err := template.New("config").Funcs(getFuncMap(powerState)).Parse(string(templateContent))
	if err != nil {
		return false, fmt.Errorf("failed to parse template: %w", err)
	}

	templateData := g.createTemplateData(profile, connectedMonitors, powerState)

	var rendered bytes.Buffer
	if err := tmpl.Execute(&rendered, templateData); err != nil {
		return false, fmt.Errorf("failed to execute template: %w", err)
	}

	renderedContent := rendered.Bytes()
	//nolint:gosec
	if existingContent, err := os.ReadFile(destination); err == nil {
		if bytes.Equal(existingContent, renderedContent) {
			logrus.WithField("destination", destination).Info(
				"Template content unchanged, skipping write")
			return false, nil
		}
	}

	if err := utils.WriteAtomic(destination, renderedContent); err != nil {
		return false, fmt.Errorf("cant write to file %s, contents %s: %w", destination, renderedContent, err)
	}

	logrus.WithFields(logrus.Fields{
		"config_file": templatePath,
		"destination": destination,
	}).Info("Successfully rendered template configuration")

	return true, nil
}

func getFuncMap(powerState detectors.PowerState) template.FuncMap {
	funcMap := template.FuncMap{
		"isOnBattery": func() bool {
			return powerState == detectors.Battery
		},
		"isOnAC": func() bool {
			return powerState == detectors.ACPower
		},
		"powerState": func() string {
			return powerState.String()
		},
	}
	return funcMap
}

func (g *ConfigGenerator) createTemplateData(profile *config.Profile,
	connectedMonitors []*hypr.MonitorSpec, powerState detectors.PowerState,
) map[string]any {
	data := make(map[string]any)
	data["Monitors"] = connectedMonitors
	data["PowerState"] = powerState.String()

	monitorsByTag := make(map[string]*hypr.MonitorSpec)

	for _, requiredMonitor := range profile.Conditions.RequiredMonitors {
		if requiredMonitor.MonitorTag == nil {
			continue
		}

		for _, connectedMonitor := range connectedMonitors {
			if g.monitorMatches(requiredMonitor, connectedMonitor) {
				monitorsByTag[*requiredMonitor.MonitorTag] = connectedMonitor
				logrus.WithFields(logrus.Fields{
					"tag":         *requiredMonitor.MonitorTag,
					"name":        connectedMonitor.Name,
					"description": connectedMonitor.Description,
				}).Debug("Mapped monitor tag")
				break
			}
		}
	}

	logrus.WithFields(logrus.Fields{
		"monitor_count":        len(connectedMonitors),
		"tagged_monitor_count": len(monitorsByTag),
		"power_state":          powerState.String(),
	}).Debug("Template data prepared")

	for tag, monitor := range monitorsByTag {
		logrus.WithFields(logrus.Fields{
			"tag":         tag,
			"name":        monitor.Name,
			"description": monitor.Description,
		}).Debug("Tagged monitor")
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

func (g *ConfigGenerator) linkConfigFile(profile *config.Profile, destination string) (bool, error) {
	source := profile.ConfigFile
	if fileInfo, err := os.Lstat(destination); err == nil {
		if fileInfo.Mode()&os.ModeSymlink != 0 {
			target, err := os.Readlink(destination)
			if err != nil {
				return false, fmt.Errorf("cant readlink %s: %w", destination, err)
			}
			if target == source {
				logrus.WithFields(logrus.Fields{
					"config_file": profile.ConfigFile,
					"destination": destination,
				}).Info("Configuration already correctly linked")
				return false, nil
			}
		}

		if err := os.Remove(destination); err != nil {
			return false, fmt.Errorf("failed to remove existing config: %w", err)
		}
	}
	if err := os.Symlink(source, destination); err != nil {
		return false, fmt.Errorf("failed to create symlink from %s to %s: %w", source, destination, err)
	}

	logrus.WithFields(logrus.Fields{
		"config_file": profile.ConfigFile,
		"destination": destination,
	}).Info("Successfully linked configuration")

	return true, nil
}
