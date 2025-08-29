package utils_test

import (
	"context"
	"testing"
	"time"

	"github.com/fiffeek/hyprdynamicmonitors/internal/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDebouncer_Basic(t *testing.T) {
	debouncer := utils.NewDebouncer()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	errCh := make(chan error, 1)
	go func() {
		errCh <- debouncer.Run(ctx)
	}()

	callCount := 0
	var lastCallTime time.Time

	testFn := func(ctx context.Context) error {
		callCount++
		lastCallTime = time.Now()
		t.Logf("Test function called, count: %d", callCount)
		return nil
	}

	// Schedule multiple calls quickly - should be debounced to one
	start := time.Now()
	debouncer.Do(ctx, 100*time.Millisecond, testFn)
	debouncer.Do(ctx, 100*time.Millisecond, testFn)
	debouncer.Do(ctx, 100*time.Millisecond, testFn)

	// Wait for debounced execution
	time.Sleep(200 * time.Millisecond)

	// Should have been called exactly once
	assert.Equal(t, 1, callCount, "Function should be called exactly once due to debouncing")

	// Check timing - function should have been called after the debounce delay
	timeSinceStart := lastCallTime.Sub(start)
	assert.True(t, timeSinceStart >= 100*time.Millisecond, "Function should be called after debounce delay")

	// Cancel and wait for shutdown
	cancel()
	select {
	case err := <-errCh:
		require.Error(t, err)
		assert.Contains(t, err.Error(), "context canceled")
	case <-time.After(500 * time.Millisecond):
		t.Fatal("timeout waiting for debouncer to shutdown")
	}
}

func TestDebouncer_Cancel(t *testing.T) {
	debouncer := utils.NewDebouncer()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Start the debouncer
	errCh := make(chan error, 1)
	go func() {
		errCh <- debouncer.Run(ctx)
	}()

	// Track function calls
	callCount := 0
	testFn := func(ctx context.Context) error {
		callCount++
		t.Logf("Test function called, count: %d", callCount)
		return nil
	}

	// Schedule a call, then cancel before it executes
	debouncer.Do(ctx, 200*time.Millisecond, testFn)
	time.Sleep(50 * time.Millisecond) // Wait a bit but not long enough for execution
	debouncer.Cancel()

	// Wait longer than the original delay
	time.Sleep(300 * time.Millisecond)

	// Function should not have been called
	assert.Equal(t, 0, callCount, "Function should not be called after cancel")

	// Cancel and wait for shutdown
	cancel()
	select {
	case err := <-errCh:
		require.Error(t, err)
		assert.Contains(t, err.Error(), "context canceled")
	case <-time.After(500 * time.Millisecond):
		t.Fatal("timeout waiting for debouncer to shutdown")
	}
}
