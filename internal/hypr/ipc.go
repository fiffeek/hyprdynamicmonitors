package hypr

import (
	"bufio"
	"context"
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
		return nil, fmt.Errorf("HYPRLAND_INSTANCE_SIGNATURE environment variable not set - are you running under Hyprland?")
	}

	xdgRuntimeDir := os.Getenv("XDG_RUNTIME_DIR")
	if signature == "" {
		return nil, fmt.Errorf("XDG_RUNTIME_DIR environment variable not set - are you running under Hyprland?")
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

	conn, err := net.Dial("unix", socketPath)
	if err != nil {
		return fmt.Errorf("failed to connect to Hyprland event socket: %w", err)
	}

	eg, ctx := errgroup.WithContext(ctx)

	eg.Go(func() error {
		<-ctx.Done()
		logrus.Debug("Hypr IPC context cancelled, closing connection to unblock scanner")
		return conn.Close()
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
			if event := h.parseMonitorEvent(line); event != nil {
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
				logrus.Debug("Hypr IPC scanner error after context cancellation, ignoring")
				return ctx.Err()
			default:
				return fmt.Errorf("scanner error: %w", err)
			}
		}

		logrus.Debug("Hypr IPC scanner finished")
		return nil
	})

	return eg.Wait()
}

func (h *IPC) extractMonitorEvent(line string, prefix string, eventType MonitorEventType) *MonitorEvent {
	after, done := strings.CutPrefix(line, prefix)
	if done {
		parts := strings.Split(after, ",")
		if len(parts) > 0 {
			return &MonitorEvent{
				Type: eventType,
				Monitor: &MonitorSpec{
					ID:          parts[0],
					Name:        parts[1],
					Description: parts[2],
				},
			}
		}
	}
	return nil
}

func (h *IPC) parseMonitorEvent(line string) *MonitorEvent {
	events := []*MonitorEvent{
		h.extractMonitorEvent(line, "monitoraddedv2>>", MonitorAdded),
		h.extractMonitorEvent(line, "monitorremovedv2>>", MonitorRemoved),
	}
	for _, event := range events {
		if event != nil {
			return event
		}
	}
	return nil
}

func (h *IPC) QueryConnectedMonitors() ([]*MonitorSpec, error) {
	socketPath := fmt.Sprintf("%s/hypr/%s/.socket.sock", h.xdgRuntimeDir, h.instanceSignature)
	if _, err := os.Stat(socketPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("hyprland command socket not found at %s", socketPath)
	}

	conn, err := net.Dial("unix", socketPath)
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
		return nil, fmt.Errorf("no monitors found in Hyprland response")
	}

	return monitors, nil
}
