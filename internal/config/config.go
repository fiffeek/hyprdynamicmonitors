package config

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
)

type Config struct {
	configPath string
	Profiles   map[string]*Profile `toml:"profiles"`
	General    *GeneralSection     `toml:"general"`
	Scoring    *ScoringSection     `toml:"scoring"`
}

type GeneralSection struct {
	Destination *string `toml:"destination"`
}

type ScoringSection struct {
	NameMatch        *int `toml:"name_match"`
	DescriptionMatch *int `toml:"description_match"`
}

type Profile struct {
	Name       string
	ConfigFile string           `toml:"config_file"`
	Conditions ProfileCondition `toml:"conditions"`
}

type ProfileCondition struct {
	RequiredMonitors []*RequiredMonitor `toml:"required_monitors"`
	PowerState       string             `toml:"power_state"` // "AC", "BAT", or "any"
}

type RequiredMonitor struct {
	Name        *string `toml:"name"`
	Description *string `toml:"description"`
}

func Load(configPath string) (*Config, error) {
	configPath = os.ExpandEnv(configPath)
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("configuration file %s not found", configPath)
	}

	absConfig, err := filepath.Abs(filepath.Dir(configPath))
	if err != nil {
		return nil, fmt.Errorf("cant convert config bath to abs %v", err)
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
		return fmt.Errorf("no profiles defined")
	}

	if c.General == nil {
		c.General = &GeneralSection{}
	}

	if c.General.Destination == nil {
		defaultDest := "$HOME/.config/hypr/monitors.conf"
		c.General.Destination = &defaultDest
	}

	defaultScore := 1

	if c.Scoring == nil {
		c.Scoring = &ScoringSection{}
	}
	if c.Scoring.NameMatch == nil {
		c.Scoring.NameMatch = &defaultScore
	}
	if c.Scoring.DescriptionMatch == nil {
		c.Scoring.DescriptionMatch = &defaultScore
	}

	dest := os.ExpandEnv(*c.General.Destination)
	c.General.Destination = &dest

	for name, profile := range c.Profiles {
		profile.Name = name

		if profile.ConfigFile == "" {
			return fmt.Errorf("profile %s: config_file is required", name)
		}

		if !strings.HasPrefix(profile.ConfigFile, "/") {
			profile.ConfigFile = filepath.Join(c.configPath, profile.ConfigFile)
		}
		log.Printf("Hello %s %s", name, profile.ConfigFile)
		profile.ConfigFile = os.ExpandEnv(profile.ConfigFile)

		if _, err := os.Stat(profile.ConfigFile); os.IsNotExist(err) {
			return fmt.Errorf("profile %s: config file %s not found", name, profile.ConfigFile)
		}

		switch profile.Conditions.PowerState {
		case "AC", "BAT", "ANY", "":
			if profile.Conditions.PowerState == "" {
				profile.Conditions.PowerState = "ANY"
			}
		default:
			return fmt.Errorf("profile %s: invalid power_state '%s' (must be AC, BAT, or any)",
				name, profile.Conditions.PowerState)
		}

		if len(profile.Conditions.RequiredMonitors) == 0 {
			return fmt.Errorf("profile %s: at least one required_monitor must be specified", name)
		}
	}

	return nil
}
