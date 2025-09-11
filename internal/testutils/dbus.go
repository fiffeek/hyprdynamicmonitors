package testutils

import (
	"crypto/rand"
	"fmt"
	"regexp"
	"testing"
	"time"

	"github.com/fiffeek/hyprdynamicmonitors/internal/config"
	"github.com/fiffeek/hyprdynamicmonitors/internal/power"
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

func GenerateTestBusName() string {
	re := regexp.MustCompile(`[0-9]+`)
	return re.ReplaceAllString("com.test.PowerDetector."+rand.Text(), "")
}

func GenerateTestObjectPath() string {
	return "/com/test/PowerDetector/" + rand.Text()
}

// TestDbusService provides a mock D-Bus service for testing
type TestDbusService struct {
	conn           *dbus.Conn
	propertyValue  bool
	testObjectPath string
	t              *testing.T
}

func NewTestDbusService(t *testing.T, conn *dbus.Conn, initial power.PowerState, objectPath string) *TestDbusService {
	s := &TestDbusService{
		t:              t,
		conn:           conn,
		testObjectPath: objectPath,
	}
	s.SetProperty(initial)
	return s
}

func (s *TestDbusService) Get(interfaceName, propertyName string) (dbus.Variant, *dbus.Error) {
	if interfaceName == testInterface && propertyName == testProperty {
		return dbus.MakeVariant(s.propertyValue), nil
	}
	return dbus.MakeVariant(false), dbus.NewError("org.freedesktop.DBus.Error.InvalidArgs", nil)
}

func (s *TestDbusService) SetProperty(value power.PowerState) {
	switch value {
	case power.ACPowerState:
		s.propertyValue = false
	case power.BatteryPowerState:
		s.propertyValue = true
	default:
		assert.True(s.t, false, "unknown power state")
	}
}

func (s *TestDbusService) EmitSignal() error {
	if err := s.conn.Emit(dbus.ObjectPath(s.testObjectPath), testSignalName); err != nil {
		return fmt.Errorf("cant emit signal: %w", err)
	}
	return nil
}

func CreatePowerConfig(busName, objectPath string) *config.PowerSection {
	return &config.PowerSection{
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
	}
}

func CreateTestPowerConfig(t *testing.T, busName, objectPath string) *config.Config {
	return NewTestConfig(t).WithPowerSection(CreatePowerConfig(busName, objectPath)).Get()
}

func SetupTestDbusService(t *testing.T) (*TestDbusService, string, string, func()) {
	conn, err := dbus.ConnectSessionBus()
	require.NoError(t, err, "failed to connect to session bus")

	testBusName := GenerateTestBusName()
	testObjectPath := GenerateTestObjectPath()
	Logf(t, "dbus, object: %s, bus: %s", testObjectPath, testBusName)
	reply, err := conn.RequestName(testBusName, dbus.NameFlagDoNotQueue)
	require.NoError(t, err, "failed to request bus name")
	require.Equal(t, dbus.RequestNameReplyPrimaryOwner, reply, "failed to become primary owner")

	service := NewTestDbusService(t, conn, power.BatteryPowerState, testObjectPath)

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

func SetupFakeDbusEventsServer(t *testing.T, service *TestDbusService, events []power.PowerState,
	initialSleep, sleepBetweenEvents time.Duration, binaryStarting <-chan struct{},
) chan struct{} {
	serverDone := make(chan struct{})
	go func() {
		defer close(serverDone)
		Logf(t, "Waiting for the server to start")
		<-binaryStarting
		Logf(t, "Starting fake dbus")
		time.Sleep(initialSleep)
		Logf(t, "Will start sending events")

		for _, event := range events {
			service.SetProperty(event)
			require.NoError(t, service.EmitSignal(), "cant emit fake dbus signal")
			Logf(t, "Emitted signal")
			time.Sleep(sleepBetweenEvents)
		}
	}()
	return serverDone
}
