// Package detectors provides monitor and power state detection functionality.
package detectors

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/fiffeek/hyprdynamicmonitors/internal/hypr"
	"golang.org/x/sync/errgroup"
)

type MonitorDetector struct {
	ipc               *hypr.IPC
	connectedMonitors map[string]*hypr.MonitorSpec
	monitorsMutex     sync.RWMutex
	monitors          chan []*hypr.MonitorSpec
}

func NewMonitorDetector(ctx context.Context, ipc *hypr.IPC) (*MonitorDetector, error) {
	monitors, err := ipc.QueryConnectedMonitors(ctx)
	if err != nil {
		return nil, fmt.Errorf("cant query connected monitors on start up: %w", err)
	}
	connectedMonitors := make(map[string]*hypr.MonitorSpec)
	for _, monitor := range monitors {
		connectedMonitors[monitor.Name] = monitor
	}
	return &MonitorDetector{
		ipc:               ipc,
		connectedMonitors: connectedMonitors,
		monitorsMutex:     sync.RWMutex{},
		monitors:          make(chan []*hypr.MonitorSpec, 1),
	}, nil
}

func (m *MonitorDetector) Listen() <-chan []*hypr.MonitorSpec {
	return m.monitors
}

func (m *MonitorDetector) Run(ctx context.Context) error {
	events := m.ipc.ListenEvents()

	eg, ctx := errgroup.WithContext(ctx)

	eg.Go(func() error {
		defer close(m.monitors)
		for {
			select {
			case event, ok := <-events:
				if !ok {
					return errors.New("monitor ipc events channel closed")
				}
				m.monitorsMutex.Lock()
				switch event.Type {
				case hypr.MonitorAdded:
					m.connectedMonitors[event.Monitor.Name] = event.Monitor
				case hypr.MonitorRemoved:
					delete(m.connectedMonitors, event.Monitor.Name)
				}
				connectedMonitors := m.GetConnectedUnsafe()
				m.monitorsMutex.Unlock()

				select {
				case m.monitors <- connectedMonitors:
				case <-ctx.Done():
					return ctx.Err()
				}
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	})

	if err := eg.Wait(); err != nil {
		return fmt.Errorf("goroutines for monitor detector failed %w", err)
	}
	return nil
}

func (m *MonitorDetector) GetConnected() []*hypr.MonitorSpec {
	m.monitorsMutex.Lock()
	defer m.monitorsMutex.Unlock()
	return m.GetConnectedUnsafe()
}

func (m *MonitorDetector) GetConnectedUnsafe() []*hypr.MonitorSpec {
	var monitors []*hypr.MonitorSpec
	for _, monitor := range m.connectedMonitors {
		monitors = append(monitors, monitor)
	}
	return monitors
}
