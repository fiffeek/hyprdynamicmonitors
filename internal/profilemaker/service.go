// Package profilemaker makes it easier to create profiles from the cli (gives a reasonable starter profile in a new monitor environment)
package profilemaker

import (
	"bytes"
	_ "embed"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/BurntSushi/toml"
	"github.com/fiffeek/hyprdynamicmonitors/internal/config"
	"github.com/fiffeek/hyprdynamicmonitors/internal/hypr"
	"github.com/fiffeek/hyprdynamicmonitors/internal/utils"
	"github.com/sirupsen/logrus"
)

//go:embed templates/monitors.go.tmpl
var monitorsTemplate string

//go:embed templates/tui.go.tmpl
var tuiTemplate string

type Service struct {
	cfg *config.Config
	ipc *hypr.IPC
}

func NewService(cfg *config.Config, ipc *hypr.IPC) *Service {
	return &Service{
		cfg: cfg,
		ipc: ipc,
	}
}

func (s *Service) FreezeCurrentAs(profileName, profileFileLocation string) error {
	currentMonitors := s.ipc.GetConnectedMonitors()
	return s.FreezeGivenAs(profileName, profileFileLocation, currentMonitors)
}

func (s *Service) FreezeGivenAs(profileName, profileFileLocation string, currentMonitors []*hypr.MonitorSpec) error {
	cfg := s.cfg.Get()
	profile, err := s.prepare(profileName, profileFileLocation, currentMonitors)
	if err != nil {
		return fmt.Errorf("cant create a new profile: %w", err)
	}

	if err := s.validate(cfg, profileName, profile); err != nil {
		return fmt.Errorf("cant validate basic new profile properties: %w", err)
	}

	profileSpec, err := s.encode(profile)
	if err != nil {
		return fmt.Errorf("cant encode new profile: %w", err)
	}

	cleanUp, err := s.render(currentMonitors, profile)
	if err != nil {
		return fmt.Errorf("cant render the new profile config: %w", err)
	}

	if err := profile.Validate(cfg.ConfigPath); err != nil {
		_ = cleanUp()
		return fmt.Errorf("cant validate a new profile: %w", err)
	}

	err = s.append(profileSpec, cfg)
	if err != nil {
		_ = cleanUp()
		return fmt.Errorf("cant replace the config file: %w", err)
	}

	return nil
}

func (s *Service) EditExisting(profileName string, currentMonitors []*hypr.MonitorSpec) error {
	cfg := s.cfg.Get()
	profile, ok := cfg.Profiles[profileName]
	if !ok {
		return errors.New("profile not found")
	}

	tmpl, err := template.New("part").Parse(tuiTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	templateData := map[string]any{
		"Monitors": currentMonitors,
	}

	var rendered bytes.Buffer
	if err := tmpl.Execute(&rendered, templateData); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	err = s.updateConfigFileWithContent(profile.ConfigFile, rendered.String())
	if err != nil {
		return fmt.Errorf("failed to update config file: %w", err)
	}

	return nil
}

// updateConfigFileWithContent reads the config file, finds TUI AUTO markers and replaces content between them,
// or appends the content to the end if markers are not found
func (s *Service) updateConfigFileWithContent(configFile, newContent string) error {
	const (
		startMarker = "# <<<<< TUI AUTO START"
		endMarker   = "# <<<<< TUI AUTO END"
	)

	existingContent, err := os.ReadFile(configFile)
	if err != nil {
		if os.IsNotExist(err) {
			return os.WriteFile(configFile, []byte(newContent), 0o644)
		}
		return fmt.Errorf("failed to read config file: %w", err)
	}

	content := string(existingContent)
	finalContent := s.newMethod(content, startMarker, endMarker, newContent)
	return utils.WriteAtomic(configFile, []byte(finalContent))
}

func (*Service) newMethod(content, startMarker, endMarker, newContent string) string {
	startIdx := strings.Index(content, startMarker)
	endIdx := strings.Index(content, endMarker)
	if startIdx != -1 && endIdx != -1 && endIdx > startIdx {
		beforeMarker := content[:startIdx]
		afterMarker := content[endIdx+len(endMarker):]

		if !strings.HasSuffix(beforeMarker, "\n") && beforeMarker != "" {
			beforeMarker += "\n"
		}
		if !strings.HasPrefix(afterMarker, "\n") && afterMarker != "" {
			afterMarker = "\n" + afterMarker
		}

		return beforeMarker + newContent + afterMarker
	}

	finalContent := content
	if !strings.HasSuffix(finalContent, "\n") && finalContent != "" {
		finalContent += "\n"
	}
	finalContent += newContent

	return finalContent
}

func (s *Service) append(profileSpec *bytes.Buffer, cfg *config.RawConfig) error {
	content, err := os.ReadFile(cfg.ConfigPath)
	if err != nil {
		return fmt.Errorf("cant read the current config file: %w", err)
	}
	logrus.Debugf("Current config content %s", string(content))

	appendContent := strings.Replace(profileSpec.String(), "[profiles]", "", 1)
	newContent := fmt.Sprintf("%s\n%s", string(content), appendContent)
	if err := utils.WriteAtomic(cfg.ConfigPath, []byte(newContent)); err != nil {
		return fmt.Errorf("cant write the final config file: %w", err)
	}
	logrus.Debugf("Wrote: %s to the configuration file %s", newContent, cfg.ConfigPath)

	return nil
}

func (*Service) render(currentMonitors hypr.MonitorSpecs, profile *config.Profile) (func() error, error) {
	tmpl, err := template.New("config").Parse(monitorsTemplate)
	if err != nil {
		return nil, fmt.Errorf("failed to parse template: %w", err)
	}

	templateData := map[string]any{
		"Monitors": currentMonitors,
	}

	var rendered bytes.Buffer
	if err := tmpl.Execute(&rendered, templateData); err != nil {
		return nil, fmt.Errorf("failed to execute template: %w", err)
	}

	renderedContent := rendered.Bytes()
	logrus.Debugf("Rendered data: %s", string(renderedContent))

	if err := utils.WriteAtomic(profile.ConfigFile, renderedContent); err != nil {
		return nil, fmt.Errorf("cant write to file: %w", err)
	}
	return func() error {
		return os.Remove(profile.ConfigFile)
	}, nil
}

func (*Service) encode(profile *config.Profile) (*bytes.Buffer, error) {
	config := config.RawConfig{
		Profiles: map[string]*config.Profile{
			profile.Name: profile,
		},
	}
	buf := new(bytes.Buffer)
	encoder := toml.NewEncoder(buf)
	encoder.Indent = ""
	if err := encoder.Encode(config); err != nil {
		return nil, fmt.Errorf("cant encode config: %w", err)
	}
	logrus.Debugf("Encoded data: %s", buf.String())
	return buf, nil
}

func (*Service) validate(cfg *config.RawConfig, profileName string, profile *config.Profile) error {
	for _, existingProfile := range cfg.Profiles {
		if existingProfile.Name == profileName {
			return errors.New("a profile with this name already exists")
		}
	}
	if fi, _ := os.Stat(profile.ConfigFile); fi != nil {
		return errors.New("template profile file already exists, pass another in --config-file-location")
	}
	if err := os.MkdirAll(filepath.Dir(profile.ConfigFile), 0o750); err != nil {
		return fmt.Errorf("cant create directory: %w", err)
	}
	return nil
}

func (s *Service) prepare(profileName, profileFileLocation string,
	currentMonitors hypr.MonitorSpecs,
) (*config.Profile, error) {
	requiredMonitors := make([]*config.RequiredMonitor, len(currentMonitors))
	for i, monitor := range currentMonitors {
		requiredMonitors[i] = &config.RequiredMonitor{
			Description: &monitor.Description,
			MonitorTag:  utils.StringPtr(fmt.Sprintf("monitor%d", *monitor.ID)),
		}
	}
	profile := config.Profile{
		Name: profileName,
		Conditions: &config.ProfileCondition{
			RequiredMonitors: requiredMonitors,
		},
		ConfigType: utils.JustPtr(config.Template),
		ConfigFile: profileFileLocation,
	}
	if err := profile.SetPath(s.cfg.Get().ConfigDirPath); err != nil {
		return nil, fmt.Errorf("cant set the profile path: %w", err)
	}
	return &profile, nil
}
