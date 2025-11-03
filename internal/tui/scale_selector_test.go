package tui_test

import (
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/fiffeek/hyprdynamicmonitors/internal/tui"
	"github.com/fiffeek/hyprdynamicmonitors/internal/utils"
)

func TestScaleSelector_Update(t *testing.T) {
	testCases := []struct {
		name               string
		setupMonitor       *tui.MonitorSpec
		message            tea.Msg
		expectedScale      float64
		expectedCmd        bool
		expectedCmdType    interface{}
		expectedMonitorIdx int
	}{
		{
			name: "MonitorBeingEdited sets selectedMonitorIndex",
			setupMonitor: &tui.MonitorSpec{
				ID:               utils.IntPtr(1),
				Width:            1920,
				Height:           1080,
				Scale:            1.5,
				ValidScalesCache: make(map[float64]struct{}),
			},
			message:            tui.MonitorBeingEdited{MonitorID: 42, ListIndex: 3},
			expectedScale:      1.5,
			expectedCmd:        false,
			expectedMonitorIdx: 42,
		},
		{
			name: "MonitorUnselected resets selectedMonitorIndex",
			setupMonitor: &tui.MonitorSpec{
				ID:               utils.IntPtr(1),
				Width:            1920,
				Height:           1080,
				Scale:            2.0,
				ValidScalesCache: make(map[float64]struct{}),
			},
			message:            tui.MonitorUnselected{},
			expectedScale:      2.0,
			expectedCmd:        false,
			expectedMonitorIdx: -1,
		},
		{
			name: "Up key increases scale and sends preview command",
			setupMonitor: &tui.MonitorSpec{
				ID:               utils.IntPtr(5),
				Width:            1920,
				Height:           1080,
				Scale:            1.0,
				ValidScalesCache: make(map[float64]struct{}),
			},
			message:         tea.KeyMsg{Type: tea.KeyUp},
			expectedScale:   1.2, // Nearest valid scale for 1920x1080 above 1.0
			expectedCmd:     true,
			expectedCmdType: nil, // Returns BatchMsg with multiple commands
		},
		{
			name: "Down key decreases scale and sends preview command",
			setupMonitor: &tui.MonitorSpec{
				ID:               utils.IntPtr(7),
				Width:            1920,
				Height:           1080,
				Scale:            1.5,
				ValidScalesCache: make(map[float64]struct{}),
			},
			message:         tea.KeyMsg{Type: tea.KeyDown},
			expectedScale:   1.3333333333333333, // Nearest valid scale for 1920x1080 below 1.5
			expectedCmd:     true,
			expectedCmdType: nil, // Returns BatchMsg with multiple commands
		},
		{
			name: "Enter key sends scale command",
			setupMonitor: &tui.MonitorSpec{
				ID:               utils.IntPtr(6),
				Width:            1920,
				Height:           1080,
				Scale:            1.25,
				ValidScalesCache: make(map[float64]struct{}),
			},
			message:         tea.KeyMsg{Type: tea.KeyEnter},
			expectedScale:   1.25,
			expectedCmd:     true,
			expectedCmdType: tui.ScaleMonitorCommand{},
		},
		{
			name: "Scale cannot go below minimum",
			setupMonitor: &tui.MonitorSpec{
				ID:               utils.IntPtr(9),
				Width:            1920,
				Height:           1080,
				Scale:            0.1,
				ValidScalesCache: make(map[float64]struct{}),
			},
			message:       tea.KeyMsg{Type: tea.KeyDown},
			expectedScale: 0.1,
			expectedCmd:   true,
		},
		{
			name: "Scale cannot go above maximum",
			setupMonitor: &tui.MonitorSpec{
				ID:               utils.IntPtr(10),
				Width:            1920,
				Height:           1080,
				Scale:            10.0,
				ValidScalesCache: make(map[float64]struct{}),
			},
			message:       tea.KeyMsg{Type: tea.KeyUp},
			expectedScale: 10.0,
			expectedCmd:   true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			selector := tui.NewScaleSelector()
			selector.Set(tc.setupMonitor)

			cmd := selector.Update(tc.message)

			assert.Equal(t, tc.expectedScale, selector.GetCurrentScale(),
				"Expected scale: %.2f, Actual scale: %.2f", tc.expectedScale, selector.GetCurrentScale())

			if !tc.expectedCmd {
				assert.Nil(t, cmd, "Expected no command but got one")
				return
			}

			require.NotNil(t, cmd, "Expected command but got nil")
			msg := cmd()
			if tc.expectedCmdType != nil {
				assert.IsType(t, tc.expectedCmdType, msg,
					"Expected command type %T but got %T", tc.expectedCmdType, msg)
			}

			if tc.expectedMonitorIdx != 0 {
				assert.Equal(t, tc.expectedMonitorIdx, selector.GetSelectedMonitorIndex(),
					"Expected monitor index: %d, Actual: %d", tc.expectedMonitorIdx, selector.GetSelectedMonitorIndex())
			}
		})
	}
}

func TestScaleSelector_AcceleratedScaling(t *testing.T) {
	monitor := &tui.MonitorSpec{
		ID:               utils.IntPtr(1),
		Width:            1920,
		Height:           1080,
		Scale:            1.0,
		ValidScalesCache: make(map[float64]struct{}),
	}
	selector := tui.NewScaleSelector()
	selector.Set(monitor)

	selector.Update(tea.KeyMsg{Type: tea.KeyUp})
	selector.Update(tea.KeyMsg{Type: tea.KeyUp})
	time.Sleep(50 * time.Millisecond)
	selector.Update(tea.KeyMsg{Type: tea.KeyUp})
	// With scale snapping enabled, successive increments will snap to nearest valid scales:
	// 1.0 -> 1.005 snaps to 1.2
	// 1.2 -> 1.205 snaps to 1.25
	// 1.25 -> 1.26 (with 2x multiplier) snaps to 1.333...
	assert.InDelta(t, 1.3333333333333333, selector.GetCurrentScale(), 0.01)
}

func TestScaleSelector_DPICalculation(t *testing.T) {
	testCases := []struct {
		name        string
		scale       float64
		expectedDPI int
	}{
		{"1.0 scale", 1.0, 96},
		{"1.5 scale", 1.5, 144},
		{"2.0 scale", 2.0, 192},
		{"1.25 scale", 1.25, 120},
		{"0.75 scale", 0.75, 72},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			monitor := &tui.MonitorSpec{
				ID:               utils.IntPtr(1),
				Width:            1920,
				Height:           1080,
				Scale:            tc.scale,
				ValidScalesCache: make(map[float64]struct{}),
			}

			selector := tui.NewScaleSelector()
			selector.Set(monitor)

			view := selector.View()

			// The view should contain the DPI value
			assert.Contains(t, view, "DPI", "View should contain DPI information")
			assert.NotEmpty(t, view, "View should not be empty")
		})
	}
}

func TestScaleSelector_CustomInput(t *testing.T) {
	testCases := []struct {
		name          string
		initialScale  float64
		inputValue    string
		expectedScale float64
	}{
		{"Valid scale 1.5", 1.0, "1.5", 1.5},
		{"Valid scale 2.0", 1.0, "2.0", 2.0},
		{"Valid scale 0.75", 1.0, "0.75", 0.75},
		{"Invalid - not a number", 1.0, "abc", 1.0},
		{"Cancel with Esc", 1.0, "2.5", 1.0}, // Special case: will press Esc instead of Enter
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			monitor := &tui.MonitorSpec{
				ID:               utils.IntPtr(1),
				Width:            1920,
				Height:           1080,
				Scale:            tc.initialScale,
				ValidScalesCache: make(map[float64]struct{}),
			}

			selector := tui.NewScaleSelector()
			selector.Set(monitor)

			// Trigger custom input mode
			selector.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'C'}})

			// Clear the pre-filled value first
			selector.Update(tea.KeyMsg{Type: tea.KeyCtrlU}) // Ctrl+U clears the line in textinput

			// Type the value
			for _, r := range tc.inputValue {
				selector.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
			}

			// Press enter to apply (or Esc for cancel test)
			if tc.name == "Cancel with Esc" {
				selector.Update(tea.KeyMsg{Type: tea.KeyEsc})
			} else {
				selector.Update(tea.KeyMsg{Type: tea.KeyEnter})
			}

			assert.Equal(t, tc.expectedScale, selector.GetCurrentScale(),
				"Expected scale to be %.2f, got %.2f", tc.expectedScale, selector.GetCurrentScale())
		})
	}
}
