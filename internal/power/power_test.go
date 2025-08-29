package power_test

import (
	"context"
	"crypto/rand"
	"fmt"
	"regexp"
	"testing"
	"time"

	"github.com/fiffeek/hyprdynamicmonitors/internal/config"
	"github.com/fiffeek/hyprdynamicmonitors/internal/power"
	"github.com/fiffeek/hyprdynamicmonitors/internal/testutils"
	"github.com/fiffeek/hyprdynamicmonitors/internal/utils"
	"github.com/godbus/dbus/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testInterface  = "com.test.PowerDetector.Interface"
	testSignalName = "com.test.PowerDetector.Interface.PowerChanged"
	testProperty   = "TestProperty"
	testMethodName = "org.freedesktop.DBus.Properties.Get"
	testMemberName = "PowerChanged"
)

func generateTestBusName() string {
	re := regexp.MustCompile(`[0-9]+`)
	return re.ReplaceAllString("com.test.PowerDetector."+rand.Text(), "")
}

func generateTestObjectPath() string {
	return "/com/test/PowerDetector/" + rand.Text()
}

// testDbusService provides a mock D-Bus service for testing
type testDbusService struct {
	conn           *dbus.Conn
	propertyValue  bool
	testObjectPath string
}

func (s *testDbusService) Get(interfaceName, propertyName string) (dbus.Variant, *dbus.Error) {
	if interfaceName == testInterface && propertyName == testProperty {
		return dbus.MakeVariant(s.propertyValue), nil
	}
	return dbus.MakeVariant(false), dbus.NewError("org.freedesktop.DBus.Error.InvalidArgs", nil)
}

func (s *testDbusService) SetProperty(value bool) {
	s.propertyValue = value
}

func (s *testDbusService) EmitSignal() error {
	if err := s.conn.Emit(dbus.ObjectPath(s.testObjectPath), testSignalName); err != nil {
		return fmt.Errorf("cant emit signal: %w", err)
	}
	return nil
}

func setupTestDbusService(t *testing.T) (*testDbusService, string, string, func()) {
	conn, err := dbus.ConnectSessionBus()
	require.NoError(t, err, "failed to connect to session bus")

	testBusName := generateTestBusName()
	testObjectPath := generateTestObjectPath()
	t.Log(testObjectPath, testBusName)
	reply, err := conn.RequestName(testBusName, dbus.NameFlagDoNotQueue)
	require.NoError(t, err, "failed to request bus name")
	require.Equal(t, dbus.RequestNameReplyPrimaryOwner, reply, "failed to become primary owner")

	service := &testDbusService{
		conn:           conn,
		propertyValue:  true,
		testObjectPath: testObjectPath,
	}

	err = conn.Export(service, dbus.ObjectPath(testObjectPath), testInterface)
	require.NoError(t, err, "failed to export test service")

	err = conn.Export(service, dbus.ObjectPath(testObjectPath), "org.freedesktop.DBus.Properties")
	require.NoError(t, err, "failed to export properties interface")

	cleanup := func() {
		_, _ = conn.ReleaseName(testBusName)
		_ = conn.Close()
	}

	return service, testBusName, testObjectPath, cleanup
}

func createTestPowerConfig(t *testing.T, busName, objectPath string) *config.Config {
	return testutils.NewTestConfig(t).WithPowerSection(
		&config.PowerSection{
			DbusSignalMatchRules: []*config.DbusSignalMatchRule{
				{
					Interface:  utils.StringPtr(testInterface),
					Member:     utils.StringPtr(testMemberName),
					ObjectPath: utils.StringPtr(objectPath),
				},
			},
			DbusSignalReceiveFilters: []*config.DbusSignalReceiveFilter{
				{Name: utils.StringPtr(testSignalName)},
			},
			DbusQueryObject: &config.DbusQueryObject{
				Destination:              busName,
				Path:                     objectPath,
				Method:                   testMethodName,
				ExpectedDischargingValue: "true", // true means on battery
				Args: []config.DbusQueryObjectArg{
					{Arg: testInterface},
					{Arg: testProperty},
				},
			},
		}).Get()
}

func TestPowerDetector_Integration(t *testing.T) {
	service, testBusName, testObjectPath, cleanup := setupTestDbusService(t)
	defer cleanup()

	time.Sleep(100 * time.Millisecond)

	cfg := createTestPowerConfig(t, testBusName, testObjectPath)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	conn, err := dbus.ConnectSessionBus()
	require.NoError(t, err, "failed to connect to session bus")
	defer conn.Close()

	detector, err := power.NewPowerDetector(ctx, cfg, conn, false)
	require.NoError(t, err, "should be able to create test power detector")

	initialState, err := detector.GetCurrentState(ctx)
	require.NoError(t, err, "should be able to get initial state")
	require.Equal(t, power.BatteryPowerState, initialState, "initial state should be ac")

	ctx2, cancel2 := context.WithCancel(ctx)
	defer cancel2()

	go func() {
		err := detector.Run(ctx2)
		if err != nil && ctx2.Err() == nil {
			t.Errorf("detector run failed: %v", err)
		}
	}()

	time.Sleep(200 * time.Millisecond)

	service.SetProperty(true)
	err = service.EmitSignal()
	require.NoError(t, err, "should be able to emit signal")

	select {
	case event := <-detector.Listen():
		require.Equal(t, power.BatteryPowerState, event.State, "should receive bat event")
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for power event")
	}

	service.SetProperty(false)
	err = service.EmitSignal()
	require.NoError(t, err, "should be able to emit signal")

	select {
	case event := <-detector.Listen():
		require.Equal(t, power.ACPowerState, event.State, "should receive ac event")
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for battery event")
	}

	// same signal does not result in an event
	service.SetProperty(false)
	err = service.EmitSignal()
	require.NoError(t, err, "should be able to emit signal")
	// but the new one does so we wait on the firt event after updates
	service.SetProperty(true)
	err = service.EmitSignal()
	require.NoError(t, err, "should be able to emit signal")
	select {
	case event := <-detector.Listen():
		require.Equal(t, power.BatteryPowerState, event.State, "should receive battery event")
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for battery event")
	}
}

func TestPowerDetector_DisabledPowerEvents_Simple(t *testing.T) {
	ctx := context.Background()
	cfg := testutils.NewTestConfig(t).Get()
	detector, err := power.NewPowerDetector(ctx, cfg, nil, true)
	require.NoError(t, err, "cant get new power detector")

	tests := []struct {
		name   string
		method func() error
	}{
		{utils.GetFunctionName(detector.GetCurrentState), func() error {
			if _, err := detector.GetCurrentState(ctx); err != nil {
				return fmt.Errorf("cant get current state: %w", err)
			}
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
