package generators_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/fiffeek/hyprdynamicmonitors/internal/config"
	"github.com/fiffeek/hyprdynamicmonitors/internal/generators"
	"github.com/fiffeek/hyprdynamicmonitors/internal/hypr"
	"github.com/fiffeek/hyprdynamicmonitors/internal/matchers"
	"github.com/fiffeek/hyprdynamicmonitors/internal/power"
	"github.com/fiffeek/hyprdynamicmonitors/internal/testutils"
	"github.com/fiffeek/hyprdynamicmonitors/internal/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfigGenerator_GenerateConfig_Static(t *testing.T) {
	cfg := testutils.NewTestConfig(t).Get()
	generator, err := generators.NewConfigGenerator(cfg)
	require.NoError(t, err, "config generators should be able to init")

	tempDir := t.TempDir()
	destination := filepath.Join(tempDir, "hyprland.conf")

	staticConfigPath, err := filepath.Abs("testdata/static_config.conf")
	if err != nil {
		t.Fatalf("Failed to get absolute path: %v", err)
	}

	profile := &config.Profile{
		ConfigFile: staticConfigPath,
		ConfigType: utils.JustPtr(config.Static),
	}

	monitors := []*hypr.MonitorSpec{
		{Name: "DP-1", ID: utils.IntPtr(0), Description: "External Monitor"},
		{Name: "eDP-1", ID: utils.IntPtr(1), Description: "Built-in Display"},
	}

	matchedProfile := matchers.NewMatchedProfile(profile, map[int]*config.RequiredMonitor{})

	changed, err := generator.GenerateConfig(cfg.Get(), matchedProfile, monitors,
		power.ACPowerState, power.OpenedLidState, destination, false)
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

	changed, err = generator.GenerateConfig(cfg.Get(), matchedProfile, monitors,
		power.ACPowerState, power.OpenedLidState, destination, false)
	assert.NoError(t, err, "GenerateConfig failed")
	assert.False(t, changed, "file was changed")

	// if the underlying file is changed report generator as updated
	err = os.Chtimes(destination, time.Now(), time.Now())
	assert.NoError(t, err, "touch failed")
	changed, err = generator.GenerateConfig(cfg.Get(), matchedProfile, monitors,
		power.ACPowerState, power.OpenedLidState, destination, false)
	assert.NoError(t, err, "GenerateConfig failed")
	assert.True(t, changed, "file was not changed")

	// assert dry runs
	require.NoError(t, os.Remove(destination), "should be able to remove the destination file")
	changed, err = generator.GenerateConfig(cfg.Get(), matchedProfile, monitors,
		power.ACPowerState, power.OpenedLidState, destination, true)
	assert.False(t, changed, "nothing should change on dry run")
	assert.NoError(t, err, "no error should be thrown on dry run")
	testutils.AssertFileDoesNotExist(t, destination)
}

func TestConfigGenerator_GenerateConfig_StaticSwap(t *testing.T) {
	cfg := testutils.NewTestConfig(t).Get()
	generator, err := generators.NewConfigGenerator(cfg)
	require.NoError(t, err, "config generators should be able to init")

	tempDir := t.TempDir()
	destination := filepath.Join(tempDir, "hyprland.conf")

	staticConfigPath, err := filepath.Abs("testdata/static_config.conf")
	require.NoError(t, err, "failed to get absolute path")
	otherConfigPath := filepath.Join(tempDir, "other_config.conf")
	require.NoError(t, os.WriteFile(otherConfigPath, []byte("# other\n"), 0o600),
		"should be able to write the other profile config")

	profile := &config.Profile{
		ConfigFile: staticConfigPath,
		ConfigType: utils.JustPtr(config.Static),
	}
	otherProfile := &config.Profile{
		ConfigFile: otherConfigPath,
		ConfigType: utils.JustPtr(config.Static),
	}

	monitors := []*hypr.MonitorSpec{
		{Name: "eDP-1", ID: utils.IntPtr(0), Description: "Built-in Display"},
	}
	matchedProfile := matchers.NewMatchedProfile(profile, map[int]*config.RequiredMonitor{})
	otherMatchedProfile := matchers.NewMatchedProfile(otherProfile, map[int]*config.RequiredMonitor{})

	// switching profiles repoints the link, the destination stays a symlink
	_, err = generator.GenerateConfig(cfg.Get(), matchedProfile, monitors,
		power.ACPowerState, power.OpenedLidState, destination, false)
	require.NoError(t, err, "GenerateConfig failed")
	changed, err := generator.GenerateConfig(cfg.Get(), otherMatchedProfile, monitors,
		power.ACPowerState, power.OpenedLidState, destination, false)
	assert.NoError(t, err, "GenerateConfig failed on profile switch")
	assert.True(t, changed, "profile switch should report a change")
	linkTarget, err := os.Readlink(destination)
	assert.NoError(t, err, "destination should still be a symlink")
	assert.Equal(t, otherConfigPath, linkTarget, "link should point at the new profile")

	// a stale temp link from an interrupted swap does not break linking
	require.NoError(t, os.Symlink(staticConfigPath, destination+".tmp"),
		"should be able to plant a stale temp link")
	changed, err = generator.GenerateConfig(cfg.Get(), matchedProfile, monitors,
		power.ACPowerState, power.OpenedLidState, destination, false)
	assert.NoError(t, err, "GenerateConfig should recover from a stale temp link")
	assert.True(t, changed, "profile switch should report a change")
	testutils.AssertFileDoesNotExist(t, destination+".tmp")

	// a regular file at the destination (e.g. left by a template profile) is replaced
	require.NoError(t, os.Remove(destination), "should be able to remove the destination")
	require.NoError(t, os.WriteFile(destination, []byte("# rendered\n"), 0o600),
		"should be able to write a regular file at the destination")
	changed, err = generator.GenerateConfig(cfg.Get(), otherMatchedProfile, monitors,
		power.ACPowerState, power.OpenedLidState, destination, false)
	assert.NoError(t, err, "GenerateConfig should replace a regular file")
	assert.True(t, changed, "replacing a regular file should report a change")
	linkInfo, err := os.Lstat(destination)
	assert.NoError(t, err, "failed to stat destination")
	assert.NotZero(t, linkInfo.Mode()&os.ModeSymlink, "destination should be a symlink again")

	// dry run must not touch an existing destination
	changed, err = generator.GenerateConfig(cfg.Get(), matchedProfile, monitors,
		power.ACPowerState, power.OpenedLidState, destination, true)
	assert.NoError(t, err, "no error should be thrown on dry run")
	assert.False(t, changed, "nothing should change on dry run")
	linkTarget, err = os.Readlink(destination)
	assert.NoError(t, err, "dry run should leave the existing link in place")
	assert.Equal(t, otherConfigPath, linkTarget, "dry run must not modify the link")
}

func TestConfigGenerator_GenerateConfig_Template(t *testing.T) {
	cfg := testutils.NewTestConfig(t).WithStaticTemplateValues(map[string]string{
		"overwritten_value": "this_shall_be_overwritten",
		"general_value":     "general",
	}).Get()
	generator, err := generators.NewConfigGenerator(cfg)
	require.NoError(t, err, "config generators should be able to init")

	tempDir := t.TempDir()
	destination := filepath.Join(tempDir, "hyprland.conf")

	templateConfigPath, err := filepath.Abs("testdata/template_config.conf.tmpl")
	if err != nil {
		t.Fatalf("Failed to get absolute path: %v", err)
	}

	profile := &config.Profile{
		ConfigFile: templateConfigPath,
		ConfigType: utils.JustPtr(config.Template),
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

	matchedProfile := matchers.NewMatchedProfile(profile, map[int]*config.RequiredMonitor{
		1: {
			Name:       utils.StringPtr("eDP-1"),
			MonitorTag: utils.StringPtr("laptop"),
		},
		0: {
			Description: utils.StringPtr("External Monitor"),
			MonitorTag:  utils.StringPtr("external"),
		},
	})

	monitors := []*hypr.MonitorSpec{
		{
			Name:        "DP-1",
			ID:          utils.IntPtr(0),
			Description: "External Monitor",
			AvailableModes: []string{
				"1920x1080@60.00Hz",
				"1920x1080@59.94Hz",
			},
			Mirror:          "none",
			CurrentFormat:   "XRGB8888",
			DpmsStatus:      true,
			ActivelyTearing: false,
			DirectScanoutTo: "0",
			Solitary:        "0",
		},
		{
			Name:        "eDP-1",
			ID:          utils.IntPtr(1),
			Description: "Built-in Display",
			AvailableModes: []string{
				"2560x1600@60.00Hz",
				"1920x1200@60.00Hz",
			},
			Mirror:          "none",
			CurrentFormat:   "XRGB8888",
			DpmsStatus:      true,
			ActivelyTearing: false,
			DirectScanoutTo: "0",
			Solitary:        "0",
		},
		{
			Name:        "DP-11",
			ID:          utils.IntPtr(2),
			Description: "Extra Monitor",
			AvailableModes: []string{
				"3840x2160@60.00Hz",
				"1920x1080@60.00Hz",
			},
			Mirror:          "none",
			CurrentFormat:   "XRGB8888",
			DpmsStatus:      false,
			ActivelyTearing: true,
			DirectScanoutTo: "HDMI-A-1",
			Solitary:        "1",
		},
	}

	// Test with battery power state
	changed, err := generator.GenerateConfig(cfg.Get(), matchedProfile, monitors,
		power.BatteryPowerState, power.OpenedLidState, destination, false)
	if err != nil {
		t.Fatalf("GenerateConfig failed: %v", err)
	}
	if !changed {
		t.Fatalf("file was not changed")
	}

	testutils.AssertFixture(t, destination, "testdata/fixtures/bat.conf", *regenerate)

	// Test with AC power state
	changed, err = generator.GenerateConfig(cfg.Get(), matchedProfile, monitors,
		power.ACPowerState, power.ClosedLidState, destination, false)
	if err != nil {
		t.Fatalf("GenerateConfig failed with AC power: %v", err)
	}
	if !changed {
		t.Fatalf("file was not changed")
	}

	testutils.AssertFixture(t, destination, "testdata/fixtures/ac.conf", *regenerate)

	changed, err = generator.GenerateConfig(cfg.Get(), matchedProfile, monitors,
		power.ACPowerState, power.ClosedLidState, destination, false)
	if err != nil {
		t.Fatalf("GenerateConfig failed with AC power: %v", err)
	}
	if changed {
		t.Fatalf("file was changed")
	}

	testutils.AssertFixture(t, destination, "testdata/fixtures/ac.conf", *regenerate)

	require.NoError(t, os.Remove(destination), "should be able to remove the destination file")
	changed, err = generator.GenerateConfig(cfg.Get(), matchedProfile, monitors,
		power.ACPowerState, power.ClosedLidState, destination, true)
	assert.False(t, changed, "should not change anything on dry run")
	assert.NoError(t, err, "should not err on dry run")
	testutils.AssertFileDoesNotExist(t, destination)
}
