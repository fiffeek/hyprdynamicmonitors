// Package hypr provides Hyprland IPC communication functionality.
package hypr

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/fiffeek/hyprdynamicmonitors/internal/dial"
	"github.com/fiffeek/hyprdynamicmonitors/internal/utils"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

type IPC struct {
	instanceSignature string
	xdgRuntimeDir     string
	events            chan HyprEvent
}

func NewIPC() (*IPC, error) {
	signature := os.Getenv("HYPRLAND_INSTANCE_SIGNATURE")
	if signature == "" {
		return nil, errors.New("HYPRLAND_INSTANCE_SIGNATURE environment variable not set - are you running under Hyprland?")
	}

	xdgRuntimeDir, err := utils.GetXDGRuntimeDir()
	if err != nil {
		return nil, fmt.Errorf("cant get xdg runtime dir: %w", err)
	}

	return &IPC{
		instanceSignature: signature,
		xdgRuntimeDir:     xdgRuntimeDir,
		events:            make(chan HyprEvent, 10),
	}, nil
}

func (h *IPC) ListenEvents() <-chan HyprEvent {
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
		return nil
	})

	eg.Go(func() error {
		defer close(h.events)
		defer func() {
			if err := conn.Close(); err != nil {
				logrus.WithError(err).Debug("Failed to close connection")
			}
		}()

		scanner := bufio.NewScanner(conn)
		for scanner.Scan() {
			select {
			case <-ctx.Done():
				logrus.Debug("Hypr IPC context cancelled during processing")
				return ctx.Err()
			default:
			}

			line := scanner.Text()
			event, err := extractHyprEvent(line)
			if err != nil {
				return fmt.Errorf("scanner gave unknown event: %w", err)
			}
			if event != nil {
				select {
				case h.events <- *event:
				case <-ctx.Done():
					logrus.Debug("Hypr IPC context cancelled during event send")
					return ctx.Err()
				}
			}
		}

		if err := scanner.Err(); err != nil {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
				return fmt.Errorf("scanner error: %w", err)
			}
		}

		logrus.Debug("Hypr IPC scanner finished")
		return nil
	})

	if err = eg.Wait(); err != nil {
		return fmt.Errorf("goroutines for hypr ipc failed %w", err)
	}
	return nil
}

func (h *IPC) QueryConnectedMonitors(ctx context.Context) (MonitorSpecs, error) {
	socketPath := GetHyprSocket(h.xdgRuntimeDir, h.instanceSignature)
	conn, teardown, err := dial.GetUnixSocketConnection(ctx, socketPath)
	defer teardown()
	if err != nil {
		return nil, fmt.Errorf("cant open socket to %s: %w", socketPath, err)
	}

	return dial.SyncQuerySocket[MonitorSpecs](conn, "j/monitors")
}
