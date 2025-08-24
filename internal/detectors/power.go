package detectors

import (
	"fmt"

	"github.com/godbus/dbus/v5"
	"github.com/sirupsen/logrus"
)

type PowerState int

const (
	Battery PowerState = iota
	ACPower
)

func (p PowerState) String() string {
	switch p {
	case Battery:
		return "BAT"
	case ACPower:
		return "AC"
	default:
		return "UNKNOWN"
	}
}

type PowerEvent struct {
	State PowerState
}

type PowerDetector struct {
	conn *dbus.Conn
}

func NewPowerDetector() (*PowerDetector, error) {
	conn, err := dbus.ConnectSystemBus()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to system D-Bus: %w", err)
	}

	detector := &PowerDetector{
		conn: conn,
	}

	if _, err := detector.GetCurrentState(); err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("UPower not available or accessible: %w", err)
	}

	logrus.Info("UPower D-Bus power detection initialized")

	return detector, nil
}

func (p *PowerDetector) GetCurrentState() (PowerState, error) {
	obj := p.conn.Object("org.freedesktop.UPower", "/org/freedesktop/UPower")

	var onBattery dbus.Variant
	err := obj.Call("org.freedesktop.DBus.Properties.Get", 0, "org.freedesktop.UPower", "OnBattery").Store(&onBattery)
	if err != nil {
		return Battery, fmt.Errorf("failed to get OnBattery property from UPower: %w", err)
	}

	if onBatteryValue, ok := onBattery.Value().(bool); ok {
		logrus.WithField("on_battery", onBatteryValue).Debug("UPower OnBattery property")
		if onBatteryValue {
			return Battery, nil
		}
		return ACPower, nil
	}

	return Battery, fmt.Errorf("unexpected OnBattery value type: %T", onBattery.Value())
}

func (p *PowerDetector) Listen() (<-chan PowerEvent, error) {
	events := make(chan PowerEvent, 10)
	rules := []string{
		"type='signal',sender='org.freedesktop.UPower',interface='org.freedesktop.DBus.Properties',member='PropertiesChanged',path='org.freedesktop.UPower'",
	}

	for _, rule := range rules {
		err := p.conn.AddMatchSignal(dbus.WithMatchInterface("org.freedesktop.UPower"))
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"rule":  rule,
				"error": err,
			}).Debug("Failed to add D-Bus match rule")
			return nil, fmt.Errorf("cant add signal rule for dbus: %v", err)
		}
	}

	signals := make(chan *dbus.Signal, 10)
	p.conn.Signal(signals)

	go func() {
		defer close(events)
		defer p.conn.RemoveSignal(signals)

		logrus.Debug("Power detector started, listening for UPower D-Bus signals")

		lastState := ACPower

		for signal := range signals {
			logrus.WithFields(logrus.Fields{
				"signal_name": signal.Name,
				"signal_path": signal.Path,
			}).Debug("Received D-Bus signal")

			var currentState PowerState
			var err error
			if signal.Name == "org.freedesktop.UPower.DeviceAdded" || signal.Name == "org.freedesktop.UPower.DeviceRemoved" {
				currentState, err = p.GetCurrentState()
				if err != nil {
					logrus.WithError(err).Debug("Failed to read power state after signal")
					continue
				}
			}

			if currentState != lastState {
				logrus.WithFields(logrus.Fields{
					"from": lastState.String(),
					"to":   currentState.String(),
				}).Info("Power state changed")
				events <- PowerEvent{State: currentState}
				lastState = currentState
			} else {
				logrus.WithField("power_state", currentState.String()).Debug("Power state unchanged after signal")
			}
		}
	}()

	return events, nil
}
