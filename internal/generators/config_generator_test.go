package generators_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/fiffeek/hyprdynamicmonitors/internal/config"
	"github.com/fiffeek/hyprdynamicmonitors/internal/detectors"
	"github.com/fiffeek/hyprdynamicmonitors/internal/generators"
	"github.com/fiffeek/hyprdynamicmonitors/internal/hypr"
	"github.com/fiffeek/hyprdynamicmonitors/internal/utils"
)

func TestConfigGenerator_GenerateConfig_Static(t *testing.T) {
	generator := generators.NewConfigGenerator()

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
		{Name: "DP-1", ID: "0", Description: "External Monitor"},
		{Name: "eDP-1", ID: "1", Description: "Built-in Display"},
	}

	err = generator.GenerateConfig(profile, monitors, detectors.ACPower, destination)
	if err != nil {
		t.Fatalf("GenerateConfig failed: %v", err)
	}

	linkInfo, err := os.Lstat(destination)
	if err != nil {
		t.Fatalf("Failed to stat destination: %v", err)
	}

	if linkInfo.Mode()&os.ModeSymlink == 0 {
		t.Error("Expected destination to be a symlink")
	}

	// Check that the symlink points to the correct file
	linkTarget, err := os.Readlink(destination)
	if err != nil {
		t.Fatalf("Failed to read symlink: %v", err)
	}

	if linkTarget != staticConfigPath {
		t.Errorf("Expected symlink target %s, got %s", staticConfigPath, linkTarget)
	}
}

func TestConfigGenerator_GenerateConfig_Template(t *testing.T) {
	generator := generators.NewConfigGenerator()

	tempDir := t.TempDir()
	destination := filepath.Join(tempDir, "hyprland.conf")

	templateConfigPath, err := filepath.Abs("testdata/template_config.conf.tmpl")
	if err != nil {
		t.Fatalf("Failed to get absolute path: %v", err)
	}

	profile := &config.Profile{
		ConfigFile: templateConfigPath,
		ConfigType: configTypePtr(config.Template),
		Conditions: config.ProfileCondition{
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
	}

	monitors := []*hypr.MonitorSpec{
		{Name: "DP-1", ID: "0", Description: "External Monitor"},
		{Name: "eDP-1", ID: "1", Description: "Built-in Display"},
	}

	// Test with battery power state
	err = generator.GenerateConfig(profile, monitors, detectors.Battery, destination)
	if err != nil {
		t.Fatalf("GenerateConfig failed: %v", err)
	}

	//nolint:gosec
	content, err := os.ReadFile(destination)
	if err != nil {
		t.Fatalf("Failed to read generated config: %v", err)
	}

	contentStr := string(content)

	// Verify template variables were substituted
	if !strings.Contains(contentStr, "Generated with power state: BAT") {
		t.Error("Expected power state BAT to be rendered")
	}

	// Verify all monitors were included
	if !strings.Contains(contentStr, "monitor=DP-1,auto,auto,1.0") {
		t.Error("Expected DP-1 monitor configuration")
	}
	if !strings.Contains(contentStr, "monitor=eDP-1,auto,auto,1.0") {
		t.Error("Expected eDP-1 monitor configuration")
	}

	// Verify MonitorsByTag worked
	if !strings.Contains(contentStr, "workspace=1,monitor:eDP-1") {
		t.Error("Expected laptop monitor tag to resolve to eDP-1")
	}
	if !strings.Contains(contentStr, "workspace=2,monitor:DP-1") {
		t.Error("Expected external monitor tag to resolve to DP-1")
	}

	// Verify power state functions worked
	if !strings.Contains(contentStr, "decoration:blur:enabled = false") {
		t.Error("Expected battery power optimizations")
	}
	if !strings.Contains(contentStr, "animation:enabled = false") {
		t.Error("Expected battery animation disabled")
	}
	if !strings.Contains(contentStr, "Power state function test: BAT") {
		t.Error("Expected powerState function to return BAT")
	}

	// Test with AC power state
	err = generator.GenerateConfig(profile, monitors, detectors.ACPower, destination)
	if err != nil {
		t.Fatalf("GenerateConfig failed with AC power: %v", err)
	}

	//nolint:gosec
	acContent, err := os.ReadFile(destination)
	if err != nil {
		t.Fatalf("Failed to read AC generated config: %v", err)
	}

	acContentStr := string(acContent)

	// Verify AC power state
	if !strings.Contains(acContentStr, "Generated with power state: AC") {
		t.Error("Expected power state AC to be rendered")
	}

	// Verify AC power functions worked
	if !strings.Contains(acContentStr, "decoration:blur:enabled = true") {
		t.Error("Expected AC power blur enabled")
	}
	if !strings.Contains(acContentStr, "animation:enabled = true") {
		t.Error("Expected AC power animation enabled")
	}
	if !strings.Contains(acContentStr, "Power state function test: AC") {
		t.Error("Expected powerState function to return AC")
	}
}

func configTypePtr(c config.ConfigFileType) *config.ConfigFileType {
	return &c
}
