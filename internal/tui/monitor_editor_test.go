package tui_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/fiffeek/hyprdynamicmonitors/internal/tui"
	"github.com/fiffeek/hyprdynamicmonitors/internal/utils"
)

func TestSetMirror_LoopDetection(t *testing.T) {
	testCases := []struct {
		name       string
		setupFile  string
		setupFunc  func([]*tui.MonitorSpec) []*tui.MonitorSpec // Optional setup function
		monitorID  int
		mirrorOf   string
		shouldFail bool
		reason     string
	}{
		// Valid operations (no loops)
		{
			name:       "Mirror A to B (no existing mirrors)",
			setupFile:  "testdata/mirror_no_loop.json",
			monitorID:  0,
			mirrorOf:   "MonitorB",
			shouldFail: false,
			reason:     "should allow simple mirroring",
		},
		{
			name:       "Mirror B to C (no existing mirrors)",
			setupFile:  "testdata/mirror_no_loop.json",
			monitorID:  1,
			mirrorOf:   "MonitorC",
			shouldFail: false,
			reason:     "should allow simple mirroring",
		},
		{
			name:       "Set mirror to none",
			setupFile:  "testdata/mirror_no_loop.json",
			monitorID:  0,
			mirrorOf:   "none",
			shouldFail: false,
			reason:     "should allow removing mirror",
		},
		{
			name:      "Extend chain (D mirrors C)",
			setupFile: "testdata/mirror_chain.json",
			setupFunc: func(monitors []*tui.MonitorSpec) []*tui.MonitorSpec {
				// Add MonitorD
				monitorD := &tui.MonitorSpec{
					Name:        "MonitorD",
					ID:          utils.JustPtr(3),
					Description: "Monitor D",
					Disabled:    false,
					Mirror:      "none",
				}
				return append(monitors, monitorD)
			},
			monitorID:  3,
			mirrorOf:   "MonitorC",
			shouldFail: false,
			reason:     "should allow extending chain",
		},
		// Invalid operations (would create loops)
		{
			name:       "Direct loop (B mirrors A when A already mirrors B)",
			setupFile:  "testdata/mirror_direct_loop.json",
			monitorID:  1,
			mirrorOf:   "MonitorA",
			shouldFail: true,
			reason:     "should detect direct loop",
		},
		{
			name:       "Indirect loop (C mirrors A in chain A->B->C)",
			setupFile:  "testdata/mirror_chain.json",
			monitorID:  2,
			mirrorOf:   "MonitorA",
			shouldFail: true,
			reason:     "should detect indirect loop",
		},
		{
			name:       "Indirect loop (C mirrors B in chain A->B->C)",
			setupFile:  "testdata/mirror_chain.json",
			monitorID:  2,
			mirrorOf:   "MonitorB",
			shouldFail: true,
			reason:     "should detect indirect loop",
		},
		{
			name:       "Self mirror",
			setupFile:  "testdata/mirror_no_loop.json",
			monitorID:  0,
			mirrorOf:   "MonitorA",
			shouldFail: true,
			reason:     "should prevent self-mirroring",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			monitors, err := tui.LoadMonitorsFromJSON(tc.setupFile)
			require.NoError(t, err, "failed to load test data")

			// Apply setup function if provided
			if tc.setupFunc != nil {
				monitors = tc.setupFunc(monitors)
			}

			editor := tui.NewMonitorEditor(monitors)
			cmd := editor.SetMirror(tc.monitorID, tc.mirrorOf)

			require.NotNil(t, cmd, "SetMirror should return a command")

			msg := cmd()
			operationStatus, ok := msg.(tui.OperationStatus)
			require.True(t, ok, "command should return OperationStatus")

			if tc.shouldFail {
				assert.True(t, operationStatus.IsError(), "%s: %s", tc.reason, operationStatus.String())
			} else {
				assert.False(t, operationStatus.IsError(), "%s: %s", tc.reason, operationStatus.String())
			}
		})
	}
}

func TestToggleDisable_MonitorStates(t *testing.T) {
	testCases := []struct {
		name       string
		setupFile  string
		monitorID  int
		shouldFail bool
		reason     string
	}{
		// Valid operations
		{
			name:       "Disable monitor when multiple are enabled",
			setupFile:  "testdata/monitors_all_enabled.json",
			monitorID:  0,
			shouldFail: false,
			reason:     "should allow disabling when other monitors are enabled",
		},
		{
			name:       "Enable disabled monitor",
			setupFile:  "testdata/monitors_mixed_state.json",
			monitorID:  1, // MonitorB is disabled
			shouldFail: false,
			reason:     "should allow enabling a disabled monitor",
		},
		{
			name:       "Disable one of multiple enabled monitors",
			setupFile:  "testdata/monitors_mixed_state.json",
			monitorID:  2, // MonitorC is enabled, A is also enabled
			shouldFail: false,
			reason:     "should allow disabling when other monitors remain enabled",
		},
		// Invalid operations
		{
			name:       "Cannot disable last enabled monitor",
			setupFile:  "testdata/monitors_one_enabled.json",
			monitorID:  0, // MonitorA is the only enabled monitor
			shouldFail: true,
			reason:     "should prevent disabling the last enabled monitor",
		},
		{
			name:       "Invalid monitor ID",
			setupFile:  "testdata/monitors_all_enabled.json",
			monitorID:  999,
			shouldFail: true,
			reason:     "should fail for non-existent monitor ID",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			monitors, err := tui.LoadMonitorsFromJSON(tc.setupFile)
			require.NoError(t, err, "failed to load test data")

			editor := tui.NewMonitorEditor(monitors)
			cmd := editor.ToggleDisable(tc.monitorID)

			require.NotNil(t, cmd, "ToggleDisable should return a command")

			msg := cmd()
			operationStatus, ok := msg.(tui.OperationStatus)
			require.True(t, ok, "command should return OperationStatus")

			if tc.shouldFail {
				assert.True(t, operationStatus.IsError(), "%s: %s", tc.reason, operationStatus.String())
			} else {
				assert.False(t, operationStatus.IsError(), "%s: %s", tc.reason, operationStatus.String())
			}
		})
	}
}

func TestMoveMonitor(t *testing.T) {
	testCases := []struct {
		name         string
		setupFile    string
		monitorID    int
		dx           tui.Delta
		dy           tui.Delta
		expectedX    int
		expectedY    int
		shouldFail   bool
		expectedErr  string
		withSnapping bool
	}{
		{
			name:         "Move monitor right with snapping disabled",
			setupFile:    "testdata/monitors_for_movement.json",
			monitorID:    0,
			dx:           tui.DeltaMore,
			dy:           tui.DeltaNone,
			expectedX:    50,
			expectedY:    0,
			shouldFail:   false,
			withSnapping: false,
		},
		{
			name:         "Move monitor left with snapping disabled",
			setupFile:    "testdata/monitors_for_movement.json",
			monitorID:    1,
			dx:           tui.DeltaLess,
			dy:           tui.DeltaNone,
			expectedX:    1870,
			expectedY:    0,
			shouldFail:   false,
			withSnapping: false,
		},
		{
			name:         "Move monitor down with snapping disabled",
			setupFile:    "testdata/monitors_for_movement.json",
			monitorID:    0,
			dx:           tui.DeltaNone,
			dy:           tui.DeltaMore,
			expectedX:    0,
			expectedY:    50,
			shouldFail:   false,
			withSnapping: false,
		},
		{
			name:         "Move monitor up with snapping disabled",
			setupFile:    "testdata/monitors_for_movement.json",
			monitorID:    2,
			dx:           tui.DeltaNone,
			dy:           tui.DeltaLess,
			expectedX:    0,
			expectedY:    1030,
			shouldFail:   false,
			withSnapping: false,
		},
		{
			name:         "Move monitor diagonally with snapping disabled",
			setupFile:    "testdata/monitors_for_movement.json",
			monitorID:    0,
			dx:           tui.DeltaMore,
			dy:           tui.DeltaMore,
			expectedX:    50,
			expectedY:    50,
			shouldFail:   false,
			withSnapping: false,
		},
		{
			name:         "No movement when deltas are DeltaNone",
			setupFile:    "testdata/monitors_for_movement.json",
			monitorID:    0,
			dx:           tui.DeltaNone,
			dy:           tui.DeltaNone,
			expectedX:    0,
			expectedY:    0,
			shouldFail:   false,
			withSnapping: false,
		},
		{
			name:         "Move monitor with snapping enabled",
			setupFile:    "testdata/monitors_close_for_snapping.json",
			monitorID:    0,
			dx:           tui.DeltaMore,
			dy:           tui.DeltaNone,
			expectedX:    30,
			expectedY:    0,
			shouldFail:   false,
			withSnapping: true,
		},
		{
			name:        "Fail to move non-existent monitor",
			setupFile:   "testdata/monitors_for_movement.json",
			monitorID:   999,
			dx:          tui.DeltaMore,
			dy:          tui.DeltaNone,
			shouldFail:  true,
			expectedErr: "cant find monitor",
		},
		{
			name:        "Fail to move disabled monitor",
			setupFile:   "testdata/monitors_single_disabled.json",
			monitorID:   0,
			dx:          tui.DeltaMore,
			dy:          tui.DeltaNone,
			shouldFail:  true,
			expectedErr: "monitor is disabled",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			monitors, err := tui.LoadMonitorsFromJSON(tc.setupFile)
			require.NoError(t, err)

			editor := tui.NewMonitorEditor(monitors)
			editor.SetSnapping(tc.withSnapping)

			originalMonitor := findMonitorByID(monitors, tc.monitorID)
			if originalMonitor == nil {
				cmd := editor.MoveMonitor(tc.monitorID, tc.dx, tc.dy)
				require.NotNil(t, cmd)
				msg := cmd()
				operationStatus, ok := msg.(tui.OperationStatus)
				require.True(t, ok)
				assert.True(t, operationStatus.IsError())
				return
			}

			originalX := originalMonitor.X
			originalY := originalMonitor.Y
			cmd := editor.MoveMonitor(tc.monitorID, tc.dx, tc.dy)

			if tc.shouldFail {
				require.NotNil(t, cmd)
				msg := cmd()
				operationStatus, ok := msg.(tui.OperationStatus)
				require.True(t, ok)
				assert.True(t, operationStatus.IsError())
				if tc.expectedErr != "" {
					assert.Contains(t, operationStatus.String(), tc.expectedErr)
				}
				assert.Equal(t, originalX, originalMonitor.X)
				assert.Equal(t, originalY, originalMonitor.Y)
				return
			}

			if cmd != nil {
				msg := cmd()
				switch m := msg.(type) {
				case tui.OperationStatus:
					assert.False(t, m.IsError(), "Operation failed: %s", m.String())
				case tui.ShowGridLineCommand:
					// Grid line command is expected for snapping
				default:
					t.Fatalf("Unexpected message type: %T", msg)
				}
			}

			assert.Equal(t, tc.expectedX, originalMonitor.X, "Expected X: %d, Actual X: %d", tc.expectedX, originalMonitor.X)
			assert.Equal(t, tc.expectedY, originalMonitor.Y, "Expected Y: %d, Actual Y: %d", tc.expectedY, originalMonitor.Y)
		})
	}
}

func findMonitorByID(monitors []*tui.MonitorSpec, id int) *tui.MonitorSpec {
	for _, monitor := range monitors {
		if monitor.ID != nil && *monitor.ID == id {
			return monitor
		}
	}
	return nil
}
