package power_test

import (
	"context"
	"testing"
	"time"

	"github.com/fiffeek/hyprdynamicmonitors/internal/power"
	"github.com/fiffeek/hyprdynamicmonitors/internal/testutils"
	"github.com/fiffeek/hyprdynamicmonitors/internal/utils"
	"github.com/godbus/dbus/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLidDetector_Integration(t *testing.T) {
	service, testBusName, testObjectPath, cleanup := testutils.SetupTestDbusService(t)
	defer cleanup()

	time.Sleep(100 * time.Millisecond)

	cfg := testutils.CreateTestLidConfig(t, testBusName, testObjectPath)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	conn, err := dbus.ConnectSessionBus()
	require.NoError(t, err, "failed to connect to session bus")
	defer conn.Close()

	detector, err := power.NewLidStateDetector(ctx, cfg, conn, true)
	require.NoError(t, err, "should be able to create test power detector")

	initialState := detector.GetCurrentState()
	require.NoError(t, err, "should be able to get initial state")
	require.Equal(t, power.ClosedLidState, initialState, "initial state should be closed")

	ctx2, cancel2 := context.WithCancel(ctx)
	defer cancel2()

	go func() {
		err := detector.Run(ctx2)
		if err != nil && ctx2.Err() == nil {
			t.Errorf("detector run failed: %v", err)
		}
	}()

	time.Sleep(200 * time.Millisecond)

	service.SetLidProperty(power.OpenedLidState)
	err = service.EmitSignal()
	require.NoError(t, err, "should be able to emit signal")

	select {
	case event := <-detector.Listen():
		require.Equal(t, power.OpenedLidState, event.State, "should receive open event")
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for power event")
	}

	service.SetLidProperty(power.ClosedLidState)
	err = service.EmitSignal()
	require.NoError(t, err, "should be able to emit signal")

	select {
	case event := <-detector.Listen():
		require.Equal(t, power.ClosedLidState, event.State, "should receive close event")
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for battery event")
	}

	// same signal does not result in an event
	service.SetLidProperty(power.ClosedLidState)
	err = service.EmitSignal()
	require.NoError(t, err, "should be able to emit signal")
	// but the new one does so we wait on the first event after updates
	service.SetLidProperty(power.OpenedLidState)
	err = service.EmitSignal()
	require.NoError(t, err, "should be able to emit signal")
	select {
	case event := <-detector.Listen():
		require.Equal(t, power.OpenedLidState, event.State, "should receive open event")
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for battery event")
	}
}

func TestLidDetector_DisabledEvents_Simple(t *testing.T) {
	ctx := context.Background()
	cfg := testutils.NewTestConfig(t).Get()
	detector, err := power.NewLidStateDetector(ctx, cfg, nil, false)
	require.NoError(t, err, "cant get new power detector")

	tests := []struct {
		name   string
		method func() error
	}{
		{utils.GetFunctionName(detector.GetCurrentState), func() error {
			state := detector.GetCurrentState()
			assert.Equal(t, power.OpenedLidState, state)
			return nil
		}},
		{utils.GetFunctionName(detector.Reload), func() error {
			return detector.Reload(ctx)
		}},
		{utils.GetFunctionName(detector.Run), func() error {
			ctx, cancel := context.WithCancel(ctx)

			errCh := make(chan error, 1)
			go func() {
				errCh <- detector.Run(ctx)
			}()
			cancel()
			select {
			case err := <-errCh:
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "context canceled")
			case <-time.After(1 * time.Second):
				t.Fatal("timeout waiting for service to shut down")
			}
			return nil
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.NoError(t, tt.method(), "cant run function", tt.name)
		})
	}
}
