// Package testutils provides utils for testing
// should not be imported by any other app packages
package testutils

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/BurntSushi/toml"
	"github.com/fiffeek/hyprdynamicmonitors/internal/config"
	"github.com/fiffeek/hyprdynamicmonitors/internal/utils"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

type TestConfig struct {
	cfg     *config.UnsafeConfig
	t       *testing.T
	cfgFile *string
}

func NewTestConfig(t *testing.T) *TestConfig {
	return &TestConfig{cfg: &config.UnsafeConfig{}, t: t}
}

func (t *TestConfig) WithProfiles(profiles map[string]*config.Profile) *TestConfig {
	t.cfg.Profiles = profiles
	for _, profile := range t.cfg.Profiles {
		if profile.ConfigFile == "" {
			tempDir := t.t.TempDir()
			cfgFile := filepath.Join(tempDir, "file")
			profile.ConfigFile = cfgFile
		}
		if _, err := os.Create(profile.ConfigFile); err != nil {
			t.t.Fatalf("Failed to create file: %v", err)
		}
	}
	return t
}

func (t *TestConfig) WithFallbackProfile(fallback *config.Profile) *TestConfig {
	t.cfg.FallbackProfile = fallback
	if fallback != nil && fallback.ConfigFile == "" {
		tempDir := t.t.TempDir()
		cfgFile := filepath.Join(tempDir, "fallback")
		fallback.ConfigFile = cfgFile
	}
	if fallback != nil {
		if _, err := os.Create(fallback.ConfigFile); err != nil {
			t.t.Fatalf("Failed to create fallback config file: %v", err)
		}
	}
	return t
}

func (t *TestConfig) WithScoring(scoring *config.ScoringSection) *TestConfig {
	t.cfg.Scoring = scoring
	return t
}

func (t *TestConfig) WithPowerSection(ps *config.PowerSection) *TestConfig {
	t.cfg.PowerEvents = ps
	return t
}

func (t *TestConfig) WithHotReload(h *config.HotReloadSection) *TestConfig {
	t.cfg.HotReload = h
	return t
}

func (t *TestConfig) WithStaticTemplateValues(s map[string]string) *TestConfig {
	t.cfg.StaticTemplateValues = s
	return t
}

func (t *TestConfig) WithConfigDir(dir string) *TestConfig {
	require.NoError(t.t, os.MkdirAll(dir, 0o750))

	cfgFile := filepath.Join(dir, "config.toml")
	// nolint:gosec
	if _, err := os.Create(cfgFile); err != nil {
		t.t.Fatalf("Failed to create file: %v", err)
	}
	t.cfgFile = &cfgFile

	return t
}

func (t *TestConfig) SaveToFile() *TestConfig {
	buf := new(bytes.Buffer)
	if err := toml.NewEncoder(buf).Encode(t.cfg); err != nil {
		t.t.Fatal("cant encode config: %w", err)
	}
	require.NotNil(t.t, t.cfgFile, "cfgFile cant be nil")
	if err := utils.WriteAtomic(*t.cfgFile, buf.Bytes()); err != nil {
		t.t.Fatal("cant write config: %w", err)
	}
	return t
}

func (t *TestConfig) createConfig() *config.Config {
	logrus.WithFields(logrus.Fields{"path": *t.cfgFile}).Debug("Creating config")
	cfg, err := config.NewConfig(*t.cfgFile)
	require.NoError(t.t, err, "cant create config")

	return cfg
}

func (t *TestConfig) FillDefaults() *TestConfig {
	if t.cfg.Profiles == nil {
		t = t.WithProfiles(map[string]*config.Profile{
			"ac": {
				Name: "ac",
				Conditions: config.ProfileCondition{
					PowerState: utils.JustPtr(config.AC),
					RequiredMonitors: []*config.RequiredMonitor{
						{Name: utils.StringPtr("eDP-1")},
					},
				},
			},
		})
	}
	if t.cfgFile == nil {
		t = t.WithConfigDir(t.t.TempDir())
	}
	return t
}

func (t *TestConfig) Get() *config.Config {
	return t.FillDefaults().SaveToFile().createConfig()
}
