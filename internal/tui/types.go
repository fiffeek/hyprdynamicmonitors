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

func (m *MonitorSpec) ToHypr() string {
	if m.Disabled {
		return fmt.Sprintf("desc:%s,disable", m.Description)
	}
	return fmt.Sprintf("desc:%s,%dx%d@%.5f,%dx%d,%.2f,transform,%d", m.Description, m.Width, m.Height, m.RefreshRate, m.X, m.Y, m.Scale, m.Transform)
}

type MonitorRectangle struct {
	startX  int
	startY  int
	endX    int
	endY    int
	monitor *MonitorSpec
}

func NewMonitorRectangle(startX, startY, endX, endY int, monitor *MonitorSpec) *MonitorRectangle {
	rec := &MonitorRectangle{
		startX:  startX,
		startY:  startY,
		endX:    endX,
		endY:    endY,
		monitor: monitor,
	}

	if rec.endX <= rec.startX {
		rec.endX = rec.startX + 4
	}
	if rec.endY <= rec.startY {
		rec.endY = rec.startY + 2
	}

	return rec
}

// Clamp to grid bounds
func (m *MonitorRectangle) Clamp(gridWidth, gridHeight int) {
	if m.startX < 0 {
		m.startX = 0
	}
	if m.startY < 0 {
		m.startY = 0
	}
	if m.endX >= gridWidth {
		m.endX = gridWidth - 1
	}
	if m.endY >= gridHeight {
		m.endY = gridHeight - 1
	}
}

func (m *MonitorRectangle) isBottomEdge(x, y int) bool {
	switch m.monitor.Transform {
	case 0: // Normal (0°) - bottom is bottom
		return (y == m.endY)
	case 1: // 90° clockwise - bottom is now left
		return (x == m.startX)
	case 2: // 180° - bottom is now top
		return (y == m.startY)
	case 3: // 270° clockwise - bottom is now right
		return (x == m.endX)
	default:
		return false
	}
}
