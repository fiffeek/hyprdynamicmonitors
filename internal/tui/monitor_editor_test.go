package tui_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/fiffeek/hyprdynamicmonitors/internal/testutils"
	"github.com/fiffeek/hyprdynamicmonitors/internal/tui"
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
					ID:          testutils.IntPtr(3),
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
			monitors, err := testutils.LoadMonitorsFromJSON(tc.setupFile)
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
