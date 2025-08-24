package detectors

import (
	"fmt"
	"log"

	"github.com/godbus/dbus/v5"
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
	verbose bool
}

func NewPowerDetector(verbose bool) (*PowerDetector, error) {
	conn, err := dbus.ConnectSystemBus()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to system D-Bus: %w", err)
	}

	detector := &PowerDetector{
		conn:    conn,
		verbose: verbose,
	}

	if _, err := detector.GetCurrentState(); err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("UPower not available or accessible: %w", err)
	}

	if verbose {
		log.Printf("UPower D-Bus power detection initialized")
	}

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
		if p.verbose {
			log.Printf("UPower OnBattery: %v", onBatteryValue)
		}
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
			if p.verbose {
				log.Printf("Failed to add match rule %s: %v", rule, err)
			}
			return nil, fmt.Errorf("cant add signal rule for dbus: %v", err)
		}
	}

	// Create signal channel
	signals := make(chan *dbus.Signal, 10)
	p.conn.Signal(signals)

	go func() {
		defer close(events)
		defer p.conn.RemoveSignal(signals)

		if p.verbose {
			log.Printf("Power detector started, listening for UPower D-Bus signals")
		}

		lastState := ACPower

		for signal := range signals {
			if p.verbose {
				log.Printf("Received D-Bus signal: %s from %s", signal.Name, signal.Path)
			}

			var currentState PowerState
			var err error
			if signal.Name == "org.freedesktop.UPower.DeviceAdded" || signal.Name == "org.freedesktop.Upower.DeviceRemoved" {
				currentState, err = p.GetCurrentState()
				if err != nil {
					if p.verbose {
						log.Printf("Failed to read power state after signal: %v", err)
					}
					continue
				}
			}

			if currentState != lastState {
				if p.verbose {
					log.Printf("Power state changed: %s -> %s", lastState, currentState)
				}
				events <- PowerEvent{State: currentState}
				lastState = currentState
			} else if p.verbose {
				log.Printf("Power state unchanged after signal: %s", currentState)
			}
		}
	}()

	return events, nil
}
