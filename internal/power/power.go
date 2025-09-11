// Package power provides power state detection functionality.
package power

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"slices"
	"sync"

	"github.com/fiffeek/hyprdynamicmonitors/internal/config"
	"github.com/godbus/dbus/v5"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

type PowerState int

const (
	UnknownPowerState PowerState = iota
	BatteryPowerState
	ACPowerState
)

func (p PowerState) String() string {
	switch p {
	case BatteryPowerState:
		return "BAT"
	case ACPowerState:
		return "AC"
	default:
		return "UNKNOWN"
	}
}

type PowerEvent struct {
	State PowerState
}

type PowerDetector struct {
	cfg                *config.Config
	conn               *dbus.Conn
	events             chan PowerEvent
	signals            chan *dbus.Signal
	stateMu            sync.RWMutex
	dbusMatchOptions   [][]dbus.MatchOption
	disablePowerEvents bool
	powerState         PowerState
	powerStateMu       sync.RWMutex
}

func NewPowerDetector(ctx context.Context, cfg *config.Config, conn *dbus.Conn, disablePowerEvents bool) (*PowerDetector, error) {
	detector := &PowerDetector{
		conn:               conn,
		cfg:                cfg,
		events:             make(chan PowerEvent, 10),
		signals:            make(chan *dbus.Signal, 10),
		disablePowerEvents: disablePowerEvents,
		powerStateMu:       sync.RWMutex{},
	}

	ps, err := detector.getCurrentState(ctx)
	if err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("UPower not available or accessible: %w", err)
	}
	detector.powerState = ps

	logrus.Info("UPower D-Bus power detection initialized")

	return detector, nil
}

func (p *PowerDetector) GetCurrentState() PowerState {
	p.powerStateMu.RLock()
	defer p.powerStateMu.RUnlock()
	return p.powerState
}

func (p *PowerDetector) getCurrentState(ctx context.Context) (PowerState, error) {
	cfg := p.cfg.Get().PowerEvents
	if p.disablePowerEvents {
		logrus.WithFields(logrus.Fields{"default": ACPowerState.String()}).Debug(
			"Power events are disabled, returning the default value")
		return ACPowerState, nil
	}

	obj := p.conn.Object(
		cfg.DbusQueryObject.Destination,
		dbus.ObjectPath(cfg.DbusQueryObject.Path))

	logrus.WithFields(logrus.Fields{
		"destination": cfg.DbusQueryObject.Destination,
		"path":        cfg.DbusQueryObject.Path,
		"method":      cfg.DbusQueryObject.Method,
		"args":        cfg.DbusQueryObject.CollectArgs(),
	}).Debug("About to make D-Bus method call")

	var onBattery dbus.Variant
	err := obj.CallWithContext(ctx, cfg.DbusQueryObject.Method, 0,
		cfg.DbusQueryObject.CollectArgs()...).Store(&onBattery)
	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"destination": cfg.DbusQueryObject.Destination,
			"path":        cfg.DbusQueryObject.Path,
			"method":      cfg.DbusQueryObject.Method,
		}).Error("D-Bus method call failed")
		return BatteryPowerState, fmt.Errorf("failed to get property from UPower: %w", err)
	}

	logrus.Debug("D-Bus method call succeeded")

	state := onBattery.String()

	logrus.WithFields(logrus.Fields{
		"upower_reported_state": state,
		"expected":              cfg.DbusQueryObject.ExpectedDischargingValue,
	}).Debug("UPower state property")
	if state == cfg.DbusQueryObject.ExpectedDischargingValue {
		return BatteryPowerState, nil
	}
	return ACPowerState, nil
}

func (p *PowerDetector) Listen() <-chan PowerEvent {
	return p.events
}

func (p *PowerDetector) createMatchRules() [][]dbus.MatchOption {
	rules := [][]dbus.MatchOption{}
	for _, rule := range p.cfg.Get().PowerEvents.DbusSignalMatchRules {
		matchRules := []dbus.MatchOption{}
		if rule.Interface != nil {
			matchRules = append(matchRules, dbus.WithMatchInterface(*rule.Interface))
		}
		if rule.Sender != nil {
			matchRules = append(matchRules, dbus.WithMatchSender(*rule.Sender))
		}
		if rule.Member != nil {
			matchRules = append(matchRules, dbus.WithMatchMember(*rule.Member))
		}
		if rule.ObjectPath != nil {
			matchRules = append(matchRules, dbus.WithMatchObjectPath(dbus.ObjectPath(*rule.ObjectPath)))
		}
		rules = append(rules, matchRules)
	}
	return rules
}

func (p *PowerDetector) getExpectedSignalNames() []string {
	result := []string{}
	for _, filter := range p.cfg.Get().PowerEvents.DbusSignalReceiveFilters {
		result = append(result, *filter.Name)
	}
	return result
}

func (p *PowerDetector) Reload(ctx context.Context) error {
	if p.disablePowerEvents {
		logrus.Debug("Power events are disabled, not reloading power rules")
		return nil
	}

	p.stateMu.Lock()
	defer p.stateMu.Unlock()
	rules := p.createMatchRules()
	if reflect.DeepEqual(rules, p.dbusMatchOptions) {
		logrus.Debug("power events rules match, nothing to be done")
	}

	for _, ruleSet := range p.dbusMatchOptions {
		if err := p.conn.RemoveMatchSignalContext(ctx, ruleSet...); err != nil {
			logrus.WithFields(logrus.Fields{
				"error": err,
			}).Debug("Failed to remove D-Bus match rule")
			return fmt.Errorf("cant remove signal rule for dbus: %w", err)
		}
	}

	for _, ruleSet := range rules {
		if err := p.conn.AddMatchSignalContext(ctx, ruleSet...); err != nil {
			logrus.WithFields(logrus.Fields{
				"error": err,
			}).Debug("Failed to add D-Bus match rule")
			return fmt.Errorf("cant add signal rule for dbus: %w", err)
		}
	}

	p.dbusMatchOptions = rules
	logrus.Debug("Reloaded power detector")
	return nil
}

func (p *PowerDetector) Run(ctx context.Context) error {
	eg, ctx := errgroup.WithContext(ctx)

	eg.Go(func() error {
		<-ctx.Done()
		logrus.Debug("Power detector context cancelled, closing D-Bus connection")
		if p.conn != nil {
			_ = p.conn.Close()
		}
		return context.Cause(ctx)
	})

	if p.disablePowerEvents {
		logrus.Info("Power events are disabled, waiting for ctx cancellation")
		return eg.Wait()
	}

	if err := p.Reload(ctx); err != nil {
		return fmt.Errorf("cant reload: %w", err)
	}
	p.conn.Signal(p.signals)

	eg.Go(func() error {
		defer close(p.events)
		defer p.conn.RemoveSignal(p.signals)

		logrus.Debug("Power detector started, listening for UPower D-Bus signals")

		lastState := UnknownPowerState

		for {
			select {
			case signal, ok := <-p.signals:
				if !ok {
					return errors.New("dbus power events channel closed")
				}
				logrus.WithFields(logrus.Fields{
					"signal_name": signal.Name,
					"signal_path": signal.Path,
				}).Debug("Received D-Bus signal")

				var currentState PowerState
				var err error

				if slices.Contains(p.getExpectedSignalNames(), signal.Name) {
					logrus.Debug("Signal matches expected name, calling getCurrentState")
					currentState, err = p.getCurrentState(ctx)
					if err != nil {
						logrus.WithError(err).Error("getCurrentState failed")
						return fmt.Errorf("failed to get power state after signal %s: %w", signal.Name, err)
					}

					logrus.WithFields(logrus.Fields{"state": currentState}).Debug("getCurrentState succeeded")
					p.powerStateMu.Lock()
					p.powerState = currentState
					p.powerStateMu.Unlock()
				} else {
					logrus.WithField("signal_name", signal.Name).Debug("Ignoring unknown UPower signal")
					continue
				}

				if currentState == lastState {
					logrus.WithField("power_state", currentState.String()).Debug("Power state unchanged after signal")
					continue
				}

				logrus.WithFields(logrus.Fields{
					"from": lastState.String(),
					"to":   currentState.String(),
				}).Info("Power state changed")

				select {
				case p.events <- PowerEvent{State: currentState}:
					lastState = currentState
				case <-ctx.Done():
					return context.Cause(ctx)
				}
			case <-ctx.Done():
				logrus.Debug("Power detector context cancelled, shutting down")
				return context.Cause(ctx)
			}
		}
	})

	if err := eg.Wait(); err != nil {
		return fmt.Errorf("goroutines for power detector failed %w", err)
	}
	return nil
}
