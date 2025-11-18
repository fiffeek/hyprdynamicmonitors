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
	cfg     *config.RawConfig
	t       *testing.T
	cfgFile *string
}

func NewTestConfig(t *testing.T) *TestConfig {
	return &TestConfig{cfg: &config.RawConfig{}, t: t}
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

func (t *TestConfig) RequireLid(power config.LidStateType) *TestConfig {
	for _, profile := range t.cfg.Profiles {
		profile.Conditions.LidState = utils.JustPtr(power)
	}
	return t
}

func (t *TestConfig) RequirePower(power config.PowerStateType) *TestConfig {
	for _, profile := range t.cfg.Profiles {
		profile.Conditions.PowerState = utils.JustPtr(power)
	}
	return t
}

func (t *TestConfig) FillFallbackProfileConfigFile(path string) *TestConfig {
	profile := t.cfg.FallbackProfile
	require.NotNil(t.t, profile, "cant find fallback profile")
	// nolint:gosec
	configFileContent, err := os.ReadFile(path)
	require.NoError(t.t, err, "cant read config file")
	require.NoError(t.t, utils.WriteAtomic(profile.ConfigFile, configFileContent),
		"cant write to fallback profile config")
	return t
}

func (t *TestConfig) FillProfileConfigFile(name, path string) *TestConfig {
	profile, ok := t.cfg.Profiles[name]
	require.True(t.t, ok, "cant find profile "+name)
	// nolint:gosec
	configFileContent, err := os.ReadFile(path)
	require.NoError(t.t, err, "cant read config file")
	require.NoError(t.t, utils.WriteAtomic(profile.ConfigFile, configFileContent), "cant write to profile config")
	return t
}

func (t *TestConfig) WithNotifications(n *config.Notifications) *TestConfig {
	t.cfg.Notifications = n
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

func (t *TestConfig) WithLidSection(ps *config.LidSection) *TestConfig {
	t.cfg.LidEvents = ps
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

func (t *TestConfig) WithPreExec(fun string) *TestConfig {
	if t.cfg.General == nil {
		t.cfg.General = &config.GeneralSection{}
	}
	t.cfg.General.PreApplyExec = utils.StringPtr(fun)
	return t
}

func (t *TestConfig) WithPostExec(fun string) *TestConfig {
	if t.cfg.General == nil {
		t.cfg.General = &config.GeneralSection{}
	}
	t.cfg.General.PostApplyExec = utils.StringPtr(fun)
	return t
}

func (t *TestConfig) WithConfigPath(path string) *TestConfig {
	return t.WithConfigDir(filepath.Dir(path))
}

func (t *TestConfig) WithConfigDir(dir string) *TestConfig {
	require.NoError(t.t, os.MkdirAll(dir, 0o750))

	cfgFile := filepath.Join(dir, "config.toml")

	// only write empty config when the file does not already exist
	if _, err := os.Stat(cfgFile); err != nil {
		// nolint:gosec
		if _, err := os.Create(cfgFile); err != nil {
			t.t.Fatalf("Failed to create file: %v", err)
		}
	}
	t.cfgFile = &cfgFile

	return t
}

func (t *TestConfig) SaveToFile() *TestConfig {
	buf := new(bytes.Buffer)
	encoder := toml.NewEncoder(buf)
	encoder.Indent = ""
	if err := encoder.Encode(t.cfg); err != nil {
		t.t.Fatal("cant encode config: %w", err)
	}
	require.NotNil(t.t, t.cfgFile, "cfgFile cant be nil")
	if err := utils.WriteAtomic(*t.cfgFile, buf.Bytes()); err != nil {
		t.t.Fatal("cant write config: %w", err)
	}
	// nolint:gosec
	contents, err := os.ReadFile(*t.cfgFile)
	if err != nil {
		t.t.Fatal("cant read self-written config: %w", err)
	}
	t.t.Logf("Saved\n%s\n to %s", string(contents), *t.cfgFile)
	return t
}

func (t *TestConfig) createConfig() *config.Config {
	logrus.WithFields(logrus.Fields{"path": *t.cfgFile}).Debug("Creating config")
	cfg, err := config.NewConfig(*t.cfgFile)
	require.NoError(t.t, err, "cant create config")

	return cfg
}

func (t *TestConfig) WithDestination(dest string) *TestConfig {
	if t.cfg.General == nil {
		t.cfg.General = &config.GeneralSection{}
	}
	t.cfg.General.Destination = utils.StringPtr(dest)
	return t
}

func (t *TestConfig) WithServiceDebounceTime(ms int) *TestConfig {
	if t.cfg.General == nil {
		t.cfg.General = &config.GeneralSection{}
	}
	t.cfg.General.DebounceTimeMs = utils.IntPtr(ms)
	return t
}

func (t *TestConfig) WithFilewatcherDebounceTime(ms int) *TestConfig {
	if t.cfg.HotReload == nil {
		t.cfg.HotReload = &config.HotReloadSection{}
	}
	t.cfg.HotReload.UpdateDebounceTimer = utils.IntPtr(ms)
	return t
}

func (t *TestConfig) WithDestinationContents(filepath string) *TestConfig {
	//nolint:gosec
	contents, err := os.ReadFile(filepath)
	require.NoError(t.t, err, "should be able to read the fixture file")
	require.NoError(t.t, utils.WriteAtomic(*t.cfg.General.Destination, contents),
		"should be able to write to destination")
	return t
}

func (t *TestConfig) FillDefaults() *TestConfig {
	if t.cfg.Profiles == nil {
		t = t.WithProfiles(map[string]*config.Profile{
			"ac": {
				Name: "ac",
				Conditions: &config.ProfileCondition{
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
	if t.cfg.General == nil || t.cfg.General.Destination == nil {
		t = t.WithDestination(filepath.Join(t.t.TempDir(), "target.conf"))
	}
	return t
}

func (t *TestConfig) Get() *config.Config {
	return t.FillDefaults().SaveToFile().createConfig()
}
