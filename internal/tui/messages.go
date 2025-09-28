package tui

import (
	tea "github.com/charmbracelet/bubbletea"
)

type MonitorBeingEdited struct {
	ListIndex     int
	Scaling       bool
	MonitorID     int
	ModesEditor   bool
	MirroringMode bool
}

type MonitorUnselected struct{}

type StateChanged struct {
	state AppState
}

type OperationName int

const (
	OperationNameNone = iota
	OperationNameScale
	OperationNameRotate
	OperationNameToggleVRR
	OperationNameToggleMonitor
	OperationNameMove
	OperationNamePreviewMode
	OperationNamePreviewMirror
)

type OperationStatus struct {
	name OperationName
	err  error
}

func (o OperationStatus) IsError() bool {
	return o.err != nil
}

func (o OperationStatus) String() string {
	operationName := ""

	switch o.name {
	case OperationNameNone:
		operationName = "None"
	case OperationNameScale:
		operationName = "Scale"
	case OperationNameRotate:
		operationName = "Rotate"
	case OperationNameToggleVRR:
		operationName = "Toggle VRR"
	case OperationNameToggleMonitor:
		operationName = "Toggle Monitor"
	case OperationNameMove:
		operationName = "Move"
	case OperationNamePreviewMode:
		operationName = "Preview Mode"
	case OperationNamePreviewMirror:
		operationName = "Preview Mirror"
	default:
		operationName = "Operation"
	}

	result := operationName
	if o.err != nil {
		result += ": " + o.err.Error()
	} else {
		result += ": success"
	}

	if len(result) > 50 {
		result = result[:50]
	}

	return result
}

func operationStatusCmd(name OperationName, err error) tea.Cmd {
	return func() tea.Msg {
		return OperationStatus{
			name: name,
			err:  err,
		}
	}
}

type Delta int

const (
	DeltaNone Delta = iota
	DeltaMore
	DeltaLess
)

type ScaleMonitorCommand struct {
	monitorID int
	delta     Delta
}

type RotateMonitorCommand struct {
	monitorID int
}

type MoveMonitorCommand struct {
	monitorID int
	stepX     Delta
	stepY     Delta
}

type ToggleMonitorVRRCommand struct {
	monitorID int
}

type ToggleMonitorCommand struct {
	monitorID int
}

type ChangeModePreviewCommand struct {
	mode string
}

type ChangeMirrorPreviewCommand struct {
	mirrorOf string
}

type ChangeMirrorCommand struct {
	mirrorOf string
}

type ChangeModeCommand struct {
	mode string
}

func changeMirrorPreviewCmd(mirror string) tea.Cmd {
	return func() tea.Msg {
		return ChangeMirrorPreviewCommand{
			mirrorOf: mirror,
		}
	}
}

func changeMirrorCmd(mirror string) tea.Cmd {
	return func() tea.Msg {
		return ChangeMirrorCommand{
			mirrorOf: mirror,
		}
	}
}

func changeModeCmd(mode string) tea.Cmd {
	return func() tea.Msg {
		return ChangeModeCommand{
			mode: mode,
		}
	}
}

func changeModePreviewCmd(mode string) tea.Cmd {
	return func() tea.Msg {
		return ChangeModePreviewCommand{
			mode: mode,
		}
	}
}

func scaleMonitorCmd(monitorID int, delta Delta) tea.Cmd {
	return func() tea.Msg {
		return ScaleMonitorCommand{
			monitorID: monitorID,
			delta:     delta,
		}
	}
}

func toggleMonitorCmd(monitor *MonitorSpec) tea.Cmd {
	return func() tea.Msg {
		return ToggleMonitorCommand{
			monitorID: *monitor.ID,
		}
	}
}

func toggleMonitorVRRCmd(monitor *MonitorSpec) tea.Cmd {
	return func() tea.Msg {
		return ToggleMonitorVRRCommand{
			monitorID: *monitor.ID,
		}
	}
}

// rotateMonitorCmd cycles through the rotations
func rotateMonitorCmd(monitor *MonitorSpec) tea.Cmd {
	return func() tea.Msg {
		return RotateMonitorCommand{
			monitorID: *monitor.ID,
		}
	}
}

func moveMonitorCmd(monitorID int, stepX, stepY Delta) tea.Cmd {
	return func() tea.Msg {
		return MoveMonitorCommand{
			monitorID: monitorID,
			stepX:     stepX,
			stepY:     stepY,
		}
	}
}
