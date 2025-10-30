package tui_test

import (
	"path/filepath"
	"testing"

	"github.com/fiffeek/hyprdynamicmonitors/internal/hypr"
	"github.com/fiffeek/hyprdynamicmonitors/internal/testutils"
	"github.com/fiffeek/hyprdynamicmonitors/internal/tui"
	"github.com/fiffeek/hyprdynamicmonitors/internal/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test__Types_Back_Forth(t *testing.T) {
	tests := []struct {
		name     string
		fixture  string
		after    func(*tui.MonitorSpec)
		validate func(*hypr.MonitorSpec)
	}{
		{
			name:    "flip to no flip",
			fixture: "flipped.json",
			after: func(mon *tui.MonitorSpec) {
				mon.ToggleFlip()
			},
			validate: func(mon *hypr.MonitorSpec) {
				assert.Equal(t, 0, mon.Transform, "flip is disabled")
			},
		},
		{
			name:    "rotate flip",
			fixture: "flipped.json",
			after: func(mon *tui.MonitorSpec) {
				mon.Rotate()
			},
			validate: func(mon *hypr.MonitorSpec) {
				assert.Equal(t, 5, mon.Transform, "rotate flip")
			},
		},
		{
			name:    "rotate flip overflow",
			fixture: "flipped.json",
			after: func(mon *tui.MonitorSpec) {
				mon.Rotate()
				mon.Rotate()
				mon.Rotate()
				mon.Rotate()
			},
			validate: func(mon *hypr.MonitorSpec) {
				assert.Equal(t, 4, mon.Transform, "rotate flip")
			},
		},
		{
			name:    "flip rotated",
			fixture: "rotated.json",
			after: func(mon *tui.MonitorSpec) {
				mon.ToggleFlip()
			},
			validate: func(mon *hypr.MonitorSpec) {
				assert.Equal(t, 7, mon.Transform, "flip rotate")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join("testdata", tt.fixture)
			monitors, err := tui.LoadMonitorsFromJSON(path)
			require.NoError(t, err, "should read monitors data")
			require.Len(t, monitors, 1, "only one monitor can be passed for this test")

			tt.after(monitors[0])

			hyprmon, err := tui.ConvertToHyprMonitors(monitors)
			assert.NoError(t, err, "should condert to hypr dat")

			tt.validate(hyprmon[0])
		})
	}
}

func Test__Types_ToHyprCommand(t *testing.T) {
	tests := []struct {
		name    string
		fixture string
		after   func(*tui.MonitorSpec)
		golden  string
	}{
		{
			name:    "disable",
			fixture: "flipped.json",
			after: func(m *tui.MonitorSpec) {
				m.ToggleMonitor()
			},
			golden: "golden/disable",
		},
		{
			name:    "unflip",
			fixture: "flipped.json",
			after: func(m *tui.MonitorSpec) {
				m.ToggleFlip()
			},
			golden: "golden/unflip",
		},
		{
			name:    "flip",
			fixture: "rotated.json",
			after: func(m *tui.MonitorSpec) {
				m.ToggleFlip()
			},
			golden: "golden/flip",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join("testdata", tt.fixture)
			monitors, err := tui.LoadMonitorsFromJSON(path)
			require.NoError(t, err, "should read monitors data")
			require.Len(t, monitors, 1, "only one monitor can be passed for this test")

			tt.after(monitors[0])

			hyprCommand := monitors[0].ToHypr()
			dir := t.TempDir()
			file := filepath.Join(dir, tt.name)
			//nolint:gosec
			utils.WriteAtomic(file, []byte(hyprCommand))

			testutils.AssertFixture(t, file, filepath.Join("testdata", tt.golden), *regenerate)
		})
	}
}
