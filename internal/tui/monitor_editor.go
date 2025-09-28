package tui

import (
	"errors"

	tea "github.com/charmbracelet/bubbletea"
)

// MonitorEditorStore handles all monitor editing operations
type MonitorEditorStore struct {
	monitors     []*MonitorSpec
	positionStep int
	scaleStep    float64
	snapDistance int
}

func NewMonitorEditor(monitors []*MonitorSpec) *MonitorEditorStore {
	return &MonitorEditorStore{
		monitors:     monitors,
		positionStep: 50,  // Move monitors by 50 pixels at a time
		scaleStep:    0.1, // Adjust scale by 0.1 at a time
		snapDistance: 50,  // Snap when within 50 pixels of another monitor
	}
}

type MoveResult int

const (
	MoveSuccess MoveResult = iota
	MoveBlockedByConnectivity
	MoveNoChange
)

// TODO like in scale first check if not connected if so snap to nearest

func (e *MonitorEditorStore) GetMoveDelta(delta Delta) int {
	switch delta {
	case DeltaLess:
		return -e.positionStep
	case DeltaMore:
		return e.positionStep
	}
	return 0
}

// MoveMonitor moves a monitor by the given delta, phasing through overlaps and maintaining connectivity
func (e *MonitorEditorStore) MoveMonitor(monitorID int, dx, dy Delta) tea.Cmd {
	monitor, index, err := e.FindByID(monitorID)
	if err != nil {
		return operationStatusCmd(OperationNameMove, err)
	}

	dxValue := e.GetMoveDelta(dx)
	dyValue := e.GetMoveDelta(dy)
	newX := monitor.X + dxValue
	newY := monitor.Y + dyValue

	// Phase through overlapping monitors to find final position
	finalX, finalY := e.phaseThrough(index, newX, newY, dxValue, dyValue)

	// Check if the final position maintains connectivity (touches another monitor)
	if !e.hasConnectivity(index, finalX, finalY) {
		// If no connectivity, don't move and return blocked status
		return nil
	}

	// Apply snapping to final position
	snappedX, snappedY := e.snapPosition(index, finalX, finalY)
	monitor.X = snappedX
	monitor.Y = snappedY

	return nil
}

// ScaleMonitor adjusts the scale of a monitor
func (e *MonitorEditorStore) ScaleMonitor(monitorID int, delta Delta) tea.Cmd {
	monitor, index, err := e.FindByID(monitorID)
	if err != nil {
		return operationStatusCmd(OperationNameScale, err)
	}

	deltaValue := 0.0
	switch delta {
	case DeltaLess:
		deltaValue = -e.scaleStep
	case DeltaMore:
		deltaValue = e.scaleStep
	}

	newScale := monitor.Scale + deltaValue

	// Ensure minimum scale of 0.1
	if newScale >= e.scaleStep {
		monitor.Scale = newScale

		// Snap the scaled monitor to the nearest edge if it's disconnected
		e.snapMonitorToNearestEdge(index)
	}

	return operationStatusCmd(OperationNameScale, err)
}

// RotateMonitor rotates a monitor by 90 degrees clockwise
func (e *MonitorEditorStore) RotateMonitor(monitorID int) tea.Cmd {
	monitor, index, err := e.FindByID(monitorID)
	if err != nil {
		return operationStatusCmd(OperationNameRotate, err)
	}

	monitor.Rotate()

	// Snap the rotated monitor to the nearest edge if it's disconnected
	e.snapMonitorToNearestEdge(index)

	return operationStatusCmd(OperationNameRotate, nil)
}

// ToggleVRR toggles Variable Refresh Rate for a monitor
func (e *MonitorEditorStore) ToggleVRR(monitorID int) tea.Cmd {
	monitor, _, err := e.FindByID(monitorID)
	if err != nil {
		return operationStatusCmd(OperationNameToggleVRR, err)
	}

	monitor.ToggleVRR()

	return operationStatusCmd(OperationNameToggleVRR, nil)
}

// ToggleDisable toggles the disabled state of a monitor
func (e *MonitorEditorStore) ToggleDisable(monitorID int) tea.Cmd {
	monitor, _, err := e.FindByID(monitorID)
	if err != nil {
		return operationStatusCmd(OperationNameToggleMonitor, err)
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

	found := false
	for _, monitor := range e.monitors {
		if monitor.Name != mirrorOf {
			continue
		}
		if monitor.Disabled {
			return operationStatusCmd(OperationNamePreviewMirror, errors.New("cant mirror disabled monitor"))
		}
		found = true

	}
	if !found && mirrorOf != "none" {
		return operationStatusCmd(OperationNamePreviewMirror, errors.New("cant find mirrored monitor"))
	}

	// todo check loops in mirroring, should be impossible to add a mirror that causes loops

	monitor.SetMirror(mirrorOf)

	return operationStatusCmd(OperationNamePreviewMirror, err)
}

func (e *MonitorEditorStore) SetMode(monitorID int, mode string) tea.Cmd {
	monitor, index, err := e.FindByID(monitorID)
	if err != nil {
		return operationStatusCmd(OperationNamePreviewMode, err)
	}

	err = monitor.SetMode(mode)
	if err != nil {
		return operationStatusCmd(OperationNamePreviewMode, err)
	}

	e.snapMonitorToNearestEdge(index)

	return operationStatusCmd(OperationNamePreviewMode, err)
}

// snapPosition attempts to snap the monitor position to nearby monitors
func (e *MonitorEditorStore) snapPosition(monitorIndex, newX, newY int) (int, int) {
	if monitorIndex < 0 || monitorIndex >= len(e.monitors) {
		return newX, newY
	}

	currentMonitor := e.monitors[monitorIndex]

	// Calculate scaled dimensions for current monitor
	scaledWidth := int(float64(currentMonitor.Width) / currentMonitor.Scale)
	scaledHeight := int(float64(currentMonitor.Height) / currentMonitor.Scale)

	// Account for rotation
	if currentMonitor.Transform == 1 || currentMonitor.Transform == 3 {
		scaledWidth, scaledHeight = scaledHeight, scaledWidth
	}

	// Calculate current edges (before movement)
	oldLeft := currentMonitor.X
	oldRight := currentMonitor.X + scaledWidth
	oldTop := currentMonitor.Y
	oldBottom := currentMonitor.Y + scaledHeight

	// Calculate edges with new position
	currentLeft := newX
	currentRight := newX + scaledWidth
	currentTop := newY
	currentBottom := newY + scaledHeight

	snappedX, snappedY := newX, newY

	// Check against all other monitors
	for i, otherMonitor := range e.monitors {
		if i == monitorIndex || otherMonitor.Disabled {
			continue
		}

		otherLeft, otherRight, otherTop, otherBottom := e.getMonitorEdges(otherMonitor)

		// Check horizontal snapping - only snap if we weren't already close before
		if abs(currentRight-otherLeft) <= e.snapDistance &&
			abs(oldRight-otherLeft) > e.snapDistance/2 {
			snappedX = otherLeft - scaledWidth
		}
		if abs(currentLeft-otherRight) <= e.snapDistance &&
			abs(oldLeft-otherRight) > e.snapDistance/2 {
			snappedX = otherRight
		}
		if abs(currentLeft-otherLeft) <= e.snapDistance &&
			abs(oldLeft-otherLeft) > e.snapDistance/2 {
			snappedX = otherLeft
		}
		if abs(currentRight-otherRight) <= e.snapDistance &&
			abs(oldRight-otherRight) > e.snapDistance/2 {
			snappedX = otherRight - scaledWidth
		}

		// Check vertical snapping - only snap if we weren't already close before
		if abs(currentBottom-otherTop) <= e.snapDistance &&
			abs(oldBottom-otherTop) > e.snapDistance/2 {
			snappedY = otherTop - scaledHeight
		}
		if abs(currentTop-otherBottom) <= e.snapDistance &&
			abs(oldTop-otherBottom) > e.snapDistance/2 {
			snappedY = otherBottom
		}
		if abs(currentTop-otherTop) <= e.snapDistance &&
			abs(oldTop-otherTop) > e.snapDistance/2 {
			snappedY = otherTop
		}
		if abs(currentBottom-otherBottom) <= e.snapDistance &&
			abs(oldBottom-otherBottom) > e.snapDistance/2 {
			snappedY = otherBottom - scaledHeight
		}
	}

	return snappedX, snappedY
}

// getMonitorEdges returns the edges of a monitor accounting for scale and rotation
func (e *MonitorEditorStore) getMonitorEdges(monitor *MonitorSpec) (left, right, top, bottom int) {
	// Calculate scaled dimensions
	scaledWidth := int(float64(monitor.Width) / monitor.Scale)
	scaledHeight := int(float64(monitor.Height) / monitor.Scale)

	// Account for rotation
	if monitor.Transform == 1 || monitor.Transform == 3 {
		scaledWidth, scaledHeight = scaledHeight, scaledWidth
	}

	left = monitor.X
	right = monitor.X + scaledWidth
	top = monitor.Y
	bottom = monitor.Y + scaledHeight

	return
}

// abs returns the absolute value of an integer
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// phaseThrough calculates the final position when phasing through overlapping monitors
func (e *MonitorEditorStore) phaseThrough(movingIndex, targetX, targetY, dx, dy int) (int, int) {
	// If there's no overlap at target position, just return it
	if !e.hasOverlapAt(movingIndex, targetX, targetY) {
		return targetX, targetY
	}

	// Find all overlapping monitors and calculate how far we need to phase through
	maxPhaseDistance := e.calculatePhaseDistance(movingIndex, targetX, targetY, dx, dy)

	// If we can't phase through, stay in original position
	if maxPhaseDistance == 0 {
		movingMonitor := e.monitors[movingIndex]
		return movingMonitor.X, movingMonitor.Y
	}

	// Calculate final position after phasing through
	finalX := targetX
	finalY := targetY

	if dx != 0 {
		if dx > 0 {
			finalX += maxPhaseDistance
		} else {
			finalX -= maxPhaseDistance
		}
	}

	if dy != 0 {
		if dy > 0 {
			finalY += maxPhaseDistance
		} else {
			finalY -= maxPhaseDistance
		}
	}

	return finalX, finalY
}

// calculatePhaseDistance calculates how far to move to completely phase through overlapping monitors
func (e *MonitorEditorStore) calculatePhaseDistance(movingIndex, startX, startY, dx, dy int) int {
	movingMonitor := e.monitors[movingIndex]
	movingWidth, movingHeight := e.getMonitorDimensions(movingMonitor)

	maxDistance := 0

	// Check each overlapping monitor and find the maximum distance needed
	for i, otherMonitor := range e.monitors {
		if i == movingIndex || otherMonitor.Disabled {
			continue
		}

		otherWidth, otherHeight := e.getMonitorDimensions(otherMonitor)

		// Check if there would be an overlap at the start position
		if e.wouldOverlap(startX, startY, movingWidth, movingHeight,
			otherMonitor.X, otherMonitor.Y, otherWidth, otherHeight) {

			// Calculate distance needed to clear this monitor
			var distance int
			if dx > 0 {
				// Moving right: need to clear the right edge
				distance = (otherMonitor.X + otherWidth) - startX
			} else if dx < 0 {
				// Moving left: need to clear the left edge
				distance = startX - otherMonitor.X + movingWidth
			} else if dy > 0 {
				// Moving down: need to clear the bottom edge
				distance = (otherMonitor.Y + otherHeight) - startY
			} else if dy < 0 {
				// Moving up: need to clear the top edge
				distance = startY - otherMonitor.Y + movingHeight
			}

			if distance > maxDistance {
				maxDistance = distance
			}
		}
	}

	return maxDistance
}

// wouldOverlap checks if two rectangles would overlap
func (e *MonitorEditorStore) wouldOverlap(x1, y1, w1, h1, x2, y2, w2, h2 int) bool {
	return x1 < x2+w2 && x1+w1 > x2 && y1 < y2+h2 && y1+h1 > y2
}

// hasOverlapAt checks if a monitor would overlap with any other monitor at the given position
func (e *MonitorEditorStore) hasOverlapAt(movingIndex, x, y int) bool {
	movingMonitor := e.monitors[movingIndex]
	movingWidth, movingHeight := e.getMonitorDimensions(movingMonitor)

	movingLeft := x
	movingRight := x + movingWidth
	movingTop := y
	movingBottom := y + movingHeight

	for i, otherMonitor := range e.monitors {
		if i == movingIndex || otherMonitor.Disabled {
			continue
		}

		otherWidth, otherHeight := e.getMonitorDimensions(otherMonitor)
		otherLeft := otherMonitor.X
		otherRight := otherMonitor.X + otherWidth
		otherTop := otherMonitor.Y
		otherBottom := otherMonitor.Y + otherHeight

		// Check for overlap
		if e.rectanglesOverlap(movingLeft, movingTop, movingRight, movingBottom,
			otherLeft, otherTop, otherRight, otherBottom) {
			return true
		}
	}

	return false
}

// rectanglesOverlap checks if two rectangles overlap
func (e *MonitorEditorStore) rectanglesOverlap(x1, y1, x2, y2, x3, y3, x4, y4 int) bool {
	return x1 < x4 && x2 > x3 && y1 < y4 && y2 > y3
}

// hasConnectivity checks if a monitor at the given position touches at least one other monitor
func (e *MonitorEditorStore) hasConnectivity(movingIndex, x, y int) bool {
	// If there's only one monitor, it's always connected
	activeMonitors := 0
	for _, monitor := range e.monitors {
		if !monitor.Disabled {
			activeMonitors++
		}
	}
	if activeMonitors <= 1 {
		return true
	}

	movingMonitor := e.monitors[movingIndex]
	movingWidth, movingHeight := e.getMonitorDimensions(movingMonitor)

	movingLeft := x
	movingRight := x + movingWidth
	movingTop := y
	movingBottom := y + movingHeight

	// Allow some tolerance for edge connections (a few pixels off is OK)
	tolerance := 10

	for i, otherMonitor := range e.monitors {
		if i == movingIndex || otherMonitor.Disabled {
			continue
		}

		otherWidth, otherHeight := e.getMonitorDimensions(otherMonitor)
		otherLeft := otherMonitor.X
		otherRight := otherMonitor.X + otherWidth
		otherTop := otherMonitor.Y
		otherBottom := otherMonitor.Y + otherHeight

		// Check if monitors are close enough to be considered connected
		// Horizontal edges (top/bottom close)
		if (abs(movingBottom-otherTop) <= tolerance ||
			abs(movingTop-otherBottom) <= tolerance) &&
			!(movingRight <= otherLeft-tolerance || movingLeft >= otherRight+tolerance) {
			return true
		}

		// Vertical edges (left/right close)
		if (abs(movingRight-otherLeft) <= tolerance ||
			abs(movingLeft-otherRight) <= tolerance) &&
			!(movingBottom <= otherTop-tolerance || movingTop >= otherBottom+tolerance) {
			return true
		}

		// Also allow corner connections (monitors touching at corners)
		if e.monitorsConnectedAtCorner(movingLeft, movingTop, movingRight, movingBottom,
			otherLeft, otherTop, otherRight, otherBottom, tolerance) {
			return true
		}
	}

	return false
}

// monitorsConnectedAtCorner checks if two monitors are connected at their corners
func (e *MonitorEditorStore) monitorsConnectedAtCorner(x1, y1, x2, y2, x3, y3, x4, y4, tolerance int) bool {
	// Check if any corner of the first monitor is close to any corner of the second monitor
	corners1 := [][2]int{{x1, y1}, {x2, y1}, {x1, y2}, {x2, y2}}
	corners2 := [][2]int{{x3, y3}, {x4, y3}, {x3, y4}, {x4, y4}}

	for _, c1 := range corners1 {
		for _, c2 := range corners2 {
			if abs(c1[0]-c2[0]) <= tolerance && abs(c1[1]-c2[1]) <= tolerance {
				return true
			}
		}
	}

	return false
}

// getMonitorDimensions returns the effective width and height of a monitor accounting for scale and rotation
func (e *MonitorEditorStore) getMonitorDimensions(monitor *MonitorSpec) (int, int) {
	// Calculate scaled dimensions
	scaledWidth := int(float64(monitor.Width) / monitor.Scale)
	scaledHeight := int(float64(monitor.Height) / monitor.Scale)

	// Account for rotation
	if monitor.Transform == 1 || monitor.Transform == 3 {
		return scaledHeight, scaledWidth
	}

	return scaledWidth, scaledHeight
}

// snapMonitorToNearestEdge snaps a specific monitor to the nearest edge of another monitor if it's disconnected
func (e *MonitorEditorStore) snapMonitorToNearestEdge(monitorIndex int) {
	if monitorIndex < 0 || monitorIndex >= len(e.monitors) {
		return
	}

	monitor := e.monitors[monitorIndex]
	if monitor.Disabled {
		return
	}

	// Count active monitors
	activeMonitors := 0
	for _, m := range e.monitors {
		if !m.Disabled {
			activeMonitors++
		}
	}

	if activeMonitors <= 1 {
		return
	}

	// Check if this monitor is still connected
	if e.hasConnectivity(monitorIndex, monitor.X, monitor.Y) {
		return // Already connected, no need to move
	}

	// Find the closest edge of any other monitor
	monitorWidth, monitorHeight := e.getMonitorDimensions(monitor)
	minDistance := float64(999999)
	bestX, bestY := monitor.X, monitor.Y

	for i, otherMonitor := range e.monitors {
		if i == monitorIndex || otherMonitor.Disabled {
			continue
		}

		otherWidth, otherHeight := e.getMonitorDimensions(otherMonitor)

		// Try all edge positions around this other monitor
		candidatePositions := []struct {
			x, y int
			desc string
		}{
			// Right edge
			{otherMonitor.X + otherWidth, otherMonitor.Y, "right"},
			// Left edge
			{otherMonitor.X - monitorWidth, otherMonitor.Y, "left"},
			// Bottom edge
			{otherMonitor.X, otherMonitor.Y + otherHeight, "bottom"},
			// Top edge
			{otherMonitor.X, otherMonitor.Y - monitorHeight, "top"},
		}

		for _, pos := range candidatePositions {
			// Check if this position would have connectivity
			if e.hasConnectivity(monitorIndex, pos.x, pos.y) {
				// Calculate distance from current position
				dx := float64(pos.x - monitor.X)
				dy := float64(pos.y - monitor.Y)
				distance := dx*dx + dy*dy

				if distance < minDistance {
					minDistance = distance
					bestX, bestY = pos.x, pos.y
				}
			}
		}
	}

	// Apply the best position found
	if bestX != monitor.X || bestY != monitor.Y {
		// Apply snapping to the new position
		snappedX, snappedY := e.snapPosition(monitorIndex, bestX, bestY)
		monitor.X = snappedX
		monitor.Y = snappedY
	}
}

func (e *MonitorEditorStore) FindByID(monitorID int) (*MonitorSpec, int, error) {
	for index, monitor := range e.monitors {
		if *monitor.ID == monitorID {
			return monitor, index, nil
		}
	}
	return nil, 0, errors.New("cant find monitor")
}
