// Package hypr provides Hyprland IPC communication functionality.
package hypr

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"reflect"
	"slices"
	"sync"

	"github.com/fiffeek/hyprdynamicmonitors/internal/dial"
	"github.com/fiffeek/hyprdynamicmonitors/internal/utils"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

type IPC struct {
	instanceSignature string
	xdgRuntimeDir     string
	events            chan MonitorSpecs
	monitors          MonitorSpecs
	mu                sync.RWMutex
}

func NewIPC(ctx context.Context) (*IPC, error) {
	signature := os.Getenv("HYPRLAND_INSTANCE_SIGNATURE")
	if signature == "" {
		return nil, errors.New("HYPRLAND_INSTANCE_SIGNATURE environment variable not set - are you running under Hyprland?")
	}

	xdgRuntimeDir, err := utils.GetXDGRuntimeDir()
	if err != nil {
		return nil, fmt.Errorf("cant get xdg runtime dir: %w", err)
	}

	ipc := &IPC{
		instanceSignature: signature,
		xdgRuntimeDir:     xdgRuntimeDir,
		events:            make(chan MonitorSpecs, 10),
		mu:                sync.RWMutex{},
	}

	monitors, err := ipc.queryConnectedMonitors(ctx)
	if err != nil {
		return nil, fmt.Errorf("cant query current monitors: %w", err)
	}
	ipc.monitors = monitors

	return ipc, nil
}

// Listen sends ALL present (not ENABLED) monitors in each message
func (h *IPC) Listen() <-chan MonitorSpecs {
	return h.events
}

func (h *IPC) RunEventLoop(ctx context.Context) error {
	socketPath := GetHyprEventsSocket(h.xdgRuntimeDir, h.instanceSignature)
	eg, ctx := errgroup.WithContext(ctx)

	conn, connTeardown, err := dial.GetUnixSocketConnection(ctx, socketPath)
	if err != nil {
		return fmt.Errorf("cant open unix events socket connection to %s: %w", socketPath, err)
	}

	eg.Go(func() error {
		<-ctx.Done()
		logrus.Debug("Hypr IPC context cancelled, closing connection to unblock scanner")
		connTeardown()
		return context.Cause(ctx)
	})

	eg.Go(func() error {
		defer close(h.events)
		defer connTeardown()

		lastSent := MonitorSpecs{}

		scanner := bufio.NewScanner(conn)
		for scanner.Scan() {
			select {
			case <-ctx.Done():
				logrus.Debug("Hypr IPC context cancelled during processing")
				return context.Cause(ctx)
			default:
			}

			line := scanner.Text()
			found, event, err := extractHyprEvent(line)
			if err != nil {
				return fmt.Errorf("scanner gave unknown event: %w", err)
			}
			// if event was not found, skip
			if !found {
				continue
			}
			// skip updates for non monitor related events
			if !slices.Contains([]HyprEventType{MonitorAdded, MonitorRemoved}, event.Type) {
				continue
			}

			// refetch the current state, it's unknown whether monitorremoved
			// event means disabled or physically removed
			monitors, err := h.queryConnectedMonitors(ctx)

			h.mu.Lock()
			h.monitors = monitors
			h.mu.Unlock()

			if err != nil {
				return fmt.Errorf("cant get monitors spec: %w", err)
			}
			if reflect.DeepEqual(monitors, lastSent) {
				logrus.Debug("Monitors are unchanged between now and the last sent value, skipping send")
				continue
			}

			logrus.WithFields(logrus.Fields{"monitors": len(monitors)}).Info("Monitors state changed")

			select {
			case h.events <- monitors:
				lastSent = monitors
				logrus.Debug("Monitors event sent")
			case <-ctx.Done():
				logrus.Debug("Hypr IPC context cancelled during event send")
				return context.Cause(ctx)
			}
		}

		if err := scanner.Err(); err != nil {
			return fmt.Errorf("scanner error: %w", err)
		}

		logrus.Debug("Hypr IPC scanner finished")
		return nil
	})

	if err = eg.Wait(); err != nil {
		return fmt.Errorf("goroutines for hypr ipc failed %w", err)
	}
	return nil
}

func (h *IPC) GetConnectedMonitors() MonitorSpecs {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.monitors
}

func (h *IPC) queryConnectedMonitors(ctx context.Context) (MonitorSpecs, error) {
	logrus.Debug("Querying connected monitors")
	socketPath := GetHyprSocket(h.xdgRuntimeDir, h.instanceSignature)
	conn, teardown, err := dial.GetUnixSocketConnection(ctx, socketPath)
	defer teardown()
	if err != nil {
		return nil, fmt.Errorf("cant open socket to %s: %w", socketPath, err)
	}

	return dial.SyncQuerySocket[MonitorSpecs](conn, "j/monitors all\n")
}
