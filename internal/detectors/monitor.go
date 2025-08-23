package detectors

import (
	"fmt"
	"log"
	"sync"

	"github.com/fiffeek/hyprdynamicmonitors/internal/hypr"
)

type MonitorDetector struct {
	ipc               *hypr.IPC
	connectedMonitors map[string]*hypr.MonitorSpec
	monitorsMutex     sync.RWMutex
}

func NewMonitorDetector(ipc *hypr.IPC) (*MonitorDetector, error) {
	monitors, err := ipc.QueryConnectedMonitors()
	if err != nil {
		return nil, err
	}
	log.Printf(" mondet %d", len(monitors))
	connectedMonitors := make(map[string]*hypr.MonitorSpec)
	for _, monitor := range monitors {
		connectedMonitors[monitor.Name] = monitor
	}
	return &MonitorDetector{ipc, connectedMonitors, sync.RWMutex{}}, nil
}

func (m *MonitorDetector) Listen() (<-chan []*hypr.MonitorSpec, error) {
	monitors := make(chan []*hypr.MonitorSpec, 1)
	events, err := m.ipc.ListenEvents()
	if err != nil {
		return nil, fmt.Errorf("failed to start event listener: %w", err)
	}

	go func() {
		defer close(monitors)
		for event := range events {
			m.monitorsMutex.Lock()
			switch event.Type {
			case hypr.MonitorAdded:
				m.connectedMonitors[event.Monitor.Name] = event.Monitor
			case hypr.MonitorRemoved:
				delete(m.connectedMonitors, event.Monitor.Name)
			}
			m.monitorsMutex.Unlock()
			monitors <- m.GetConnected()
		}
	}()

	return monitors, nil
}

func (m *MonitorDetector) GetConnected() []*hypr.MonitorSpec {
	m.monitorsMutex.Lock()
	defer m.monitorsMutex.Unlock()
	var monitors []*hypr.MonitorSpec
	for _, monitor := range m.connectedMonitors {
		monitors = append(monitors, monitor)
	}
	return monitors
}
