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
				ID:    utils.IntPtr(1),
				Scale: 1.5,
			},
			message:            tui.MonitorBeingEdited{MonitorID: 42, ListIndex: 3},
			expectedScale:      1.5,
			expectedCmd:        false,
			expectedMonitorIdx: 42,
		},
		{
			name: "MonitorUnselected resets selectedMonitorIndex",
			setupMonitor: &tui.MonitorSpec{
				ID:    utils.IntPtr(1),
				Scale: 2.0,
			},
			message:            tui.MonitorUnselected{},
			expectedScale:      2.0,
			expectedCmd:        false,
			expectedMonitorIdx: -1,
		},
		{
			name: "Up key increases scale and sends preview command",
			setupMonitor: &tui.MonitorSpec{
				ID:    utils.IntPtr(5),
				Scale: 1.0,
			},
			message:         tea.KeyMsg{Type: tea.KeyUp},
			expectedScale:   1.01,
			expectedCmd:     true,
			expectedCmdType: tui.PreviewScaleMonitorCommand{},
		},
		{
			name: "Down key decreases scale and sends preview command",
			setupMonitor: &tui.MonitorSpec{
				ID:    utils.IntPtr(7),
				Scale: 1.5,
			},
			message:         tea.KeyMsg{Type: tea.KeyDown},
			expectedScale:   1.49,
			expectedCmd:     true,
			expectedCmdType: tui.PreviewScaleMonitorCommand{},
		},
		{
			name: "Enter key sends scale command",
			setupMonitor: &tui.MonitorSpec{
				ID:    utils.IntPtr(6),
				Scale: 1.25,
			},
			message:         tea.KeyMsg{Type: tea.KeyEnter},
			expectedScale:   1.25,
			expectedCmd:     true,
			expectedCmdType: tui.ScaleMonitorCommand{},
		},
		{
			name: "Scale cannot go below minimum",
			setupMonitor: &tui.MonitorSpec{
				ID:    utils.IntPtr(9),
				Scale: 0.1,
			},
			message:       tea.KeyMsg{Type: tea.KeyDown},
			expectedScale: 0.1,
			expectedCmd:   true,
		},
		{
			name: "Scale cannot go above maximum",
			setupMonitor: &tui.MonitorSpec{
				ID:    utils.IntPtr(10),
				Scale: 10.0,
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
	monitor := &tui.MonitorSpec{ID: utils.IntPtr(1), Scale: 1.0}
	selector := tui.NewScaleSelector()
	selector.Set(monitor)

	selector.Update(tea.KeyMsg{Type: tea.KeyUp})
	selector.Update(tea.KeyMsg{Type: tea.KeyUp})
	time.Sleep(50 * time.Millisecond)
	selector.Update(tea.KeyMsg{Type: tea.KeyUp})
	assert.InDelta(t, 1.04, selector.GetCurrentScale(), 0.001)
}
