// Package generators provides configuration file generation functionality.
package generators

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"sync"
	"text/template"
	"time"

	"github.com/fiffeek/hyprdynamicmonitors/internal/config"
	"github.com/fiffeek/hyprdynamicmonitors/internal/hypr"
	"github.com/fiffeek/hyprdynamicmonitors/internal/matchers"
	"github.com/fiffeek/hyprdynamicmonitors/internal/power"
	"github.com/fiffeek/hyprdynamicmonitors/internal/utils"
	"github.com/sirupsen/logrus"
)

type ConfigGenerator struct {
	mtime   map[string]time.Time
	mtimeMu sync.RWMutex
}

func NewConfigGenerator(cfg *config.Config) *ConfigGenerator {
	mtime := make(map[string]time.Time)
	for _, profile := range cfg.Get().Profiles {
		mtime[profile.ConfigFile] = profile.ConfigFileModTime
	}

	return &ConfigGenerator{mtime: mtime, mtimeMu: sync.RWMutex{}}
}

// GenerateConfig either renders a template or links a file, and returns if any changed were done
// this includes stating the config files to catch if the user modified them by hand (in linking scenario)
func (g *ConfigGenerator) GenerateConfig(cfg *config.RawConfig, profile *matchers.MatchedProfile,
	connectedMonitors []*hypr.MonitorSpec, powerState power.PowerState, lidState power.LidState, destination string,
) (bool, error) {
	switch *profile.Profile.ConfigType {
	case config.Static:
		return g.linkConfigFile(profile.Profile, destination)
	case config.Template:
		return g.renderTemplateFile(cfg, profile, connectedMonitors, powerState, lidState, destination)
	default:
		return false, fmt.Errorf("unsupported config type: %v", *profile.Profile.ConfigType)
	}
}

func (g *ConfigGenerator) renderTemplateFile(cfg *config.RawConfig, profile *matchers.MatchedProfile,
	connectedMonitors []*hypr.MonitorSpec, powerState power.PowerState, lidState power.LidState, destination string,
) (bool, error) {
	templatePath := profile.Profile.ConfigFile

	//nolint:gosec
	templateContent, err := os.ReadFile(templatePath)
	if err != nil {
		return false, fmt.Errorf("failed to read template file %s: %w", templatePath, err)
	}

	tmpl, err := template.New("config").Funcs(getFuncMap(powerState, lidState)).Parse(string(templateContent))
	if err != nil {
		return false, fmt.Errorf("failed to parse template: %w", err)
	}

	templateData := g.createTemplateData(cfg, profile, connectedMonitors, powerState, lidState)

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

func getFuncMap(powerState power.PowerState, lidState power.LidState) template.FuncMap {
	funcMap := template.FuncMap{
		"isOnBattery": func() bool {
			return powerState == power.BatteryPowerState
		},
		"isOnAC": func() bool {
			return powerState == power.ACPowerState
		},
		"powerState": func() string {
			return powerState.String()
		},
		"lidState": func() string {
			return lidState.String()
		},
		"isLidClosed": func() bool {
			return lidState == power.ClosedLidState
		},
		"isLidOpened": func() bool {
			return lidState == power.OpenedLidState
		},
	}
	return funcMap
}

func (g *ConfigGenerator) createTemplateData(cfg *config.RawConfig, profile *matchers.MatchedProfile,
	connectedMonitors []*hypr.MonitorSpec, powerState power.PowerState, lidState power.LidState,
) map[string]any {
	data := make(map[string]any)

	// strip the data for templating, no need for extra fields such as the current resolution as that might confuse users too much
	// and lead to e.g. infinite config applications
	monitorsStripped := []*MonitorSpec{}
	for _, monitor := range connectedMonitors {
		monitorsStripped = append(monitorsStripped, NewMonitorSpec(monitor))
	}

	// Monitors represents all connected monitors, might not be defined in the profile, might be disabled etc.
	data["Monitors"] = monitorsStripped
	data["PowerState"] = powerState.String()
	data["LidState"] = lidState.String()

	requiredMonitors := []*MonitorSpec{}
	extraMonitors := []*MonitorSpec{}

	monitorsByTag := make(map[string]*MonitorSpec)

	for _, connectedMonitor := range monitorsStripped {
		requiredMonitor, ok := profile.MonitorToRule[*connectedMonitor.ID]
		if !ok {
			extraMonitors = append(extraMonitors, connectedMonitor)
			continue
		}

		if requiredMonitor.MonitorTag != nil {
			monitorsByTag[*requiredMonitor.MonitorTag] = connectedMonitor
			logrus.WithFields(logrus.Fields{
				"tag":         *requiredMonitor.MonitorTag,
				"name":        connectedMonitor.Name,
				"description": connectedMonitor.Description,
			}).Debug("Mapped monitor tag")
		}

		requiredMonitors = append(requiredMonitors, connectedMonitor)
	}

	data["ExtraMonitors"] = extraMonitors

	for _, monitor := range requiredMonitors {
		logrus.WithFields(logrus.Fields{
			"name":        monitor.Name,
			"description": monitor.Description,
			"enabled":     !monitor.Disabled,
		}).Debug("Required monitor")
	}
	data["RequiredMonitors"] = requiredMonitors

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

	logrus.Debug("Adding user defined values")
	for key, value := range cfg.StaticTemplateValues {
		data[key] = value
		logrus.WithFields(logrus.Fields{
			"key":   key,
			"value": value,
		}).Debug("Added user kv pair")
	}

	logrus.Debug("Adding profile defined values")
	for key, value := range profile.Profile.StaticTemplateValues {
		_, ok := data[key]
		data[key] = value

		action := "Added"
		if ok {
			action = "Overwritten"
		}
		logrus.WithFields(logrus.Fields{
			"key":   key,
			"value": value,
		}).Debug(action + " user kv pair")
	}

	return data
}

func (g *ConfigGenerator) linkConfigFile(profile *config.Profile, destination string) (bool, error) {
	source := profile.ConfigFile
	differentContents, err := g.compareSymlinks(destination, source, profile)
	if err == nil {
		return differentContents, nil
	}

	if _, err := os.Stat(destination); err == nil || !os.IsNotExist(err) {
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

// checkCurrentLink returns if both files were symlinks and whether the underlying contents are different
func (g *ConfigGenerator) compareSymlinks(destination, source string, profile *config.Profile) (bool, error) {
	// if there is a symlink already see where it points to,
	// then compare the locations and mtime
	//
	fileInfo, err := os.Lstat(destination)
	if err != nil {
		return false, fmt.Errorf("not a symlink %s: %w", destination, err)
	}
	if fileInfo.Mode()&os.ModeSymlink == 0 {
		return false, errors.New("not a symlink")
	}
	target, err := os.Readlink(destination)
	if err != nil {
		return false, fmt.Errorf("cant readlink %s: %w", destination, err)
	}
	if target == source {
		sourceFileInfo, err := os.Lstat(source)
		if err != nil {
			return false, fmt.Errorf("cant stat %s: %w", source, err)
		}

		g.mtimeMu.Lock()
		defer g.mtimeMu.Unlock()

		prevMtime, ok := g.mtime[source]
		changed := false
		if ok {
			changed = prevMtime.Compare(sourceFileInfo.ModTime()) != 0
		}
		g.mtime[source] = sourceFileInfo.ModTime()

		logrus.WithFields(logrus.Fields{
			"config_file": profile.ConfigFile,
			"destination": destination,
		}).Info("Configuration already correctly linked")
		return changed, nil
	}

	return false, errors.New("not a symlink")
}
