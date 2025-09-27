package tui

import (
	"fmt"

	"github.com/fiffeek/hyprdynamicmonitors/internal/hypr"
)

type MonitorSpec struct {
	Name           string   `json:"name"`
	ID             *int     `json:"id"`
	Description    string   `json:"description"`
	Disabled       bool     `json:"disabled"`
	Width          int      `json:"width"`
	Height         int      `json:"height"`
	RefreshRate    float64  `json:"refreshRate"`
	Transform      int      `json:"transform"`
	Vrr            bool     `json:"vrr"`
	Scale          float64  `json:"scale"`
	X              int      `json:"x"`
	Y              int      `json:"y"`
	AvailableModes []string `json:"availableModes"`
}

func NewMonitorSpec(spec *hypr.MonitorSpec) *MonitorSpec {
	return &MonitorSpec{
		Name:           spec.Name,
		ID:             spec.ID,
		Description:    spec.Description,
		Disabled:       spec.Disabled,
		Width:          spec.Width,
		Height:         spec.Height,
		RefreshRate:    spec.RefreshRate,
		Transform:      spec.Transform,
		Vrr:            spec.Vrr,
		Scale:          spec.Scale,
		X:              spec.X,
		Y:              spec.Y,
		AvailableModes: spec.AvailableModes,
	}
}

func (m *MonitorSpec) RotationPretty() string {
	switch m.Transform {
	case 0:
		return "Rotation: 0°"
	case 1:
		return "Rotation: 90°"
	case 2:
		return "Rotation: 180°"
	case 3:
		return "Rotation: 270°"
	default:
		return fmt.Sprintf("Transform: %d", m.Transform)
	}
}

func (m *MonitorSpec) PositionPretty() string {
	return fmt.Sprintf("Position: %d,%d", m.X, m.Y)
}

func (m *MonitorSpec) VRRPretty() string {
	if m.Vrr {
		return "VRR: On"
	}
	return "VRR: Off"
}

func (m *MonitorSpec) ScalePretty() string {
	return fmt.Sprintf("Scale: %.2f", m.Scale)
}

func (m *MonitorSpec) StatusPretty() string {
	if m.Disabled {
		return "Disabled"
	}
	return "Active"
}

func (m *MonitorSpec) Mode() string {
	return fmt.Sprintf("%dx%d@%.5f",
		m.Width,
		m.Height,
		m.RefreshRate)
}

func (m *MonitorSpec) PositionArrowView() string {
	switch m.Transform {
	case 0: // Normal (0°) - top is up
		return "↑"
	case 1: // 90° clockwise - top is right
		return "→"
	case 2: // 180° - top is down
		return "↓"
	case 3: // 270° clockwise - top is left
		return "←"
	default:
		return "↑" // Default to up
	}
}

func (m *MonitorSpec) NeedsDimensionsSwap() bool {
	return m.Transform == 1 || m.Transform == 3
}

func (m *MonitorSpec) isBottomEdge(x, y, startX, startY, endX, endY int) bool {
	switch m.Transform {
	case 0: // Normal (0°) - bottom is bottom
		return (y == endY)
	case 1: // 90° clockwise - bottom is now left
		return (x == startX)
	case 2: // 180° - bottom is now top
		return (y == startY)
	case 3: // 270° clockwise - bottom is now right
		return (x == endX)
	default:
		return false
	}
}
