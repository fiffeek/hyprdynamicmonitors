// Package config handles loading and validation of TOML configuration files.
package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/fiffeek/hyprdynamicmonitors/internal/utils"
	"github.com/sirupsen/logrus"
)

type Config struct {
	configPath  string
	Profiles    map[string]*Profile `toml:"profiles"`
	General     *GeneralSection     `toml:"general"`
	Scoring     *ScoringSection     `toml:"scoring"`
	PowerEvents *PowerSection       `toml:"power_events"`
}

type PowerSection struct {
	Disabled                 *bool                      `toml:"disabled"`
	DbusSignalMatchRules     []*DbusSignalMatchRule     `toml:"dbus_signal_match_rules"`
	DbusSignalReceiveFilters []*DbusSignalReceiveFilter `toml:"dbus_signal_receive_filters"`
}

type DbusSignalReceiveFilter struct {
	Name *string `toml:"name"`
}

type DbusSignalMatchRule struct {
	Sender     *string `toml:"sender"`
	Interface  *string `toml:"interface"`
	Member     *string `toml:"member"`
	ObjectPath *string `toml:"object_path"`
}

type GeneralSection struct {
	Destination *string `toml:"destination"`
}

type ScoringSection struct {
	NameMatch        *int `toml:"name_match"`
	DescriptionMatch *int `toml:"description_match"`
	PowerStateMatch  *int `toml:"power_state_match"`
}

type ConfigFileType int

const (
	Static ConfigFileType = iota
	Template
)

func (e *ConfigFileType) Value() string {
	switch *e {
	case Static:
		return "static"
	case Template:
		return "template"
	}
	return ""
}

func (e *ConfigFileType) UnmarshalTOML(value any) error {
	sValue, ok := value.(string)
	if !ok {
		return fmt.Errorf("value %v is not a string type", value)
	}
	for _, enum := range []ConfigFileType{Static, Template} {
		if enum.Value() == sValue {
			*e = enum
			return nil
		}
	}
	return errors.New("invalid enum value")
}

type Profile struct {
	Name       string
	ConfigFile string           `toml:"config_file"`
	ConfigType *ConfigFileType  `toml:"config_file_type"`
	Conditions ProfileCondition `toml:"conditions"`
}

type PowerStateType int

const (
	BAT PowerStateType = iota
	AC
)

func (e *PowerStateType) Value() string {
	switch *e {
	case BAT:
		return "BAT"
	case AC:
		return "AC"
	}
	return ""
}

func (e *PowerStateType) UnmarshalTOML(value any) error {
	sValue, ok := value.(string)
	if !ok {
		return fmt.Errorf("value %v is not a string type", value)
	}
	for _, enum := range []PowerStateType{BAT, AC} {
		if enum.Value() == sValue {
			*e = enum
			return nil
		}
	}
	return errors.New("invalid enum value")
}

type ProfileCondition struct {
	RequiredMonitors []*RequiredMonitor `toml:"required_monitors"`
	PowerState       *PowerStateType    `toml:"power_state"`
}

type RequiredMonitor struct {
	Name        *string `toml:"name"`
	Description *string `toml:"description"`
	MonitorTag  *string `toml:"monitor_tag"`
}

func Load(configPath string) (*Config, error) {
	configPath = os.ExpandEnv(configPath)
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("configuration file %s not found", configPath)
	}

	absConfig, err := filepath.Abs(filepath.Dir(configPath))
	if err != nil {
		return nil, fmt.Errorf("cant convert config bath to abs %w", err)
	}

	var config Config
	config.configPath = absConfig
	if _, err := toml.DecodeFile(configPath, &config); err != nil {
		return nil, fmt.Errorf("failed to decode TOML: %w", err)
	}

	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &config, nil
}

func (c *Config) Validate() error {
	if len(c.Profiles) == 0 {
		return errors.New("no profiles defined")
	}

	if c.General == nil {
		c.General = &GeneralSection{}
	}
	if err := c.General.Validate(); err != nil {
		return fmt.Errorf("general section validation failed: %w", err)
	}

	if c.Scoring == nil {
		c.Scoring = &ScoringSection{}
	}
	if err := c.Scoring.Validate(); err != nil {
		return fmt.Errorf("scoring section validation failed: %w", err)
	}

	for name, profile := range c.Profiles {
		profile.Name = name
		if err := profile.Validate(c.configPath); err != nil {
			return fmt.Errorf("profile %s validation failed: %w", name, err)
		}
	}

	if c.PowerEvents == nil {
		c.PowerEvents = &PowerSection{}
	}
	if err := c.PowerEvents.Validate(); err != nil {
		return fmt.Errorf("power events section validation failed: %w", err)
	}

	return nil
}

func (g *GeneralSection) Validate() error {
	if g.Destination == nil {
		defaultDest := "$HOME/.config/hypr/monitors.conf"
		g.Destination = &defaultDest
	}

	dest := os.ExpandEnv(*g.Destination)
	g.Destination = &dest

	return nil
}

func (s *ScoringSection) Validate() error {
	defaultScore := 1

	if s.NameMatch == nil {
		s.NameMatch = &defaultScore
	}
	if s.DescriptionMatch == nil {
		s.DescriptionMatch = &defaultScore
	}
	if s.PowerStateMatch == nil {
		s.PowerStateMatch = &defaultScore
	}

	fields := []int{*s.DescriptionMatch, *s.NameMatch, *s.PowerStateMatch}
	for _, field := range fields {
		if 1 > field {
			return errors.New("scoring section validation failed, score needs to be > 1")
		}
	}

	return nil
}

func (p *Profile) Validate(configPath string) error {
	if p.ConfigFile == "" {
		return errors.New("config_file is required")
	}

	if p.ConfigType == nil {
		defaultType := Static
		p.ConfigType = &defaultType
	}

	if !strings.HasPrefix(p.ConfigFile, "/") {
		p.ConfigFile = filepath.Join(configPath, p.ConfigFile)
	}

	logrus.WithFields(logrus.Fields{
		"profile":     p.Name,
		"config_file": p.ConfigFile,
	}).Debug("Profile config file resolved")

	p.ConfigFile = os.ExpandEnv(p.ConfigFile)

	if _, err := os.Stat(p.ConfigFile); os.IsNotExist(err) {
		return fmt.Errorf("config file %s not found", p.ConfigFile)
	}

	if err := p.Conditions.Validate(); err != nil {
		return fmt.Errorf("conditions validation failed: %w", err)
	}

	return nil
}

func (pc *ProfileCondition) Validate() error {
	if len(pc.RequiredMonitors) == 0 {
		return errors.New("at least one required_monitor must be specified")
	}

	for i, monitor := range pc.RequiredMonitors {
		if err := monitor.Validate(); err != nil {
			return fmt.Errorf("required_monitor[%d] validation failed: %w", i, err)
		}
	}

	return nil
}

func (rm *RequiredMonitor) Validate() error {
	if rm.Name == nil && rm.Description == nil {
		return errors.New("at least one of name, or description must be specified")
	}

	return nil
}

func (ps *PowerSection) Validate() error {
	if ps.Disabled == nil {
		ps.Disabled = utils.BoolPtr(false)
	}
	if len(ps.DbusSignalMatchRules) == 0 {
		ps.DbusSignalMatchRules = []*DbusSignalMatchRule{
			{
				Sender:     utils.StringPtr("org.freedesktop.UPower"),
				Interface:  utils.StringPtr("org.freedesktop.UPower"),
				Member:     utils.StringPtr("DeviceAdded"),
				ObjectPath: utils.StringPtr("/org/freedesktop/UPower"),
			},
			{
				Sender:     utils.StringPtr("org.freedesktop.UPower"),
				Interface:  utils.StringPtr("org.freedesktop.UPower"),
				Member:     utils.StringPtr("DeviceRemoved"),
				ObjectPath: utils.StringPtr("/org/freedesktop/UPower"),
			},
			{
				Sender:     utils.StringPtr("org.freedesktop.UPower"),
				Interface:  utils.StringPtr("org.freedesktop.UPower.Properties"),
				Member:     utils.StringPtr("PropertiesChanged"),
				ObjectPath: utils.StringPtr("/org/freedesktop/UPower"),
			},
		}
	}

	for _, rule := range ps.DbusSignalMatchRules {
		if err := rule.Validate(); err != nil {
			return fmt.Errorf("one of the dbus match rules is invalid: %w", err)
		}
	}

	if ps.DbusSignalReceiveFilters == nil {
		ps.DbusSignalReceiveFilters = []*DbusSignalReceiveFilter{
			{Name: utils.StringPtr("org.freedesktop.DBus.Properties.PropertiesChanged")},
			{Name: utils.StringPtr("org.freedesktop.UPower.DeviceAdded")},
			{Name: utils.StringPtr("org.freedesktop.UPower.DeviceRemoved")},
		}
	}

	for _, signalFilter := range ps.DbusSignalReceiveFilters {
		if err := signalFilter.Validate(); err != nil {
			return fmt.Errorf("one of the dbus receive filter is invalid: %w", err)
		}
	}

	return nil
}

func (dr *DbusSignalMatchRule) Validate() error {
	if dr.Interface == nil && dr.Sender == nil && dr.Member == nil && dr.ObjectPath == nil {
		return errors.New("dbus rule cant be empty")
	}

	return nil
}

func (d *DbusSignalReceiveFilter) Validate() error {
	if d.Name == nil {
		return errors.New("name cant be emtpy")
	}

	return nil
}
