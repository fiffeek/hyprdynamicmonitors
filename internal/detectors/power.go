package detectors

import (
	"context"
	"fmt"

	"github.com/godbus/dbus/v5"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
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
	conn    *dbus.Conn
	events  chan PowerEvent
	signals chan *dbus.Signal
}

func NewPowerDetector(ctx context.Context) (*PowerDetector, error) {
	conn, err := dbus.ConnectSystemBus()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to system D-Bus: %w", err)
	}

	detector := &PowerDetector{
		conn:    conn,
		events:  make(chan PowerEvent, 10),
		signals: make(chan *dbus.Signal, 10),
	}

	if _, err := detector.GetCurrentState(ctx); err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("UPower not available or accessible: %w", err)
	}

	logrus.Info("UPower D-Bus power detection initialized")

	return detector, nil
}

func (p *PowerDetector) GetCurrentState(ctx context.Context) (PowerState, error) {
	obj := p.conn.Object("org.freedesktop.UPower", "/org/freedesktop/UPower")

	var onBattery dbus.Variant
	err := obj.CallWithContext(ctx, "org.freedesktop.DBus.Properties.Get", 0, "org.freedesktop.UPower", "OnBattery").Store(&onBattery)
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

func (p *PowerDetector) Listen() <-chan PowerEvent {
	return p.events
}

func (p *PowerDetector) Run(ctx context.Context) error {
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
			return fmt.Errorf("cant add signal rule for dbus: %v", err)
		}
	}

	p.conn.Signal(p.signals)

	eg, ctx := errgroup.WithContext(ctx)

	eg.Go(func() error {
		<-ctx.Done()
		logrus.Debug("Power detector context cancelled, closing D-Bus connection")
		return p.conn.Close()
	})

	eg.Go(func() error {
		defer close(p.events)
		defer p.conn.RemoveSignal(p.signals)

		logrus.Debug("Power detector started, listening for UPower D-Bus signals")

		lastState := ACPower

		for {
			select {
			case signal, ok := <-p.signals:
				if !ok {
					return fmt.Errorf("dbus power events channel closed")
				}
				logrus.WithFields(logrus.Fields{
					"signal_name": signal.Name,
					"signal_path": signal.Path,
				}).Debug("Received D-Bus signal")

				var currentState PowerState
				var err error
				if signal.Name == "org.freedesktop.UPower.DeviceAdded" || signal.Name == "org.freedesktop.UPower.DeviceRemoved" {
					currentState, err = p.GetCurrentState(ctx)
					if err != nil {
						return fmt.Errorf("error extracting the current state: %v", currentState)
					}
				}

				if currentState != lastState {
					logrus.WithFields(logrus.Fields{
						"from": lastState.String(),
						"to":   currentState.String(),
					}).Info("Power state changed")
					select {
					case p.events <- PowerEvent{State: currentState}:
						lastState = currentState
					case <-ctx.Done():
						return ctx.Err()
					}
				} else {
					logrus.WithField("power_state", currentState.String()).Debug("Power state unchanged after signal")
				}
			case <-ctx.Done():
				logrus.Debug("Power detector context cancelled, shutting down")
				return ctx.Err()
			}
		}
	})

	return eg.Wait()
}
