package hypr

import (
	"errors"
	"fmt"
)

type MonitorSpec struct {
	Name        string `json:"name"`
	ID          *int   `json:"id"`
	Description string `json:"description"`
}

func (m *MonitorSpec) Validate() error {
	if m.ID == nil || m.Description == "" || m.Name == "" {
		return errors.New("monitor spec is invalid")
	}

	return nil
}

type MonitorSpecs []*MonitorSpec

func (m MonitorSpecs) Validate() error {
	if len(m) == 0 {
		return errors.New("no monitors")
	}

	for _, monitor := range m {
		if err := monitor.Validate(); err != nil {
			return fmt.Errorf("monitor spec is invalid: %w", err)
		}
	}

	return nil
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
