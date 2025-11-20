// Package config handles loading and validation of TOML configuration files.
package config

import (
	"bytes"
	_ "embed"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"sync"
	"text/template"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/fiffeek/hyprdynamicmonitors/internal/utils"
	"github.com/sirupsen/logrus"
)

//go:embed templates/default_config.toml.go.tmpl
var defaultConfigTemplate string

const LeaveEmpty = "leaveEmptyToken"

type Config struct {
	cfg  *RawConfig
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

func (c *Config) Get() *RawConfig {
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

type RawConfig struct {
	ConfigDirPath        string              `toml:"-"`
	ConfigPath           string              `toml:"-"`
	Profiles             map[string]*Profile `toml:"profiles"`
	FallbackProfile      *Profile            `toml:"fallback_profile"`
	General              *GeneralSection     `toml:"general"`
	Scoring              *ScoringSection     `toml:"scoring"`
	PowerEvents          *PowerSection       `toml:"power_events"`
	LidEvents            *LidSection         `toml:"lid_events"`
	HotReload            *HotReloadSection   `toml:"hot_reload_section"`
	Notifications        *Notifications      `toml:"notifications"`
	StaticTemplateValues map[string]string   `toml:"static_template_values"`
	KeysOrder            []string            `toml:"-"`
	TUISection           *TUISection         `toml:"tui"`
}

type TUISection struct {
	Colors *TUIColors `toml:"colors"`
}

type TUIColors struct {
	SourceFile    *string `toml:"source"`
	SourceFileDir *string `toml:"-"`

	// Pane borders
	ActivePaneColor   *string `toml:"active_pane_color"`
	InactivePaneColor *string `toml:"inactive_pane_color"`

	// Header colors
	HeaderColor          *string `toml:"header_color"`
	HeaderIndicatorColor *string `toml:"header_indicator_color"`
	HeaderIndicatorBg    *string `toml:"header_indicator_bg"`

	// Program name in the top bar
	ProgramNameColor *string `toml:"program_name_color"`
	ProgramNameBg    *string `toml:"program_name_bg"`

	// Status colors
	ErrorColor   *string `toml:"error_color"`
	ErrorBg      *string `toml:"error_bg"`
	SuccessColor *string `toml:"success_color"`
	SuccessBg    *string `toml:"success_bg"`
	WarningColor *string `toml:"warning_color"`
	MutedColor   *string `toml:"muted_color"`

	// Tab colors
	TabActiveColor   *string `toml:"tab_active_color"`
	TabActiveBg      *string `toml:"tab_active_bg"`
	TabInactiveColor *string `toml:"tab_inactive_color"`
	TabInactiveBg    *string `toml:"tab_inactive_bg"`

	// Monitor lists colors
	MonitorColorModeColor     *string `toml:"monitor_color_mode_color"`
	MonitorScaleModeColor     *string `toml:"monitor_scale_mode_color"`
	MonitorMirroringModeColor *string `toml:"monitor_mirroring_mode_color"`
	MonitorEditingModeColor   *string `toml:"monitor_editing_mode_color"`
	MonitorModeSelectionColor *string `toml:"monitor_mode_selection_color"`

	// Lists
	ListItemSubtitleColor   *string `toml:"list_item_subtitle_color"`
	ListItemUnselectedColor *string `toml:"list_item_unselected_color"`
	ListItemSelectedColor   *string `toml:"list_item_selected_color"`

	// Legend and selection
	LegendColor                *string `toml:"legend_color"`
	LegendSelectedMonitorColor *string `toml:"legend_selected_monitor_color"`
	LegendSelectedMonitorBg    *string `toml:"legend_selected_monitor_bg"`

	// Title and subtitle colors
	TitleColor    *string `toml:"title_color"`
	SubtitleColor *string `toml:"subtitle_color"`
	InfoColor     *string `toml:"info_color"`
	LinkColor     *string `toml:"link_color"`

	// Grid and monitor edge colors
	GridDotColor                           *string   `toml:"grid_dot_color"`
	MonitorEdgeColors                      *[]string `toml:"monitor_edge_colors"`
	MonitorFillColorsUnselected            *[]string `toml:"monitor_fill_colors_unselected"`
	MonitorFillColorsSelected              *[]string `toml:"monitor_fill_colors_selected"`
	MonitorBottomColors                    *[]string `toml:"monitor_bottom_colors"`
	MonitorPreviewPaneLabelBackgroundColor *string   `toml:"monitor_preview_pane_label_bg"`
	GridSnappingLineColor                  *string   `toml:"grid_snapping_line_color"`

	// Help
	HelpKeyColor         *string `toml:"help_key_color"`
	HelpDescriptionColor *string `toml:"help_description_color"`
	HelpSeparatorColor   *string `toml:"help_separator_color"`
}

type HotReloadSection struct {
	UpdateDebounceTimer *int `toml:"debounce_time_ms"`
}

type Notifications struct {
	Disabled  *bool  `toml:"disabled"`
	TimeoutMs *int32 `toml:"timeout_ms"`
}

type LidSection struct {
	DbusSignalMatchRules     []*DbusSignalMatchRule     `toml:"dbus_signal_match_rules"`
	DbusSignalReceiveFilters []*DbusSignalReceiveFilter `toml:"dbus_signal_receive_filters"`
	DbusQueryObject          *DbusQueryObject           `toml:"dbus_query_object"`
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
	ExpectedLidClosingValue  string               `toml:"expected_lid_closing_value"`
}

type DbusQueryObjectArg struct {
	Arg string `toml:"arg"`
}

type DbusSignalReceiveFilter struct {
	Name *string `toml:"name"`
	Body *string `toml:"body"`
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
	PostApplyExec  *string `toml:"post_apply_exec"`
	PreApplyExec   *string `toml:"pre_apply_exec"`
}

type ScoringSection struct {
	NameMatch        *int `toml:"name_match"`
	DescriptionMatch *int `toml:"description_match"`
	PowerStateMatch  *int `toml:"power_state_match"`
	LidStateMatch    *int `toml:"lid_state_match"`
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

func (e ConfigFileType) Value() string {
	switch e {
	case Static:
		return "static"
	case Template:
		return "template"
	}
	return ""
}

var allConfigFileTypes = []ConfigFileType{Static, Template}

func (e *ConfigFileType) UnmarshalTOML(value any) error {
	sValue, ok := value.(string)
	if !ok {
		return fmt.Errorf("value %v is not a string type", value)
	}
	for _, enum := range allConfigFileTypes {
		if enum.Value() == sValue {
			*e = enum
			return nil
		}
	}
	return fmt.Errorf("invalid enum value, expecting one of %s",
		utils.FormatEnumTypes(allConfigFileTypes))
}

func (e *ConfigFileType) MarshalTOML() ([]byte, error) {
	return []byte("\"" + e.Value() + "\""), nil
}

type Profile struct {
	Name                 string            `toml:"-"`
	ConfigFileModTime    time.Time         `toml:"-"`
	ConfigFileDir        string            `toml:"-"`
	ConfigFile           string            `toml:"config_file"`
	ConfigType           *ConfigFileType   `toml:"config_file_type"`
	Conditions           *ProfileCondition `toml:"conditions"`
	StaticTemplateValues map[string]string `toml:"static_template_values"`
	IsFallbackProfile    bool              `toml:"-"`
	PostApplyExec        *string           `toml:"post_apply_exec"`
	PreApplyExec         *string           `toml:"pre_apply_exec"`
	KeyOrder             int               `toml:"-"`
}

type LidStateType int

const (
	UnknownLidStateType LidStateType = iota
	OpenedLidStateType
	ClosedLidStateType
)

func (e LidStateType) Value() string {
	switch e {
	case OpenedLidStateType:
		return "Opened"
	case ClosedLidStateType:
		return "Closed"
	default:
		return "UNKNOWN"
	}
}

var allLidStateTypes = []LidStateType{OpenedLidStateType, ClosedLidStateType}

func (e *LidStateType) UnmarshalTOML(value any) error {
	sValue, ok := value.(string)
	if !ok {
		return fmt.Errorf("value %v is not a string type", value)
	}
	for _, enum := range allLidStateTypes {
		if enum.Value() == sValue {
			*e = enum
			return nil
		}
	}
	return fmt.Errorf("invalid enum value, expecting one of %s",
		utils.FormatEnumTypes(allLidStateTypes))
}

func (e *LidStateType) MarshalTOML() ([]byte, error) {
	return []byte("\"" + e.Value() + "\""), nil
}

type PowerStateType int

const (
	BAT PowerStateType = iota
	AC
)

func (e PowerStateType) Value() string {
	switch e {
	case BAT:
		return "BAT"
	case AC:
		return "AC"
	}
	return ""
}

var allPowerStateTypes = []PowerStateType{BAT, AC}

func (e *PowerStateType) UnmarshalTOML(value any) error {
	sValue, ok := value.(string)
	if !ok {
		return fmt.Errorf("value %v is not a string type", value)
	}
	for _, enum := range allPowerStateTypes {
		if enum.Value() == sValue {
			*e = enum
			return nil
		}
	}
	return fmt.Errorf("invalid enum value, expecting one of %s",
		utils.FormatEnumTypes(allPowerStateTypes))
}

func (e *PowerStateType) MarshalTOML() ([]byte, error) {
	return []byte("\"" + e.Value() + "\""), nil
}

type ProfileCondition struct {
	RequiredMonitors []*RequiredMonitor `toml:"required_monitors"`
	PowerState       *PowerStateType    `toml:"power_state"`
	LidState         *LidStateType      `toml:"lid_state"`
}

type RequiredMonitor struct {
	Name                       *string        `toml:"name"`
	Description                *string        `toml:"description"`
	MonitorTag                 *string        `toml:"monitor_tag"`
	MatchDescriptionUsingRegex *bool          `toml:"match_description_using_regex"`
	MatchNameUsingRegex        *bool          `toml:"match_name_using_regex"`
	DescriptionRegex           *regexp.Regexp `toml:"-"`
	NameRegex                  *regexp.Regexp `toml:"-"`
}

func Cond(cond, a *string, b string) string {
	if cond != nil {
		return *a
	}
	return b
}

func CreateDefaultConfig(path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		return fmt.Errorf("cant create directory: %w", err)
	}

	objectPath, err := utils.GetPowerLine()
	if err != nil {
		logrus.Warning("No power line available, will use a default")
	}

	funcMap := template.FuncMap{
		"cond": Cond,
	}

	tmpl, err := template.New("hdm_config").Funcs(funcMap).Parse(defaultConfigTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	templateData := map[string]any{
		"ObjectPath": objectPath,
	}

	var rendered bytes.Buffer
	if err := tmpl.Execute(&rendered, templateData); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	if err := utils.WriteAtomic(path, rendered.Bytes()); err != nil {
		return fmt.Errorf("cant write to the config file: %w", err)
	}

	return nil
}

func Load(configPath string) (*RawConfig, error) {
	configPath = os.ExpandEnv(configPath)
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		if err := CreateDefaultConfig(configPath); err != nil {
			return nil, fmt.Errorf("cant create a default configuration %s: %w", configPath, err)
		}
	}

	logrus.WithFields(logrus.Fields{"expanded": configPath}).Debug("Expanded config path")

	absConfig, err := filepath.Abs(configPath)
	if err != nil {
		return nil, fmt.Errorf("cant convert config path to absolute path %w", err)
	}

	logrus.WithFields(logrus.Fields{"abs": absConfig}).Debug("Found absolute config path")

	// nolint:gosec
	contents, err := os.ReadFile(absConfig)
	if err != nil {
		return nil, fmt.Errorf("cant read config file %s: %w", absConfig, err)
	}
	logrus.Debugf("Config contents: %s", contents)

	var config RawConfig
	m, err := toml.DecodeFile(configPath, &config)
	if err != nil {
		return nil, fmt.Errorf("failed to decode TOML: %w", err)
	}
	keys := []string{}
	for _, k := range m.Keys() {
		keys = append(keys, strings.Join(k, "."))
	}

	config.ConfigPath = absConfig
	config.ConfigDirPath = filepath.Dir(config.ConfigPath)
	config.KeysOrder = keys

	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	logrus.WithFields(logrus.Fields{"path": config.ConfigPath, "dir": config.ConfigDirPath}).Debug("Config is valid")

	return &config, nil
}

// OrderedProfileKeys returns the profile names in the order they appear in the toml file
func (c *RawConfig) OrderedProfileKeys() []string {
	profileNames := make([]string, 0, len(c.Profiles))
	for name := range c.Profiles {
		profileNames = append(profileNames, name)
	}

	slices.SortFunc(profileNames, func(a, b string) int {
		orderA := c.Profiles[a].KeyOrder
		orderB := c.Profiles[b].KeyOrder
		return orderA - orderB
	})

	return profileNames
}

func (c *RawConfig) Validate() error {
	if c.ConfigPath == "" {
		return errors.New("config path cant be empty")
	}

	if c.ConfigDirPath == "" {
		return errors.New("config dir path cant be empty")
	}

	if len(c.Profiles) == 0 {
		logrus.Warning("No profiles are defined, not config will be automatically applied")
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
		profile.IsFallbackProfile = false
		profile.KeyOrder = slices.Index(c.KeysOrder, "profiles."+name)
		logrus.Debugf("Profile %s has order %d", profile.Name, profile.KeyOrder)
		if err := profile.Validate(c.ConfigDirPath); err != nil {
			return fmt.Errorf("profile %s validation failed: %w", name, err)
		}
	}

	if c.FallbackProfile != nil {
		c.FallbackProfile.Name = "fallback"
		c.FallbackProfile.IsFallbackProfile = true
		if err := c.FallbackProfile.Validate(c.ConfigDirPath); err != nil {
			return fmt.Errorf("fallback profile validation failed: %w", err)
		}
	}

	if c.PowerEvents == nil {
		c.PowerEvents = &PowerSection{}
	}
	if err := c.PowerEvents.Validate(); err != nil {
		return fmt.Errorf("power events section validation failed: %w", err)
	}

	if c.LidEvents == nil {
		c.LidEvents = &LidSection{}
	}
	if err := c.LidEvents.Validate(); err != nil {
		return fmt.Errorf("lid events section validation failed: %w", err)
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

	if c.TUISection == nil {
		c.TUISection = &TUISection{}
	}
	if err := c.TUISection.Validate(c.ConfigDirPath); err != nil {
		return fmt.Errorf("tui section validation failed: %w", err)
	}

	return nil
}

func (t *TUISection) Validate(configDirPath string) error {
	if t.Colors == nil {
		t.Colors = &TUIColors{}
	}
	if t.Colors.SourceFile != nil {
		logrus.WithFields(logrus.Fields{"theme": *t.Colors.SourceFile}).Info("Sourcing theme file")
		if !strings.HasPrefix(*t.Colors.SourceFile, "/") &&
			!strings.HasPrefix(*t.Colors.SourceFile, "$") &&
			!strings.HasPrefix(*t.Colors.SourceFile, "~") {
			t.Colors.SourceFile = utils.JustPtr(filepath.Join(configDirPath, *t.Colors.SourceFile))
		}

		t.Colors.SourceFile = utils.JustPtr(os.ExpandEnv(*t.Colors.SourceFile))

		if strings.HasPrefix(*t.Colors.SourceFile, "~/") {
			homeDir, err := os.UserHomeDir()
			if err != nil {
				return fmt.Errorf("cant get user home directory: %w", err)
			}
			t.Colors.SourceFile = utils.JustPtr(filepath.Join(homeDir, (*t.Colors.SourceFile)[2:]))
		}

		absConfigFile, err := filepath.Abs(*t.Colors.SourceFile)
		if err != nil {
			return fmt.Errorf("cant get absolute path to colors file %s: %w", *t.Colors.SourceFile, err)
		}

		t.Colors.SourceFile = utils.JustPtr(absConfigFile)

		contents, err := os.ReadFile(*t.Colors.SourceFile)
		if err != nil {
			return fmt.Errorf("cant read colors file %s: %w", *t.Colors.SourceFile, err)
		}
		logrus.Debugf("Config contents: %s", contents)

		var config TUIColors
		_, err = toml.DecodeFile(*t.Colors.SourceFile, &config)
		if err != nil {
			return fmt.Errorf("failed to decode TOML: %w", err)
		}

		config.SourceFileDir = utils.JustPtr(filepath.Dir(*t.Colors.SourceFile))
		t.Colors = &config
	}
	if err := t.Colors.Validate(); err != nil {
		return fmt.Errorf("tui colors validation failed: %w", err)
	}

	return nil
}

func (t *TUIColors) Validate() error {
	// Pane borders
	if t.ActivePaneColor == nil {
		t.ActivePaneColor = utils.JustPtr("62")
	}
	if t.InactivePaneColor == nil {
		t.InactivePaneColor = utils.JustPtr("240")
	}

	// Header colors
	if t.HeaderColor == nil {
		t.HeaderColor = utils.JustPtr("205")
	}
	if t.HeaderIndicatorColor == nil {
		t.HeaderIndicatorColor = utils.JustPtr("226")
	}
	if t.HeaderIndicatorBg == nil {
		t.HeaderIndicatorBg = utils.JustPtr("235")
	}

	// Prog name
	if t.ProgramNameColor == nil {
		t.ProgramNameColor = utils.JustPtr("0")
	}
	if t.ProgramNameBg == nil {
		t.ProgramNameBg = utils.JustPtr("250")
	}

	// Status colors
	if t.ErrorColor == nil {
		t.ErrorColor = utils.JustPtr("196")
	}
	if t.ErrorBg == nil {
		t.ErrorBg = utils.JustPtr("235")
	}
	if t.SuccessColor == nil {
		t.SuccessColor = utils.JustPtr("82")
	}
	if t.SuccessBg == nil {
		t.SuccessBg = utils.JustPtr("235")
	}

	// Tab colors
	if t.TabActiveColor == nil {
		t.TabActiveColor = utils.JustPtr("255")
	}
	if t.TabActiveBg == nil {
		t.TabActiveBg = utils.JustPtr("105")
	}
	if t.TabInactiveColor == nil {
		t.TabInactiveColor = utils.JustPtr("243")
	}
	if t.TabInactiveBg == nil {
		t.TabInactiveBg = utils.JustPtr("235")
	}

	// Monitor/Item colors
	if t.ListItemSubtitleColor == nil {
		t.ListItemSubtitleColor = utils.JustPtr("180")
	}
	if t.MonitorColorModeColor == nil {
		t.MonitorColorModeColor = utils.JustPtr("15")
	}
	if t.MonitorScaleModeColor == nil {
		t.MonitorScaleModeColor = utils.JustPtr("39")
	}
	if t.MonitorMirroringModeColor == nil {
		t.MonitorMirroringModeColor = utils.JustPtr("140")
	}
	if t.MonitorEditingModeColor == nil {
		t.MonitorEditingModeColor = utils.JustPtr("226")
	}
	if t.MonitorModeSelectionColor == nil {
		t.MonitorModeSelectionColor = utils.JustPtr("208")
	}
	if t.MutedColor == nil {
		t.MutedColor = utils.JustPtr("243")
	}
	if t.ListItemUnselectedColor == nil {
		t.ListItemUnselectedColor = utils.JustPtr("255")
	}
	if t.ListItemSelectedColor == nil {
		t.ListItemSelectedColor = utils.JustPtr("212")
	}

	// Legend and selection
	if t.LegendColor == nil {
		t.LegendColor = utils.JustPtr("243")
	}
	if t.LegendSelectedMonitorColor == nil {
		t.LegendSelectedMonitorColor = utils.JustPtr("226")
	}
	if t.LegendSelectedMonitorBg == nil {
		t.LegendSelectedMonitorBg = utils.JustPtr("235")
	}

	// Title and subtitle colors
	if t.TitleColor == nil {
		t.TitleColor = utils.JustPtr("105")
	}
	if t.SubtitleColor == nil {
		t.SubtitleColor = utils.JustPtr("180")
	}
	if t.InfoColor == nil {
		t.InfoColor = utils.JustPtr("108")
	}
	if t.LinkColor == nil {
		t.LinkColor = utils.JustPtr("42")
	}

	// Grid and monitor edge colors
	if t.GridDotColor == nil {
		t.GridDotColor = utils.JustPtr("240")
	}
	if t.GridSnappingLineColor == nil {
		t.GridSnappingLineColor = utils.JustPtr("243")
	}
	if t.MonitorEdgeColors == nil {
		t.MonitorEdgeColors = &[]string{"105", "208", "39", "226", "196", "99"}
	}
	if t.MonitorPreviewPaneLabelBackgroundColor == nil {
		t.MonitorPreviewPaneLabelBackgroundColor = utils.JustPtr("#000000")
	}

	// Config pane colors
	if t.WarningColor == nil {
		t.WarningColor = utils.JustPtr("196")
	}

	// Monitor fill and bottom colors (parallel arrays with MonitorEdgeColors)
	// Default MonitorEdgeColors: ["105", "208", "39", "226", "196", "99"]
	if t.MonitorFillColorsUnselected == nil {
		t.MonitorFillColorsUnselected = &[]string{
			"99",  // 105: Bright purple -> darker purple
			"166", // 208: Bright orange -> darker orange
			"25",  // 39: Bright blue -> darker blue
			"220", // 226: Bright yellow -> darker golden yellow
			"124", // 196: Bright red -> darker red
			"60",  // 99: Bright pink -> darker magenta
		}
	}
	if t.MonitorFillColorsSelected == nil {
		t.MonitorFillColorsSelected = &[]string{
			"98",  // 105: Bright purple -> darker purple
			"94",  // 208: Bright orange -> much darker orange/brown
			"17",  // 39: Bright blue -> much darker blue
			"100", // 226: Bright yellow -> much darker gold
			"52",  // 196: Bright red -> much darker red/maroon
			"53",  // 99: Bright pink -> much darker magenta/purple
		}
	}
	if t.MonitorBottomColors == nil {
		t.MonitorBottomColors = &[]string{
			"141", // 105: Purple -> Bright purple
			"214", // 208: Orange -> Bright orange
			"45",  // 39: Blue -> Bright blue
			"228", // 226: Yellow -> Bright yellow
			"197", // 196: Red -> Bright red
			"105", // 99: Pink -> Bright pink
		}
	}

	// help
	if t.HelpKeyColor == nil {
		t.HelpKeyColor = utils.JustPtr("#909090")
	}
	if t.HelpDescriptionColor == nil {
		t.HelpDescriptionColor = utils.JustPtr("#4A4A4A")
	}
	if t.HelpSeparatorColor == nil {
		t.HelpSeparatorColor = utils.JustPtr("#3C3C3C")
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
	if s.LidStateMatch == nil {
		s.LidStateMatch = &defaultScore
	}

	fields := []int{*s.DescriptionMatch, *s.NameMatch, *s.PowerStateMatch, *s.LidStateMatch}
	for _, field := range fields {
		if 1 > field {
			return errors.New("scoring section validation failed, score needs to be > 1")
		}
	}

	return nil
}

func (p *Profile) SetPath(configPath string) error {
	if p.ConfigFile == "" {
		return errors.New("config_file is required")
	}

	if !strings.HasPrefix(p.ConfigFile, "/") && !strings.HasPrefix(p.ConfigFile, "$") && !strings.HasPrefix(p.ConfigFile, "~") {
		p.ConfigFile = filepath.Join(configPath, p.ConfigFile)
	}

	return nil
}

func (p *Profile) Validate(configPath string) error {
	if err := p.SetPath(configPath); err != nil {
		return fmt.Errorf("cant set config path: %w", err)
	}

	if p.ConfigType == nil {
		defaultType := Static
		p.ConfigType = &defaultType
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
	if os.IsNotExist(err) || fi == nil {
		return fmt.Errorf("config file %s not found", p.ConfigFile)
	}

	p.ConfigFileDir = filepath.Dir(p.ConfigFile)
	p.ConfigFileModTime = fi.ModTime()

	if p.Conditions == nil {
		p.Conditions = &ProfileCondition{}
	}

	if p.IsFallbackProfile && !p.Conditions.IsEmpty() {
		return errors.New("fallback profile cant define any conditions")
	}

	if err := p.Conditions.Validate(); err != nil && !p.IsFallbackProfile {
		return fmt.Errorf("conditions validation failed: %w", err)
	}

	for key := range p.StaticTemplateValues {
		if _, ok := reservedTemplateVariables[key]; ok {
			return errors.New("key " + key + " cant be used since it is a reserved keyword")
		}
	}

	return nil
}

func (pc *ProfileCondition) IsEmpty() bool {
	if pc == nil {
		return true
	}
	return len(pc.RequiredMonitors) == 0 && pc.PowerState == nil && pc.LidState == nil
}

func (pc *ProfileCondition) Validate() error {
	if pc == nil {
		return errors.New("profile conditions cant be empty")
	}

	if len(pc.RequiredMonitors) == 0 {
		return errors.New("at least one required_monitors must be specified")
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
	if rm.MatchDescriptionUsingRegex == nil {
		rm.MatchDescriptionUsingRegex = utils.JustPtr(false)
	}
	if rm.MatchNameUsingRegex == nil {
		rm.MatchNameUsingRegex = utils.JustPtr(false)
	}
	if *rm.MatchDescriptionUsingRegex && rm.Description != nil {
		r, err := regexp.Compile(*rm.Description)
		if err != nil {
			return fmt.Errorf("regex %s is not valid: %w", *rm.Description, err)
		}
		rm.DescriptionRegex = r
	}
	if *rm.MatchNameUsingRegex && rm.Name != nil {
		r, err := regexp.Compile(*rm.Name)
		if err != nil {
			return fmt.Errorf("regex %s is not valid: %w", *rm.Name, err)
		}
		rm.NameRegex = r
	}

	return nil
}

func (rm *RequiredMonitor) MatchName(name string) bool {
	if rm.Name == nil {
		return false
	}

	if rm.MatchNameUsingRegex == nil || !*rm.MatchNameUsingRegex {
		return name == *rm.Name
	}

	return rm.NameRegex.MatchString(name)
}

func (rm *RequiredMonitor) HasName() bool {
	return rm.Name != nil
}

func (rm *RequiredMonitor) MatchDescription(desc string) bool {
	if rm.Description == nil {
		return false
	}

	if rm.MatchDescriptionUsingRegex == nil || !*rm.MatchDescriptionUsingRegex {
		return desc == *rm.Description
	}

	return rm.DescriptionRegex.MatchString(desc)
}

func (rm *RequiredMonitor) HasDescription() bool {
	return rm.Description != nil
}

func (ls *LidSection) Validate() error {
	if len(ls.DbusSignalMatchRules) == 0 {
		ls.DbusSignalMatchRules = []*DbusSignalMatchRule{
			{},
		}
	}

	defaultInterface := "org.freedesktop.DBus.Properties"
	defaultMember := "PropertiesChanged"
	defaultObjectPath := "/org/freedesktop/UPower"
	for _, rule := range ls.DbusSignalMatchRules {
		if err := rule.Validate(defaultInterface, defaultMember, defaultObjectPath); err != nil {
			return fmt.Errorf("one of the dbus match rules is invalid: %w", err)
		}
	}

	if ls.DbusSignalReceiveFilters == nil {
		ls.DbusSignalReceiveFilters = []*DbusSignalReceiveFilter{
			{Name: utils.StringPtr("org.freedesktop.DBus.Properties.PropertiesChanged"), Body: utils.StringPtr("LidIsClosed")},
		}
	}

	for _, signalFilter := range ls.DbusSignalReceiveFilters {
		if err := signalFilter.Validate(); err != nil {
			return fmt.Errorf("one of the dbus receive filter is invalid: %w", err)
		}
	}

	if ls.DbusQueryObject == nil {
		ls.DbusQueryObject = &DbusQueryObject{}
	}

	defaultDestination := "org.freedesktop.UPower"
	defaultMethod := "org.freedesktop.DBus.Properties.Get"
	defaultPath := "/org/freedesktop/UPower"
	defaultArgs := []DbusQueryObjectArg{
		{Arg: "org.freedesktop.UPower"},
		{Arg: "LidIsClosed"},
	}
	defaultExpectedLidClosingValue := "true"
	if err := ls.DbusQueryObject.Validate(defaultDestination, defaultMethod, defaultPath, "",
		defaultArgs, defaultExpectedLidClosingValue); err != nil {
		return fmt.Errorf("dbus query object for the battery stats is invalid: %w", err)
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

	defaultPowerLine := "/org/freedesktop/UPower/devices/line_power_ACAD"
	powerLine, err := utils.GetPowerLine()
	if err != nil {
		logrus.WithError(err).Warningf("No power line available, will use a default: %s", defaultPowerLine)
		powerLine = utils.JustPtr(defaultPowerLine)
	}
	logrus.WithFields(logrus.Fields{"power_line": *powerLine}).Info(
		"Inferred power line")

	defaultInterface := "org.freedesktop.DBus.Properties"
	defaultMember := "PropertiesChanged"
	defaultObjectPath := *powerLine
	for _, rule := range ps.DbusSignalMatchRules {
		if err := rule.Validate(defaultInterface, defaultMember, defaultObjectPath); err != nil {
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

	defaultDestination := "org.freedesktop.UPower"
	defaultMethod := "org.freedesktop.DBus.Properties.Get"
	defaultPath := *powerLine
	defaultArgs := []DbusQueryObjectArg{
		{Arg: "org.freedesktop.UPower.Device"},
		{Arg: "Online"},
	}
	defaultExpectedDischargingValue := "false"
	if err := ps.DbusQueryObject.Validate(defaultDestination, defaultMethod, defaultPath,
		defaultExpectedDischargingValue, defaultArgs, ""); err != nil {
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

func (d *DbusQueryObject) Validate(defaultDestination, defaultMethod, defaultPath,
	defaultExpectedDischargingValue string,
	defaultArgs []DbusQueryObjectArg, defaultExpectedLidClosingValue string,
) error {
	// dbus-send --system --print-reply \
	//   --dest=org.freedesktop.UPower \
	//   /org/freedesktop/UPower/devices/line_power_ACAD \
	//   org.freedesktop.DBus.Properties.Get \
	//   string:org.freedesktop.UPower.Device \
	//   string:Online

	if d.Destination == "" {
		d.Destination = defaultDestination
	}
	if d.Method == "" {
		d.Method = defaultMethod
	}
	if d.Path == "" {
		d.Path = defaultPath
	}
	if len(d.Args) == 0 {
		d.Args = defaultArgs
	}
	if d.ExpectedDischargingValue == "" {
		d.ExpectedDischargingValue = defaultExpectedDischargingValue
	}
	if d.ExpectedLidClosingValue == "" {
		d.ExpectedLidClosingValue = defaultExpectedLidClosingValue
	}
	for _, arg := range d.Args {
		if arg.Arg == "" {
			return errors.New("arg cant be empty")
		}
	}
	return nil
}

func (dr *DbusSignalMatchRule) Validate(defaultInterface, defaultMember, defaultObjectPath string) error {
	if dr.Interface != nil && *dr.Interface == LeaveEmpty {
		dr.Interface = nil
	} else if dr.Interface == nil {
		dr.Interface = utils.StringPtr(defaultInterface)
	}
	if dr.Member != nil && *dr.Member == LeaveEmpty {
		dr.Member = nil
	} else if dr.Member == nil {
		dr.Member = utils.StringPtr(defaultMember)
	}
	if dr.ObjectPath != nil && *dr.ObjectPath == LeaveEmpty {
		dr.ObjectPath = nil
	} else if dr.ObjectPath == nil {
		dr.ObjectPath = utils.StringPtr(defaultObjectPath)
	}
	if dr.Interface == nil && dr.Sender == nil && dr.Member == nil && dr.ObjectPath == nil {
		return errors.New("dbus rule cant be empty, at least one of interface, sender, member or object_path has to be provided")
	}

	return nil
}

func (d *DbusSignalReceiveFilter) Validate() error {
	if d.Name == nil && d.Body == nil {
		return errors.New("name and body cant both be empty")
	}

	return nil
}
