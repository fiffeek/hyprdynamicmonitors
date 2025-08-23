package hypr

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strings"
)

type IPC struct {
	instanceSignature string
	xdgRuntimeDir     string
	verbose           bool
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

func NewIPC(verbose bool) (*IPC, error) {
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
		verbose:           true,
	}

	return ipc, nil
}

func (h *IPC) ListenEvents() (<-chan MonitorEvent, error) {
	socketPath := fmt.Sprintf("%s/hypr/%s/.socket2.sock", h.xdgRuntimeDir, h.instanceSignature)

	if _, err := os.Stat(socketPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("hyprland event socket not found at %s", socketPath)
	}

	conn, err := net.Dial("unix", socketPath)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Hyprland event socket: %w", err)
	}

	events := make(chan MonitorEvent, 10)

	go func() {
		defer close(events)
		defer func() {
			if err := conn.Close(); err != nil {
				log.Printf("Failed to close connection: %v", err)
			}
		}()

		for {
			scanner := bufio.NewScanner(conn)
			for scanner.Scan() {
				line := scanner.Text()

				if event := h.parseMonitorEvent(line); event != nil {
					events <- *event
				}
			}

			if err := scanner.Err(); err != nil {
				log.Printf("scanner error %v", err)
			}
		}
	}()

	return events, nil
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
			log.Printf("Failed to close connection: %v", err)
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
	if h.verbose {
		log.Printf("Read: %s,", response)
	}
	monitorBlocks := strings.Split(response, "Monitor")

	for _, block := range monitorBlocks {
		id := ""
		desc := ""
		name := ""

		for lineNumber, line := range strings.Split(block, "\n") {
			line = strings.TrimSpace(line)
			if lineNumber == 0 {
				log.Printf("%s", line)
				parts := strings.Fields(line)
				log.Printf("%d", len(parts))
				if len(parts) != 3 {
					continue
				}

				name = parts[0]
				id = strings.Trim(parts[2], ":()")
			}
			log.Printf("%s %s %s", name, desc, id)

			descriptionSep := "description: "
			if strings.Contains(line, descriptionSep) {
				parts := strings.Split(line, descriptionSep)
				desc = parts[1]
			}
		}

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
