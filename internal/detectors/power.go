// Package detectors provides power state detection functionality.
package detectors

import (
	"context"
	"errors"
	"fmt"
	"slices"

	"github.com/fiffeek/hyprdynamicmonitors/internal/config"
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
	cfg     *config.PowerSection
	conn    *dbus.Conn
	events  chan PowerEvent
	signals chan *dbus.Signal
}

func NewPowerDetector(ctx context.Context, cfg *config.PowerSection, conn *dbus.Conn) (*PowerDetector, error) {
	detector := &PowerDetector{
		conn:    conn,
		cfg:     cfg,
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
	obj := p.conn.Object(
		p.cfg.DbusQueryObject.Destination,
		dbus.ObjectPath(p.cfg.DbusQueryObject.Path))

	var onBattery dbus.Variant
	err := obj.CallWithContext(ctx, p.cfg.DbusQueryObject.Method, 0,
		p.cfg.DbusQueryObject.CollectArgs()...).Store(&onBattery)
	if err != nil {
		return Battery, fmt.Errorf("failed to get property from UPower: %w", err)
	}

	state := onBattery.String()

	logrus.WithFields(logrus.Fields{
		"upower_reported_state": state,
		"expected":              p.cfg.DbusQueryObject.ExpectedDischargingValue,
	}).Debug("UPower state property")
	if state == p.cfg.DbusQueryObject.ExpectedDischargingValue {
		return Battery, nil
	}
	return ACPower, nil
}

func (p *PowerDetector) Listen() <-chan PowerEvent {
	return p.events
}

func (p *PowerDetector) createMatchRules() [][]dbus.MatchOption {
	rules := [][]dbus.MatchOption{}
	for _, rule := range p.cfg.DbusSignalMatchRules {
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
	for _, filter := range p.cfg.DbusSignalReceiveFilters {
		result = append(result, *filter.Name)
	}
	return result
}

func (p *PowerDetector) Run(ctx context.Context) error {
	rules := p.createMatchRules()
	for _, ruleSet := range rules {
		if err := p.conn.AddMatchSignalContext(ctx, ruleSet...); err != nil {
			logrus.WithFields(logrus.Fields{
				"error": err,
			}).Debug("Failed to add D-Bus match rule")
			return fmt.Errorf("cant add signal rule for dbus: %w", err)
		}
	}
	p.conn.Signal(p.signals)

	eg, ctx := errgroup.WithContext(ctx)

	eg.Go(func() error {
		<-ctx.Done()
		logrus.Debug("Power detector context cancelled, closing D-Bus connection")
		_ = p.conn.Close()
		return ctx.Err()
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
					return errors.New("dbus power events channel closed")
				}
				logrus.WithFields(logrus.Fields{
					"signal_name": signal.Name,
					"signal_path": signal.Path,
				}).Debug("Received D-Bus signal")

				var currentState PowerState
				var err error

				if slices.Contains(p.getExpectedSignalNames(), signal.Name) {
					currentState, err = p.GetCurrentState(ctx)
				} else {
					logrus.WithField("signal_name", signal.Name).Debug("Ignoring unknown UPower signal")
					continue
				}

				if err != nil {
					return fmt.Errorf("failed to get power state after signal %s: %w", signal.Name, err)
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
					return ctx.Err()
				}
			case <-ctx.Done():
				logrus.Debug("Power detector context cancelled, shutting down")
				return ctx.Err()
			}
		}
	})

	if err := eg.Wait(); err != nil {
		return fmt.Errorf("goroutines for power detector failed %w", err)
	}
	return nil
}
