package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/fiffeek/hyprdynamicmonitors/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInclude(t *testing.T) {
	tmpDir := t.TempDir()

	mainContent := `
[general]
destination = "/tmp/dest"

include = "extra_profiles.toml"

[profiles.base]
config_file = "base.conf"
conditions.required_monitors = [{name = "DP-3"}]
`
	extraProfilesContent := `
[profiles.extra]
config_file = "extra.conf"
conditions.required_monitors = [{name = "DP-1"}]

include = "nested.toml"
`

	nestedContent := `
[profiles.nested]
config_file = "nested.conf"
conditions.required_monitors = [{name = "DP-2"}]
`

	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "main.toml"), []byte(mainContent), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "extra_profiles.toml"), []byte(extraProfilesContent), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "nested.toml"), []byte(nestedContent), 0644))

	// Create dummy config files
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "base.conf"), []byte(""), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "extra.conf"), []byte(""), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "nested.conf"), []byte(""), 0644))

	cfg, err := config.Load(filepath.Join(tmpDir, "main.toml"))
	require.NoError(t, err)

	assert.Contains(t, cfg.Profiles, "base")
	assert.Contains(t, cfg.Profiles, "extra")
	assert.Contains(t, cfg.Profiles, "nested")

	assert.NotNil(t, cfg.Profiles["base"].Conditions)
	assert.NotEmpty(t, cfg.Profiles["base"].Conditions.RequiredMonitors)
	assert.Equal(t, "DP-3", *cfg.Profiles["base"].Conditions.RequiredMonitors[0].Name)

	assert.NotNil(t, cfg.Profiles["extra"].Conditions)
	assert.NotEmpty(t, cfg.Profiles["extra"].Conditions.RequiredMonitors)
	assert.Equal(t, "DP-1", *cfg.Profiles["extra"].Conditions.RequiredMonitors[0].Name)

	assert.NotNil(t, cfg.Profiles["nested"].Conditions)
	assert.NotEmpty(t, cfg.Profiles["nested"].Conditions.RequiredMonitors)
	assert.Equal(t, "DP-2", *cfg.Profiles["nested"].Conditions.RequiredMonitors[0].Name)
}

func TestInclude_RelativePaths(t *testing.T) {
	tmpDir := t.TempDir()
	subDir := filepath.Join(tmpDir, "subdir")
	require.NoError(t, os.Mkdir(subDir, 0755))

	mainContent := `
[general]
destination = "/tmp/dest"
include = "subdir/sub.toml"
`
	subContent := `
[profiles.sub]
config_file = "sub.conf" 
conditions.required_monitors = [{name = "DP-4"}]

include = "nested.toml"
`
	nestedContent := `
[profiles.nested]
config_file = "nested.conf"
conditions.required_monitors = [{name = "DP-5"}]
`

	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "main.toml"), []byte(mainContent), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(subDir, "sub.toml"), []byte(subContent), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(subDir, "nested.toml"), []byte(nestedContent), 0644))

	// Dummy configs. 
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "sub.conf"), []byte(""), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "nested.conf"), []byte(""), 0644))

	cfg, err := config.Load(filepath.Join(tmpDir, "main.toml"))
	require.NoError(t, err)

	assert.Contains(t, cfg.Profiles, "sub")
	assert.Contains(t, cfg.Profiles, "nested")

	assert.NotNil(t, cfg.Profiles["sub"].Conditions)
	assert.NotEmpty(t, cfg.Profiles["sub"].Conditions.RequiredMonitors)
	assert.Equal(t, "DP-4", *cfg.Profiles["sub"].Conditions.RequiredMonitors[0].Name)

	assert.NotNil(t, cfg.Profiles["nested"].Conditions)
	assert.NotEmpty(t, cfg.Profiles["nested"].Conditions.RequiredMonitors)
	assert.Equal(t, "DP-5", *cfg.Profiles["nested"].Conditions.RequiredMonitors[0].Name)
}

func TestInclude_List(t *testing.T) {
	tmpDir := t.TempDir()

	mainContent := `
[general]
destination = "/tmp/dest"

include = ["part1.toml", "part2.toml"]
`
	part1 := `
[profiles.p1]
config_file = "p1.conf"
conditions.required_monitors = [{name = "DP-6"}]
`
	part2 := `
[profiles.p2]
config_file = "p2.conf"
conditions.required_monitors = [{name = "DP-7"}]
`
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "main.toml"), []byte(mainContent), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "part1.toml"), []byte(part1), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "part2.toml"), []byte(part2), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "p1.conf"), []byte(""), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "p2.conf"), []byte(""), 0644))

	cfg, err := config.Load(filepath.Join(tmpDir, "main.toml"))
	require.NoError(t, err)

	assert.Contains(t, cfg.Profiles, "p1")
	assert.Contains(t, cfg.Profiles, "p2")

	assert.NotNil(t, cfg.Profiles["p1"].Conditions)
	assert.NotEmpty(t, cfg.Profiles["p1"].Conditions.RequiredMonitors)
	assert.Equal(t, "DP-6", *cfg.Profiles["p1"].Conditions.RequiredMonitors[0].Name)

	assert.NotNil(t, cfg.Profiles["p2"].Conditions)
	assert.NotEmpty(t, cfg.Profiles["p2"].Conditions.RequiredMonitors)
	assert.Equal(t, "DP-7", *cfg.Profiles["p2"].Conditions.RequiredMonitors[0].Name)
}
