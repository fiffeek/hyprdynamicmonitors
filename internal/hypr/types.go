package hypr

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"
)

type MonitorSpec struct {
	Name            string   `json:"name"`
	ID              *int     `json:"id"`
	Description     string   `json:"description"`
	Disabled        bool     `json:"disabled"`
	Width           int      `json:"width"`
	Height          int      `json:"height"`
	RefreshRate     float64  `json:"refreshRate"`
	Transform       int      `json:"transform"`
	Vrr             bool     `json:"vrr"`
	Scale           float64  `json:"scale"`
	X               int      `json:"x"`
	Y               int      `json:"y"`
	AvailableModes  []string `json:"availableModes"`
	Mirror          string   `json:"mirrorOf"`
	CurrentFormat   string   `json:"currentFormat"`
	DpmsStatus      bool     `json:"dpmsStatus"`
	ActivelyTearing bool     `json:"activelyTearing"`
	DirectScanoutTo string   `json:"directScanoutTo"`
	Solitary        string   `json:"solitary"`
	TenBitdepth     bool     `json:"-"`
	// TODO(fmikina, 14.10.25): fix when hyprctl supports these
	SdrBrightness float64 `json:"-"`
	SdrSaturation float64 `json:"-"`
	ColorPreset   string  `json:"-"`
}

func (m *MonitorSpec) IsTenBitdepth() bool {
	switch m.CurrentFormat {
	case "XRGB2101010":
		return true
	case "XBGR2101010":
		return true
	}
	return false
}

func (m *MonitorSpec) HDR() bool {
	return m.ColorPreset == "hdr"
}

func (m *MonitorSpec) HasNonDefaultColorPreset() bool {
	return m.ColorPreset != "" && m.ColorPreset != "srgb" && m.ColorPreset != "auto"
}

func (m *MonitorSpec) HasMirror() bool {
	return m.Mirror != "none" && m.Mirror != ""
}

func (m *MonitorSpec) Validate() error {
	if m.ID == nil {
		return errors.New("id cant be nil")
	}
	if *m.ID < 0 {
		return errors.New("id cant < 0")
	}
	if m.Description == "" {
		return errors.New("desc cant be empty")
	}
	if m.Name == "" {
		return errors.New("name cant be empty")
	}
	if !m.TenBitdepth {
		m.TenBitdepth = m.IsTenBitdepth()
	}
	if m.SdrBrightness == 0.0 {
		m.SdrBrightness = 1.0
	}
	if m.SdrSaturation == 0.0 {
		m.SdrSaturation = 1.0
	}

	return nil
}

type MonitorSpecs []*MonitorSpec

func (m MonitorSpecs) Validate() error {
	if len(m) == 0 {
		return errors.New("no monitors detected")
	}

	for _, monitor := range m {
		if err := monitor.Validate(); err != nil {
			return fmt.Errorf("invalid monitor: %w", err)
		}
	}

	return nil
}

type HyprEventType int

const (
	MonitorUnknown HyprEventType = iota
	MonitorAdded
	MonitorRemoved
)

func (m HyprEventType) Value() string {
	switch m {
	case MonitorAdded:
		return "monitoraddedv2>>"
	case MonitorRemoved:
		return "monitorremovedv2>>"
	}
	return "unknownevent>>"
}

type HyprEvent struct {
	Type    HyprEventType
	Monitor *MonitorSpec
}

func (m HyprEvent) Validate() error {
	switch m.Type {
	case MonitorAdded:
	case MonitorRemoved:
		if m.Monitor == nil {
			return errors.New("hypr event monitor type does not have the monitor description")
		}
	}

	return nil
}

func extractTypedHyprEvent(line string, eventType HyprEventType) (bool, *HyprEvent, error) {
	after, done := strings.CutPrefix(line, eventType.Value())
	if !done {
		return done, nil, nil
	}

	logrus.WithFields(logrus.Fields{"event": line, "as": eventType.Value()}).Debug("trying to parse event")

	parts := strings.Split(after, ",")
	if len(parts) != 3 {
		return done, nil, fmt.Errorf("cant parse event %s", after)
	}

	id, err := strconv.Atoi(parts[0])
	if err != nil {
		return done, nil, fmt.Errorf("cant parse %s as int: %w", parts[0], err)
	}

	return done, &HyprEvent{
		Type: eventType,
		Monitor: &MonitorSpec{
			ID:          &id,
			Name:        parts[1],
			Description: parts[2],
		},
	}, nil
}

func extractHyprEvent(line string) (bool, *HyprEvent, error) {
	possibleEvents := []HyprEventType{MonitorAdded, MonitorRemoved}
	for _, event := range possibleEvents {
		ok, parsedEvent, err := extractTypedHyprEvent(line, event)
		if !ok {
			continue
		}
		if err != nil {
			return false, nil, fmt.Errorf("cant parse event %s as %s: %w", line, event.Value(), err)
		}
		if err := parsedEvent.Validate(); err != nil {
			return false, nil, fmt.Errorf("invalid hypr event: %w", err)
		}
		return true, parsedEvent, nil
	}
	return false, nil, nil
}
