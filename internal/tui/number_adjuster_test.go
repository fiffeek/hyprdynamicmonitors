package tui_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/fiffeek/hyprdynamicmonitors/internal/tui"
)

func TestNumberAdjuster_Update(t *testing.T) {
	testCases := []struct {
		name          string
		current       float64
		max           float64
		min           float64
		baseIncrement float64
		action        func(*tui.NumberAdjuster)
		expectedValue float64
	}{
		{
			name:          "Increase from initial value",
			current:       1.0,
			max:           10.0,
			min:           0.1,
			baseIncrement: 0.01,
			action: func(na *tui.NumberAdjuster) {
				na.Increase()
			},
			expectedValue: 1.01,
		},
		{
			name:          "Decrease from initial value",
			current:       1.5,
			max:           10.0,
			min:           0.1,
			baseIncrement: 0.01,
			action: func(na *tui.NumberAdjuster) {
				na.Decrease()
			},
			expectedValue: 1.49,
		},
		{
			name:          "Cannot go below minimum",
			current:       0.1,
			max:           10.0,
			min:           0.1,
			baseIncrement: 0.01,
			action: func(na *tui.NumberAdjuster) {
				na.Decrease()
			},
			expectedValue: 0.1,
		},
		{
			name:          "Cannot go above maximum",
			current:       10.0,
			max:           10.0,
			min:           0.1,
			baseIncrement: 0.01,
			action: func(na *tui.NumberAdjuster) {
				na.Increase()
			},
			expectedValue: 10.0,
		},
		{
			name:          "SetCurrent updates value",
			current:       1.0,
			max:           10.0,
			min:           0.1,
			baseIncrement: 0.01,
			action: func(na *tui.NumberAdjuster) {
				na.SetCurrent(5.5)
			},
			expectedValue: 5.5,
		},
		{
			name:          "Increase with larger base increment",
			current:       100.0,
			max:           1000.0,
			min:           0.0,
			baseIncrement: 10.0,
			action: func(na *tui.NumberAdjuster) {
				na.Increase()
			},
			expectedValue: 110.0,
		},
		{
			name:          "Decrease with larger base increment",
			current:       100.0,
			max:           1000.0,
			min:           0.0,
			baseIncrement: 10.0,
			action: func(na *tui.NumberAdjuster) {
				na.Decrease()
			},
			expectedValue: 90.0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			adjuster := tui.NewNumberAdjuster(tc.current, tc.max, tc.min, tc.baseIncrement)

			tc.action(adjuster)

			assert.InDelta(t, tc.expectedValue, adjuster.Value(), 0.001,
				"Expected value: %.2f, Actual value: %.2f", tc.expectedValue, adjuster.Value())
		})
	}
}

func TestNumberAdjuster_AcceleratedAdjustment(t *testing.T) {
	adjuster := tui.NewNumberAdjuster(1.0, 10.0, 0.1, 0.01)

	// First increase - normal speed
	adjuster.Increase()
	assert.InDelta(t, 1.01, adjuster.Value(), 0.001)

	// Second increase immediately after - should start accelerating
	adjuster.Increase()
	assert.InDelta(t, 1.02, adjuster.Value(), 0.001)

	// Wait more than 200ms - should reset acceleration
	time.Sleep(250 * time.Millisecond)
	adjuster.Increase()
	assert.InDelta(t, 1.03, adjuster.Value(), 0.001)
}

func TestNumberAdjuster_MultipleRapidIncreases(t *testing.T) {
	adjuster := tui.NewNumberAdjuster(1.0, 10.0, 0.1, 0.01)

	// Perform multiple rapid increases to test acceleration
	for i := 0; i < 7; i++ {
		adjuster.Increase()
	}

	// After 7 rapid increases with acceleration, value should be significantly higher
	// Expected: 1.0 + 0.01 + 0.01 + 0.02 + 0.02 + 0.05 + 0.05 + 0.10 = 1.26
	assert.Greater(t, adjuster.Value(), 1.2, "Value should have increased with acceleration")
}

func TestNumberAdjuster_Value(t *testing.T) {
	adjuster := tui.NewNumberAdjuster(2.5, 10.0, 0.1, 0.01)
	assert.Equal(t, 2.5, adjuster.Value(), "Value() should return current value")
}
