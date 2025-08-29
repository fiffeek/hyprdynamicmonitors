// Package config handles loading and validation of TOML configuration files.
package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/fiffeek/hyprdynamicmonitors/internal/utils"
	"github.com/sirupsen/logrus"
)

const LeaveEmpty = "leaveEmptyToken"

type Config struct {
	cfg  *UnsafeConfig
	path string
	mu   sync.RWMutex
}

func NewConfig(path string) (*Config, error) {
	cfg := &Config{
		cfg:  nil,
		path: path,
		mu:   sync.RWMutex{},
	}
	logrus.WithFields(logrus.Fields{"path": path}).Debug("Creating config wrapper")
	if err := cfg.Reload(); err != nil {
		return nil, fmt.Errorf("cant initialize config: %w", err)
	}
	return cfg, nil
}

func (c *Config) Get() *UnsafeConfig {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.cfg
}

func (c *Config) Reload() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	cfg, err := Load(c.path)
	if err != nil {
		return fmt.Errorf("cant reload config from %s: %w", c.path, err)
	}
	c.cfg = cfg
	return nil
}

type UnsafeConfig struct {
	ConfigDirPath        string
	ConfigPath           string
	Profiles             map[string]*Profile `toml:"profiles"`
	General              *GeneralSection     `toml:"general"`
	Scoring              *ScoringSection     `toml:"scoring"`
	PowerEvents          *PowerSection       `toml:"power_events"`
	HotReload            *HotReloadSection   `toml:"hot_reload_section"`
	Notifications        *Notifications      `toml:"notifications"`
	StaticTemplateValues map[string]string   `toml:"static_template_values"`
}

type HotReloadSection struct {
	UpdateDebounceTimer *int `toml:"debounce_time_ms"`
}

type Notifications struct {
	Disabled  *bool  `toml:"disabled"`
	TimeoutMs *int32 `toml:"timeout_ms"`
}

type PowerSection struct {
	DbusSignalMatchRules     []*DbusSignalMatchRule     `toml:"dbus_signal_match_rules"`
	DbusSignalReceiveFilters []*DbusSignalReceiveFilter `toml:"dbus_signal_receive_filters"`
	DbusQueryObject          *DbusQueryObject           `toml:"dbus_query_object"`
}

type DbusQueryObject struct {
	Destination              string               `toml:"destination"`
	Path                     string               `toml:"path"`
	Method                   string               `toml:"method"`
	Args                     []DbusQueryObjectArg `toml:"args"`
	ExpectedDischargingValue string               `toml:"expected_discharging_value"`
}

type DbusQueryObjectArg struct {
	Arg string `toml:"arg"`
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
	Destination    *string `toml:"destination"`
	DebounceTimeMs *int    `toml:"debounce_time_ms"`
}

type ScoringSection struct {
	NameMatch        *int `toml:"name_match"`
	DescriptionMatch *int `toml:"description_match"`
	PowerStateMatch  *int `toml:"power_state_match"`
}

var reservedTemplateVariables = map[string]bool{
	"MonitorsByTag": true,
	"Monitors":      true,
	"PowerState":    true,
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

func (e *ConfigFileType) MarshalTOML() ([]byte, error) {
	return []byte("\"" + e.Value() + "\""), nil
}

type Profile struct {
	Name                 string
	ConfigFileModTime    time.Time
	ConfigFileDir        string
	ConfigFile           string            `toml:"config_file"`
	ConfigType           *ConfigFileType   `toml:"config_file_type"`
	Conditions           ProfileCondition  `toml:"conditions"`
	StaticTemplateValues map[string]string `toml:"static_template_values"`
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

func (e *PowerStateType) MarshalTOML() ([]byte, error) {
	return []byte("\"" + e.Value() + "\""), nil
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

func Load(configPath string) (*UnsafeConfig, error) {
	configPath = os.ExpandEnv(configPath)
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("configuration file %s not found", configPath)
	}

	logrus.WithFields(logrus.Fields{"expanded": configPath}).Debug("Expnaded config path")

	absConfig, err := filepath.Abs(configPath)
	if err != nil {
		return nil, fmt.Errorf("cant convert config bath to abs %w", err)
	}

	logrus.WithFields(logrus.Fields{"abs": absConfig}).Debug("Found absolute config path")

	var config UnsafeConfig
	if _, err := toml.DecodeFile(configPath, &config); err != nil {
		return nil, fmt.Errorf("failed to decode TOML: %w", err)
	}

	config.ConfigPath = absConfig
	config.ConfigDirPath = filepath.Dir(config.ConfigPath)

	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	logrus.WithFields(logrus.Fields{"path": config.ConfigPath, "dir": config.ConfigDirPath}).Debug("Config is valid")

	return &config, nil
}

func (c *UnsafeConfig) Validate() error {
	if c.ConfigPath == "" {
		return errors.New("config path cant be empty")
	}

	if c.ConfigDirPath == "" {
		return errors.New("config dir path cant be empty")
	}

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
		if err := profile.Validate(c.ConfigDirPath); err != nil {
			return fmt.Errorf("profile %s validation failed: %w", name, err)
		}
	}

	if c.PowerEvents == nil {
		c.PowerEvents = &PowerSection{}
	}
	if err := c.PowerEvents.Validate(); err != nil {
		return fmt.Errorf("power events section validation failed: %w", err)
	}

	if c.Notifications == nil {
		c.Notifications = &Notifications{}
	}
	if err := c.Notifications.Validate(); err != nil {
		return fmt.Errorf("notifications section validation failed: %w", err)
	}

	if c.HotReload == nil {
		c.HotReload = &HotReloadSection{}
	}
	if err := c.HotReload.Validate(); err != nil {
		return fmt.Errorf("hot reload section validation failed: %w", err)
	}

	for key := range c.StaticTemplateValues {
		if _, ok := reservedTemplateVariables[key]; ok {
			return errors.New("key " + key + " cant be used since it is a reserved keyword")
		}
	}

	return nil
}

func (h *HotReloadSection) Validate() error {
	if h.UpdateDebounceTimer == nil {
		h.UpdateDebounceTimer = utils.IntPtr(1000)
	}
	return nil
}

func (n *Notifications) Validate() error {
	if n.Disabled == nil {
		n.Disabled = utils.BoolPtr(false)
	}
	if n.TimeoutMs == nil {
		n.TimeoutMs = utils.JustPtr[int32](10000)
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

	if g.DebounceTimeMs == nil {
		g.DebounceTimeMs = utils.IntPtr(3000)
	}

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
	absConfigFile, err := filepath.Abs(p.ConfigFile)
	if err != nil {
		return fmt.Errorf("cant get absolute path to config file %s: %w", p.ConfigFile, err)
	}

	p.ConfigFile = absConfigFile

	fi, err := os.Stat(p.ConfigFile)
	if os.IsNotExist(err) {
		return fmt.Errorf("config file %s not found", p.ConfigFile)
	}

	p.ConfigFileDir = filepath.Dir(p.ConfigFile)
	p.ConfigFileModTime = fi.ModTime()

	if err := p.Conditions.Validate(); err != nil {
		return fmt.Errorf("conditions validation failed: %w", err)
	}

	for key := range p.StaticTemplateValues {
		if _, ok := reservedTemplateVariables[key]; ok {
			return errors.New("key " + key + " cant be used since it is a reserved keyword")
		}
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
	if len(ps.DbusSignalMatchRules) == 0 {
		// listen to
		// gdbus monitor -y -d org.freedesktop.UPower | grep -E "PropertiesChanged|Device(Added|Removed)"
		// to see the events
		// e.g. /org/freedesktop/UPower/devices/line_power_ACAD: org.freedesktop.DBus.Properties.PropertiesChanged ('org.freedesktop.UPower.Device', {'UpdateTime': <uint64 1756242314>, 'Online': <true>}, @as [])
		ps.DbusSignalMatchRules = []*DbusSignalMatchRule{
			{},
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
		}
	}

	for _, signalFilter := range ps.DbusSignalReceiveFilters {
		if err := signalFilter.Validate(); err != nil {
			return fmt.Errorf("one of the dbus receive filter is invalid: %w", err)
		}
	}

	if ps.DbusQueryObject == nil {
		ps.DbusQueryObject = &DbusQueryObject{}
	}

	if err := ps.DbusQueryObject.Validate(); err != nil {
		return fmt.Errorf("dbus query object for the battery stats is invalid: %w", err)
	}

	return nil
}

func (d *DbusQueryObject) CollectArgs() []interface{} {
	args := []any{}
	for _, arg := range d.Args {
		args = append(args, arg.Arg)
	}
	return args
}

func (d *DbusQueryObject) Validate() error {
	// dbus-send --system --print-reply \
	//   --dest=org.freedesktop.UPower \
	//   /org/freedesktop/UPower/devices/line_power_ACAD \
	//   org.freedesktop.DBus.Properties.Get \
	//   string:org.freedesktop.UPower.Device \
	//   string:Online

	if d.Destination == "" {
		d.Destination = "org.freedesktop.UPower"
	}
	if d.Method == "" {
		d.Method = "org.freedesktop.DBus.Properties.Get"
	}
	if d.Path == "" {
		d.Path = "/org/freedesktop/UPower/devices/line_power_ACAD"
	}
	if len(d.Args) == 0 {
		d.Args = []DbusQueryObjectArg{
			{Arg: "org.freedesktop.UPower.Device"},
			{Arg: "Online"},
		}
	}
	if d.ExpectedDischargingValue == "" {
		d.ExpectedDischargingValue = "false"
	}
	for _, arg := range d.Args {
		if arg.Arg == "" {
			return errors.New("arg cant be empty")
		}
	}
	return nil
}

func (dr *DbusSignalMatchRule) Validate() error {
	if dr.Interface != nil && *dr.Interface == LeaveEmpty {
		dr.Interface = nil
	} else if dr.Interface == nil {
		dr.Interface = utils.StringPtr("org.freedesktop.DBus.Properties")
	}
	if dr.Member != nil && *dr.Member == LeaveEmpty {
		dr.Member = nil
	} else if dr.Member == nil {
		dr.Member = utils.StringPtr("PropertiesChanged")
	}
	if dr.ObjectPath != nil && *dr.ObjectPath == LeaveEmpty {
		dr.ObjectPath = nil
	} else if dr.ObjectPath == nil {
		dr.ObjectPath = utils.StringPtr("/org/freedesktop/UPower/devices/line_power_ACAD")
	}
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
