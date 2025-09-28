package tui

import tea "github.com/charmbracelet/bubbletea"

type MonitorBeingEdited struct {
	ListIndex int
	Scaling   bool
	MonitorID int
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
)

type OperationStatus struct {
	name string
	err  error
}

func operationStatusCmd(name OperationName, err error) tea.Cmd {
	return func() tea.Msg {
		return OperationStatus{}
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
