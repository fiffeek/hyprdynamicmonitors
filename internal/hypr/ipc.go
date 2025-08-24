// Package hypr provides Hyprland IPC communication functionality.
package hypr

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

type IPC struct {
	instanceSignature string
	xdgRuntimeDir     string
	events            chan MonitorEvent
}

type MonitorSpec struct {
	Name        string
	ID          string
	Description string
}

type MonitorEventType int

const (
	MonitorAdded MonitorEventType = iota
	MonitorRemoved
)

type MonitorEvent struct {
	Type    MonitorEventType
	Monitor *MonitorSpec
}

func NewIPC() (*IPC, error) {
	signature := os.Getenv("HYPRLAND_INSTANCE_SIGNATURE")
	if signature == "" {
		return nil, errors.New("HYPRLAND_INSTANCE_SIGNATURE environment variable not set - are you running under Hyprland?")
	}

	xdgRuntimeDir := os.Getenv("XDG_RUNTIME_DIR")
	if xdgRuntimeDir == "" {
		return nil, errors.New("XDG_RUNTIME_DIR environment variable not set - are you running under Hyprland?")
	}

	ipc := &IPC{
		instanceSignature: signature,
		xdgRuntimeDir:     xdgRuntimeDir,
		events:            make(chan MonitorEvent, 10),
	}

	return ipc, nil
}

func (h *IPC) ListenEvents() <-chan MonitorEvent {
	return h.events
}

func (h *IPC) Run(ctx context.Context) error {
	socketPath := fmt.Sprintf("%s/hypr/%s/.socket2.sock", h.xdgRuntimeDir, h.instanceSignature)

	if _, err := os.Stat(socketPath); os.IsNotExist(err) {
		return fmt.Errorf("hyprland event socket not found at %s", socketPath)
	}

	d := &net.Dialer{}
	conn, err := d.DialContext(ctx, "unix", socketPath)
	if err != nil {
		return fmt.Errorf("failed to connect to Hyprland event socket: %w", err)
	}

	eg, ctx := errgroup.WithContext(ctx)

	eg.Go(func() error {
		<-ctx.Done()
		logrus.Debug("Hypr IPC context cancelled, closing connection to unblock scanner")
		_ = conn.Close()
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
			event, err := h.parseMonitorEvent(line)
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

func (h *IPC) extractMonitorEvent(line, prefix string, eventType MonitorEventType) (*MonitorEvent, error) {
	after, done := strings.CutPrefix(line, prefix)
	if !done {
		return nil, nil
	}

	logrus.WithFields(logrus.Fields{"event": line, "as": prefix}).Debug("trying to parse event")

	parts := strings.Split(after, ",")
	if len(parts) != 3 {
		return nil, fmt.Errorf("cant parse event %s", after)
	}

	return &MonitorEvent{
		Type: eventType,
		Monitor: &MonitorSpec{
			ID:          parts[0],
			Name:        parts[1],
			Description: parts[2],
		},
	}, nil
}

func (h *IPC) parseMonitorEvent(line string) (*MonitorEvent, error) {
	events := []func(string) (*MonitorEvent, error){
		func(line string) (*MonitorEvent, error) {
			return h.extractMonitorEvent(line, "monitoraddedv2>>", MonitorAdded)
		},
		func(line string) (*MonitorEvent, error) {
			return h.extractMonitorEvent(line, "monitorremovedv2>>", MonitorRemoved)
		},
	}
	for _, fun := range events {
		event, err := fun(line)
		if err != nil {
			return nil, fmt.Errorf("cant parse: %w", err)
		}
		if event != nil {
			return event, nil
		}
	}
	return nil, nil
}

func (h *IPC) QueryConnectedMonitors(ctx context.Context) ([]*MonitorSpec, error) {
	socketPath := fmt.Sprintf("%s/hypr/%s/.socket.sock", h.xdgRuntimeDir, h.instanceSignature)
	if _, err := os.Stat(socketPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("hyprland command socket not found at %s", socketPath)
	}

	d := &net.Dialer{}
	conn, err := d.DialContext(ctx, "unix", socketPath)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Hyprland command socket: %w", err)
	}

	defer func() {
		if err := conn.Close(); err != nil {
			logrus.WithError(err).Debug("Failed to close connection")
		}
	}()

	_, err = conn.Write([]byte("monitors"))
	if err != nil {
		return nil, fmt.Errorf("failed to send monitors command: %w", err)
	}

	response, err := io.ReadAll(conn)
	if err != nil {
		return nil, fmt.Errorf("failed to read monitors response: %w", err)
	}

	return h.parseMonitorsResponse(string(response))
}

func (h *IPC) parseMonitorsResponse(response string) ([]*MonitorSpec, error) {
	var monitors []*MonitorSpec
	logrus.WithField("response", response).Debug("Hyprland monitors response")
	monitorBlocks := strings.Split(response, "Monitor")

	for _, block := range monitorBlocks {
		id := ""
		desc := ""
		name := ""

		for lineNumber, line := range strings.Split(block, "\n") {
			line = strings.TrimSpace(line)
			if lineNumber == 0 {
				logrus.WithField("line", line).Debug("Monitor header line")
				parts := strings.Fields(line)
				logrus.WithField("parts_count", len(parts)).Debug("Monitor header parts")
				if len(parts) != 3 {
					continue
				}

				name = parts[0]
				id = strings.Trim(parts[2], ":()")
			}

			descriptionSep := "description: "
			if strings.Contains(line, descriptionSep) {
				parts := strings.Split(line, descriptionSep)
				desc = parts[1]
			}
		}

		logrus.WithFields(logrus.Fields{
			"name": name,
			"desc": desc,
			"id":   id,
		}).Debug("Monitor parsed")

		if id == "" || desc == "" || name == "" {
			continue
		}
		monitors = append(monitors, &MonitorSpec{
			name,
			id,
			desc,
		})
	}

	if len(monitors) == 0 {
		return nil, errors.New("no monitors found in Hyprland response")
	}

	return monitors, nil
}
