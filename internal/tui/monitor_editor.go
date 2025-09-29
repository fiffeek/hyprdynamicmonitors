package tui

import (
	"errors"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/fiffeek/hyprdynamicmonitors/internal/utils"
)

var ErrMonitorDisabled = errors.New("monitor is disabled")

// MonitorEditorStore handles all monitor editing operations
type MonitorEditorStore struct {
	monitors     []*MonitorSpec
	positionStep int
	scaleStep    float64
	snapDistance int
	snapping     bool
}

func NewMonitorEditor(monitors []*MonitorSpec) *MonitorEditorStore {
	return &MonitorEditorStore{
		monitors:     monitors,
		positionStep: 50,
		scaleStep:    0.1,
		snapDistance: 50,
		snapping:     true,
	}
}

func (e *MonitorEditorStore) SetSnapping(snap bool) {
	e.snapping = snap
}

func (e *MonitorEditorStore) GetMonitors() []*MonitorSpec {
	return e.monitors
}

func (e *MonitorEditorStore) GetMoveDelta(delta Delta) int {
	switch delta {
	case DeltaLess:
		return -e.positionStep
	case DeltaMore:
		return e.positionStep
	}
	return 0
}

func (e *MonitorEditorStore) MoveMonitor(monitorID int, dx, dy Delta) tea.Cmd {
	monitor, index, err := e.FindByID(monitorID)
	if err != nil {
		return operationStatusCmd(OperationNameMove, err)
	}

	if monitor.Disabled {
		return operationStatusCmd(OperationNameMove, ErrMonitorDisabled)
	}

	dxValue := e.GetMoveDelta(dx)
	dyValue := e.GetMoveDelta(dy)
	newX := monitor.X + dxValue
	newY := monitor.Y + dyValue

	if !e.snapping {
		monitor.X = newX
		monitor.Y = newY
		return nil
	}

	snappedX, snappedY, cmd := e.snapToEdges(index, newX, newY)
	monitor.X = snappedX
	monitor.Y = snappedY

	return cmd
}

func (e *MonitorEditorStore) ScaleMonitor(monitorID int, newScale float64) tea.Cmd {
	monitor, _, err := e.FindByID(monitorID)
	if err != nil {
		return operationStatusCmd(OperationNameScale, err)
	}

	if monitor.Disabled {
		return operationStatusCmd(OperationNameScale, ErrMonitorDisabled)
	}

	if newScale >= e.scaleStep {
		monitor.Scale = newScale
	}

	return operationStatusCmd(OperationNameScale, err)
}

func (e *MonitorEditorStore) RotateMonitor(monitorID int) tea.Cmd {
	monitor, _, err := e.FindByID(monitorID)
	if err != nil {
		return operationStatusCmd(OperationNameRotate, err)
	}

	if monitor.Disabled {
		return operationStatusCmd(OperationNameRotate, ErrMonitorDisabled)
	}

	monitor.Rotate()

	return operationStatusCmd(OperationNameRotate, nil)
}

func (e *MonitorEditorStore) ToggleVRR(monitorID int) tea.Cmd {
	monitor, _, err := e.FindByID(monitorID)
	if err != nil {
		return operationStatusCmd(OperationNameToggleVRR, err)
	}

	if monitor.Disabled {
		return operationStatusCmd(OperationNameToggleVRR, ErrMonitorDisabled)
	}

	monitor.ToggleVRR()

	return operationStatusCmd(OperationNameToggleVRR, nil)
}

func (e *MonitorEditorStore) ToggleDisable(monitorID int) tea.Cmd {
	monitor, _, err := e.FindByID(monitorID)
	if err != nil {
		return operationStatusCmd(OperationNameToggleMonitor, err)
	}

	currentEnabled := false
	anyOtherEnabled := false
	for _, monitor := range e.monitors {
		if *monitor.ID == monitorID {
			currentEnabled = !monitor.Disabled
			continue
		}
		if !monitor.Disabled {
			anyOtherEnabled = true
		}
	}

	if currentEnabled && !anyOtherEnabled {
		return operationStatusCmd(OperationNameToggleMonitor, errors.New("only one monitor left"))
	}

	monitor.ToggleMonitor()
	return operationStatusCmd(OperationNameToggleMonitor, nil)
}

func (e *MonitorEditorStore) SetMirror(monitorID int, mirrorOf string) tea.Cmd {
	monitor, _, err := e.FindByID(monitorID)
	if err != nil {
		return operationStatusCmd(OperationNamePreviewMirror, err)
	}

	if monitor.Name == mirrorOf {
		return operationStatusCmd(OperationNamePreviewMirror, errors.New("cant mirror itself"))
	}

	if monitor.Disabled {
		return operationStatusCmd(OperationNamePreviewMirror, ErrMonitorDisabled)
	}

	found := false
	for _, monitor := range e.monitors {
		if monitor.Name != mirrorOf {
			continue
		}
		if monitor.Disabled {
			return operationStatusCmd(OperationNamePreviewMirror,
				errors.New("cant mirror disabled monitor"))
		}
		found = true

	}
	if !found && mirrorOf != "none" {
		return operationStatusCmd(OperationNamePreviewMirror,
			errors.New("cant find mirrored monitor"))
	}

	if e.wouldCreateMirrorLoop(monitor.Name, mirrorOf) {
		return operationStatusCmd(OperationNamePreviewMirror, errors.New("would create mirror loop"))
	}

	monitor.SetMirror(mirrorOf)

	return operationStatusCmd(OperationNamePreviewMirror, err)
}

// wouldCreateMirrorLoop checks if setting monitorName to mirror mirrorOf would create a loop
func (e *MonitorEditorStore) wouldCreateMirrorLoop(monitorName, mirrorOf string) bool {
	if mirrorOf == "none" {
		return false
	}

	// Create a map of current mirror relationships for efficient lookup
	mirrorMap := make(map[string]string)
	for _, monitor := range e.monitors {
		if monitor.Mirror != "none" && monitor.Mirror != "" {
			mirrorMap[monitor.Name] = monitor.Mirror
		}
	}

	// Simulate adding the new mirror relationship
	mirrorMap[monitorName] = mirrorOf

	// Check if this creates a cycle using DFS
	visited := make(map[string]bool)
	recStack := make(map[string]bool)

	var hasCycle func(node string) bool
	hasCycle = func(node string) bool {
		if recStack[node] {
			return true // Found a cycle
		}
		if visited[node] {
			return false // Already processed this node
		}

		visited[node] = true
		recStack[node] = true

		// Follow the mirror chain
		if next, exists := mirrorMap[node]; exists {
			if hasCycle(next) {
				return true
			}
		}

		recStack[node] = false
		return false
	}

	// Check for cycles starting from any node
	for name := range mirrorMap {
		if !visited[name] {
			if hasCycle(name) {
				return true
			}
		}
	}

	return false
}

func (e *MonitorEditorStore) SetMode(monitorID int, mode string) tea.Cmd {
	monitor, _, err := e.FindByID(monitorID)
	if err != nil {
		return operationStatusCmd(OperationNamePreviewMode, err)
	}

	if monitor.Disabled {
		return operationStatusCmd(OperationNamePreviewMode, ErrMonitorDisabled)
	}

	err = monitor.SetMode(mode)
	if err != nil {
		return operationStatusCmd(OperationNamePreviewMode, err)
	}

	return operationStatusCmd(OperationNamePreviewMode, err)
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func (e *MonitorEditorStore) getMonitorDimensions(monitor *MonitorSpec) (int, int) {
	scaledWidth := int(float64(monitor.Width) / monitor.Scale)
	scaledHeight := int(float64(monitor.Height) / monitor.Scale)

	if monitor.NeedsDimensionsSwap() {
		return scaledHeight, scaledWidth
	}

	return scaledWidth, scaledHeight
}

func (e *MonitorEditorStore) snapToEdges(monitorIndex, x, y int) (int, int, tea.Cmd) {
	if !e.snapping {
		return x, y, nil
	}

	monitor := e.monitors[monitorIndex]
	if monitor.Disabled {
		return x, y, nil
	}

	monWidth, monHeight := e.getMonitorDimensions(monitor)
	thresh := e.snapDistance
	newX, newY := x, y
	var snappedX, snappedY *int

	for i, other := range e.monitors {
		if i == monitorIndex || other.Disabled {
			continue
		}

		otherWidth, otherHeight := e.getMonitorDimensions(other)

		// check all corners for both x and y
		switch {
		case abs(x-other.X-otherWidth) < thresh:
			newX = other.X + otherWidth
			snappedX = utils.IntPtr(other.X + otherWidth)
		case abs(x+monWidth-other.X) < thresh:
			newX = other.X - monWidth
			snappedX = utils.IntPtr(other.X)
		case abs(x-other.X) < thresh:
			newX = other.X
			snappedX = utils.IntPtr(other.X)
		case abs(x+monWidth-other.X-otherWidth) < thresh:
			newX = other.X + otherWidth - monWidth
			snappedX = utils.IntPtr(other.X + otherWidth)
		}

		switch {
		case abs(y-other.Y-otherHeight) < thresh:
			newY = other.Y + otherHeight
			snappedY = utils.IntPtr(other.Y + otherHeight)
		case abs(y+monHeight-other.Y) < thresh:
			newY = other.Y - monHeight
			snappedY = utils.IntPtr(other.Y)
		case abs(y-other.Y) < thresh:
			newY = other.Y
			snappedY = utils.IntPtr(other.Y)
		case abs(y+monHeight-other.Y-otherHeight) < thresh:
			newY = other.Y + otherHeight - monHeight
			snappedY = utils.IntPtr(other.Y + otherHeight)
		}
	}

	return newX, newY, showGridLineCmd(snappedX, snappedY)
}

func (e *MonitorEditorStore) FindByID(monitorID int) (*MonitorSpec, int, error) {
	for index, monitor := range e.monitors {
		if *monitor.ID == monitorID {
			return monitor, index, nil
		}
	}
	return nil, 0, errors.New("cant find monitor")
}
