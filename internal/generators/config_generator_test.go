package generators_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/fiffeek/hyprdynamicmonitors/internal/config"
	"github.com/fiffeek/hyprdynamicmonitors/internal/generators"
	"github.com/fiffeek/hyprdynamicmonitors/internal/hypr"
	"github.com/fiffeek/hyprdynamicmonitors/internal/power"
	"github.com/fiffeek/hyprdynamicmonitors/internal/testutils"
	"github.com/fiffeek/hyprdynamicmonitors/internal/utils"
	"github.com/stretchr/testify/assert"
)

func TestConfigGenerator_GenerateConfig_Static(t *testing.T) {
	cfg := testutils.NewTestConfig(t).Get()
	generator := generators.NewConfigGenerator(cfg)

	tempDir := t.TempDir()
	destination := filepath.Join(tempDir, "hyprland.conf")

	staticConfigPath, err := filepath.Abs("testdata/static_config.conf")
	if err != nil {
		t.Fatalf("Failed to get absolute path: %v", err)
	}

	profile := &config.Profile{
		ConfigFile: staticConfigPath,
		ConfigType: configTypePtr(config.Static),
	}

	monitors := []*hypr.MonitorSpec{
		{Name: "DP-1", ID: utils.IntPtr(0), Description: "External Monitor"},
		{Name: "eDP-1", ID: utils.IntPtr(1), Description: "Built-in Display"},
	}

	changed, err := generator.GenerateConfig(cfg.Get(), profile, monitors, power.ACPowerState, power.OpenedLidState, destination)
	assert.NoError(t, err, "GenerateConfig failed")
	assert.True(t, changed, "file was not changed")

	linkInfo, err := os.Lstat(destination)
	assert.NoError(t, err, "Failed to stat destination")

	if linkInfo.Mode()&os.ModeSymlink == 0 {
		t.Error("Expected destination to be a symlink")
	}

	// Check that the symlink points to the correct file
	linkTarget, err := os.Readlink(destination)
	assert.NoError(t, err, "Failed to read symlink")

	if linkTarget != staticConfigPath {
		t.Errorf("Expected symlink target %s, got %s", staticConfigPath, linkTarget)
	}

	changed, err = generator.GenerateConfig(cfg.Get(), profile, monitors, power.ACPowerState, power.OpenedLidState, destination)
	assert.NoError(t, err, "GenerateConfig failed")
	assert.False(t, changed, "file was changed")

	// if the underlying file is changed report generator as updated
	err = os.Chtimes(destination, time.Now(), time.Now())
	assert.NoError(t, err, "touch failed")
	changed, err = generator.GenerateConfig(cfg.Get(), profile, monitors, power.ACPowerState, power.OpenedLidState, destination)
	assert.NoError(t, err, "GenerateConfig failed")
	assert.True(t, changed, "file was not changed")
}

func TestConfigGenerator_GenerateConfig_Template(t *testing.T) {
	cfg := testutils.NewTestConfig(t).WithStaticTemplateValues(map[string]string{
		"overwritten_value": "this_shall_be_overwritten",
		"general_value":     "general",
	}).Get()
	generator := generators.NewConfigGenerator(cfg)

	tempDir := t.TempDir()
	destination := filepath.Join(tempDir, "hyprland.conf")

	templateConfigPath, err := filepath.Abs("testdata/template_config.conf.tmpl")
	if err != nil {
		t.Fatalf("Failed to get absolute path: %v", err)
	}

	profile := &config.Profile{
		ConfigFile: templateConfigPath,
		ConfigType: configTypePtr(config.Template),
		Conditions: &config.ProfileCondition{
			RequiredMonitors: []*config.RequiredMonitor{
				{
					Name:       utils.StringPtr("eDP-1"),
					MonitorTag: utils.StringPtr("laptop"),
				},
				{
					Description: utils.StringPtr("External Monitor"),
					MonitorTag:  utils.StringPtr("external"),
				},
			},
		},
		StaticTemplateValues: map[string]string{
			"profile_value":     "profile",
			"overwritten_value": "overwritten_profile",
		},
	}

	monitors := []*hypr.MonitorSpec{
		{Name: "DP-1", ID: utils.IntPtr(0), Description: "External Monitor"},
		{Name: "eDP-1", ID: utils.IntPtr(1), Description: "Built-in Display"},
		{Name: "DP-11", ID: utils.IntPtr(2), Description: "Extra Monitor"},
	}

	// Test with battery power state
	changed, err := generator.GenerateConfig(cfg.Get(), profile, monitors,
		power.BatteryPowerState, power.OpenedLidState, destination)
	if err != nil {
		t.Fatalf("GenerateConfig failed: %v", err)
	}
	if !changed {
		t.Fatalf("file was not changed")
	}

	testutils.AssertFixture(t, destination, "testdata/fixtures/bat.conf", *regenerate)

	// Test with AC power state
	changed, err = generator.GenerateConfig(cfg.Get(), profile, monitors, power.ACPowerState, power.ClosedLidState, destination)
	if err != nil {
		t.Fatalf("GenerateConfig failed with AC power: %v", err)
	}
	if !changed {
		t.Fatalf("file was not changed")
	}

	testutils.AssertFixture(t, destination, "testdata/fixtures/ac.conf", *regenerate)

	changed, err = generator.GenerateConfig(cfg.Get(), profile, monitors, power.ACPowerState, power.ClosedLidState, destination)
	if err != nil {
		t.Fatalf("GenerateConfig failed with AC power: %v", err)
	}
	if changed {
		t.Fatalf("file was changed")
	}

	testutils.AssertFixture(t, destination, "testdata/fixtures/ac.conf", *regenerate)
}

func configTypePtr(c config.ConfigFileType) *config.ConfigFileType {
	return &c
}
