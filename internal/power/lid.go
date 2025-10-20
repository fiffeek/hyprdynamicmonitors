package power

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"sync"

	"github.com/fiffeek/hyprdynamicmonitors/internal/config"
	"github.com/fiffeek/hyprdynamicmonitors/internal/errs"
	"github.com/fiffeek/hyprdynamicmonitors/internal/utils"
	"github.com/godbus/dbus/v5"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

type LidState int

const (
	UnknownLidState LidState = iota
	OpenedLidState
	ClosedLidState
)

func (p LidState) String() string {
	switch p {
	case OpenedLidState:
		return "Opened"
	case ClosedLidState:
		return "Closed"
	default:
		return "UNKNOWN"
	}
}

type LidEvent struct {
	State LidState
}

type LidStateDetector struct {
	enableLidEvents bool
	cfg             *config.Config

	conn             *dbus.Conn
	events           chan LidEvent
	signals          chan *dbus.Signal
	stateMu          sync.RWMutex
	dbusMatchOptions [][]dbus.MatchOption

	lidState   LidState
	lidStateMu sync.RWMutex
}

func NewLidStateDetector(ctx context.Context, cfg *config.Config, conn *dbus.Conn, enableLidEvents bool) (*LidStateDetector, error) {
	detector := &LidStateDetector{
		conn:            conn,
		cfg:             cfg,
		events:          make(chan LidEvent, 10),
		signals:         make(chan *dbus.Signal, 10),
		enableLidEvents: enableLidEvents,
		lidStateMu:      sync.RWMutex{},
		stateMu:         sync.RWMutex{},
	}

	ps, err := detector.getCurrentState(ctx)
	if err != nil {
		_ = conn.Close()
		//nolint:errorlint
		return nil, fmt.Errorf("%w: %v", errs.ErrUPowerMisconfigured, err)
	}
	detector.lidState = ps

	logrus.Info("UPower D-Bus lid detection initialized")

	return detector, nil
}

func (l *LidStateDetector) GetCurrentState() LidState {
	l.lidStateMu.RLock()
	defer l.lidStateMu.RUnlock()
	return l.lidState
}

func (l *LidStateDetector) Listen() <-chan LidEvent {
	return l.events
}

func (l *LidStateDetector) shouldHandleSignal(sig *dbus.Signal) bool {
	for _, filter := range l.cfg.Get().LidEvents.DbusSignalReceiveFilters {
		if filter.Name != nil && *filter.Name != sig.Name {
			logrus.WithFields(logrus.Fields{"filter": *filter.Name}).Debug("Filter not matching the name")
			continue
		}
		sigBody := utils.SignalBodyToString(sig.Body)
		if filter.Body != nil && !strings.Contains(sigBody, *filter.Body) {
			logrus.WithFields(logrus.Fields{"filter": *filter.Body}).Debug("Filter not matching the body")
			continue
		}
		logrus.Debug("Filter matching the signal")
		return true
	}
	return false
}

func (l *LidStateDetector) createMatchRules() [][]dbus.MatchOption {
	rules := [][]dbus.MatchOption{}
	for _, rule := range l.cfg.Get().LidEvents.DbusSignalMatchRules {
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

func (l *LidStateDetector) Reload(ctx context.Context) error {
	if !l.enableLidEvents {
		logrus.Debug("Lid events are disabled, not reloading power rules")
		return nil
	}

	l.stateMu.Lock()
	defer l.stateMu.Unlock()
	rules := l.createMatchRules()
	if reflect.DeepEqual(rules, l.dbusMatchOptions) {
		logrus.Debug("power events rules match, nothing to be done")
	}

	for _, ruleSet := range l.dbusMatchOptions {
		if err := l.conn.RemoveMatchSignalContext(ctx, ruleSet...); err != nil {
			logrus.WithFields(logrus.Fields{
				"error": err,
			}).Debug("Failed to remove D-Bus match rule")
			return fmt.Errorf("cant remove signal rule for dbus: %w", err)
		}
	}

	for _, ruleSet := range rules {
		if err := l.conn.AddMatchSignalContext(ctx, ruleSet...); err != nil {
			logrus.WithFields(logrus.Fields{
				"error": err,
			}).Debug("Failed to add D-Bus match rule")
			return fmt.Errorf("cant add signal rule for dbus: %w", err)
		}
	}

	l.dbusMatchOptions = rules
	logrus.Debug("Reloaded power detector")
	return nil
}

func (l *LidStateDetector) getCurrentState(ctx context.Context) (LidState, error) {
	cfg := l.cfg.Get().LidEvents
	if !l.enableLidEvents {
		logrus.WithFields(logrus.Fields{"default": OpenedLidState.String()}).Debug(
			"Lid events are disabled, returning the default value")
		return OpenedLidState, nil
	}

	obj := l.conn.Object(
		cfg.DbusQueryObject.Destination,
		dbus.ObjectPath(cfg.DbusQueryObject.Path))

	logrus.WithFields(logrus.Fields{
		"destination": cfg.DbusQueryObject.Destination,
		"path":        cfg.DbusQueryObject.Path,
		"method":      cfg.DbusQueryObject.Method,
		"args":        cfg.DbusQueryObject.CollectArgs(),
	}).Debug("About to make D-Bus method call")

	var lidClosed dbus.Variant
	err := obj.CallWithContext(ctx, cfg.DbusQueryObject.Method, 0,
		cfg.DbusQueryObject.CollectArgs()...).Store(&lidClosed)
	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"destination": cfg.DbusQueryObject.Destination,
			"path":        cfg.DbusQueryObject.Path,
			"method":      cfg.DbusQueryObject.Method,
		}).Error("D-Bus method call failed")
		return UnknownLidState, fmt.Errorf("failed to get property from UPower: %w", err)
	}

	logrus.Debug("D-Bus method call succeeded")

	state := lidClosed.String()

	logrus.WithFields(logrus.Fields{
		"upower_reported_state": state,
		"expected":              cfg.DbusQueryObject.ExpectedLidClosingValue,
	}).Debug("UPower state property")
	if state == cfg.DbusQueryObject.ExpectedLidClosingValue {
		return ClosedLidState, nil
	}
	return OpenedLidState, nil
}

func (l *LidStateDetector) Run(ctx context.Context) error {
	eg, ctx := errgroup.WithContext(ctx)

	eg.Go(func() error {
		<-ctx.Done()
		logrus.Debug("Lid detector context cancelled, closing D-Bus connection")
		if l.conn != nil {
			_ = l.conn.Close()
		}
		return context.Cause(ctx)
	})

	if !l.enableLidEvents {
		logrus.Info("Lid events are disabled, waiting for ctx cancellation")
		return eg.Wait()
	}

	if err := l.Reload(ctx); err != nil {
		return fmt.Errorf("cant reload: %w", err)
	}
	l.conn.Signal(l.signals)

	eg.Go(func() error {
		defer close(l.events)
		defer l.conn.RemoveSignal(l.signals)

		logrus.Debug("Lid detector started, listening for UPower D-Bus signals")

		lastState := UnknownLidState

		for {
			select {
			case signal, ok := <-l.signals:
				if !ok {
					return errors.New("dbus lid events channel closed")
				}
				logrus.WithFields(logrus.Fields{
					"signal_name": signal.Name,
					"signal_path": signal.Path,
					"signal_body": signal.Body,
				}).Debug("Received D-Bus signal")

				var currentState LidState
				var err error

				if !l.shouldHandleSignal(signal) {
					logrus.WithField("signal_name", signal.Name).Debug("Ignoring unknown UPower signal")
					continue
				}

				logrus.Debug("Signal, calling getCurrentState")
				currentState, err = l.getCurrentState(ctx)
				if err != nil {
					logrus.WithError(err).Error("getCurrentState failed")
					return fmt.Errorf("failed to get power state after signal %s: %w", signal.Name, err)
				}

				logrus.WithFields(logrus.Fields{"state": currentState}).Debug("getCurrentState succeeded")
				l.lidStateMu.Lock()
				l.lidState = currentState
				l.lidStateMu.Unlock()

				if currentState == lastState {
					logrus.WithField("power_state", currentState.String()).Debug("Lid state unchanged after signal")
					continue
				}

				logrus.WithFields(logrus.Fields{
					"from": lastState.String(),
					"to":   currentState.String(),
				}).Info("Lid state changed")

				select {
				case l.events <- LidEvent{State: currentState}:
					lastState = currentState
				case <-ctx.Done():
					return context.Cause(ctx)
				}
			case <-ctx.Done():
				logrus.Debug("Lid detector context cancelled, shutting down")
				return context.Cause(ctx)
			}
		}
	})

	if err := eg.Wait(); err != nil {
		return fmt.Errorf("goroutines for lid detector failed %w", err)
	}
	return nil
}
