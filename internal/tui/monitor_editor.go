package tui

import (
	"fmt"
)

// MonitorEditor handles all monitor editing operations
type MonitorEditor struct {
	monitors     []*MonitorSpec
	positionStep int
	scaleStep    float64
	snapDistance int
}

func NewMonitorEditor(monitors []*MonitorSpec) *MonitorEditor {
	return &MonitorEditor{
		monitors:     monitors,
		positionStep: 50,  // Move monitors by 50 pixels at a time
		scaleStep:    0.1, // Adjust scale by 0.1 at a time
		snapDistance: 50,  // Snap when within 50 pixels of another monitor
	}
}

func (e *MonitorEditor) GetMonitors() []*MonitorSpec {
	return e.monitors
}

func (e *MonitorEditor) GetMonitor(i int) *MonitorSpec {
	return e.monitors[i]
}

func (e *MonitorEditor) UpdateMonitors(monitors []*MonitorSpec) {
	e.monitors = monitors
}

type MoveResult int

const (
	MoveSuccess MoveResult = iota
	MoveBlockedByConnectivity
	MoveNoChange
)

// MoveMonitor moves a monitor by the given delta, phasing through overlaps and maintaining connectivity
func (e *MonitorEditor) MoveMonitor(index int, dx, dy int) (MoveResult, error) {
	if index < 0 || index >= len(e.monitors) {
		return MoveNoChange, fmt.Errorf("invalid monitor index: %d", index)
	}

	monitor := e.monitors[index]
	newX := monitor.X + dx
	newY := monitor.Y + dy

	// Phase through overlapping monitors to find final position
	finalX, finalY := e.phaseThrough(index, newX, newY, dx, dy)

	// Check if the final position maintains connectivity (touches another monitor)
	if !e.hasConnectivity(index, finalX, finalY) {
		// If no connectivity, don't move and return blocked status
		return MoveBlockedByConnectivity, nil
	}

	// Apply snapping to final position
	snappedX, snappedY := e.snapPosition(index, finalX, finalY)
	monitor.X = snappedX
	monitor.Y = snappedY

	return MoveSuccess, nil
}

// MoveMonitorTo moves a monitor to absolute coordinates, phasing through overlaps and maintaining connectivity
func (e *MonitorEditor) MoveMonitorTo(index int, x, y int) (MoveResult, error) {
	if index < 0 || index >= len(e.monitors) {
		return MoveNoChange, fmt.Errorf("invalid monitor index: %d", index)
	}

	monitor := e.monitors[index]

	// Calculate movement delta for phase-through logic
	dx := x - monitor.X
	dy := y - monitor.Y

	// Phase through overlapping monitors to find final position
	finalX, finalY := e.phaseThrough(index, x, y, dx, dy)

	// Check if the final position maintains connectivity (touches another monitor)
	if !e.hasConnectivity(index, finalX, finalY) {
		// If no connectivity, don't move and return blocked status
		return MoveBlockedByConnectivity, nil
	}

	// Apply snapping to final position
	snappedX, snappedY := e.snapPosition(index, finalX, finalY)
	monitor.X = snappedX
	monitor.Y = snappedY

	return MoveSuccess, nil
}

// ScaleMonitor adjusts the scale of a monitor
func (e *MonitorEditor) ScaleMonitor(index int, delta float64) error {
	if index < 0 || index >= len(e.monitors) {
		return fmt.Errorf("invalid monitor index: %d", index)
	}

	monitor := e.monitors[index]
	newScale := monitor.Scale + delta

	// Ensure minimum scale of 0.1
	if newScale >= 0.1 {
		monitor.Scale = newScale

		// Snap the scaled monitor to the nearest edge if it's disconnected
		e.snapMonitorToNearestEdge(index)
	}

	return nil
}

// SetMonitorScale sets the absolute scale of a monitor
func (e *MonitorEditor) SetMonitorScale(index int, scale float64) error {
	if index < 0 || index >= len(e.monitors) {
		return fmt.Errorf("invalid monitor index: %d", index)
	}

	if scale < 0.1 {
		scale = 0.1
	}

	e.monitors[index].Scale = scale
	return nil
}

// RotateMonitor rotates a monitor by 90 degrees clockwise
func (e *MonitorEditor) RotateMonitor(index int) error {
	if index < 0 || index >= len(e.monitors) {
		return fmt.Errorf("invalid monitor index: %d", index)
	}

	monitor := e.monitors[index]
	monitor.Transform = (monitor.Transform + 1) % 4

	// Snap the rotated monitor to the nearest edge if it's disconnected
	e.snapMonitorToNearestEdge(index)

	return nil
}

// ToggleVRR toggles Variable Refresh Rate for a monitor
func (e *MonitorEditor) ToggleVRR(index int) error {
	if index < 0 || index >= len(e.monitors) {
		return fmt.Errorf("invalid monitor index: %d", index)
	}

	monitor := e.monitors[index]
	monitor.Vrr = !monitor.Vrr
	return nil
}

// ToggleDisable toggles the disabled state of a monitor
func (e *MonitorEditor) ToggleDisable(index int) error {
	if index < 0 || index >= len(e.monitors) {
		return fmt.Errorf("invalid monitor index: %d", index)
	}

	monitor := e.monitors[index]
	monitor.Disabled = !monitor.Disabled
	return nil
}

// SetMode sets the display mode for a monitor
func (e *MonitorEditor) SetMode(index int, modeIndex int) error {
	if index < 0 || index >= len(e.monitors) {
		return fmt.Errorf("invalid monitor index: %d", index)
	}

	monitor := e.monitors[index]
	if modeIndex < 0 || modeIndex >= len(monitor.AvailableModes) {
		return fmt.Errorf("invalid mode index: %d", modeIndex)
	}

	mode := monitor.AvailableModes[modeIndex]

	// Parse mode string (format: "2880x1920@120.00Hz")
	var width, height int
	var refreshRate float64

	n, err := fmt.Sscanf(mode, "%dx%d@%fHz", &width, &height, &refreshRate)
	if err != nil || n != 3 {
		return fmt.Errorf("failed to parse mode: %s", mode)
	}

	// Update monitor properties
	monitor.Width = width
	monitor.Height = height
	monitor.RefreshRate = refreshRate

	// Snap the monitor to the nearest edge if it's disconnected after mode change
	e.snapMonitorToNearestEdge(index)

	return nil
}

// FindCurrentModeIndex finds the index of the current monitor mode in AvailableModes
func (e *MonitorEditor) FindCurrentModeIndex(index int) int {
	if index < 0 || index >= len(e.monitors) {
		return 0
	}

	monitor := e.monitors[index]
	currentMode := fmt.Sprintf("%dx%d@%.2fHz", monitor.Width, monitor.Height, monitor.RefreshRate)

	for i, mode := range monitor.AvailableModes {
		if mode == currentMode {
			return i
		}
	}

	// If exact match not found, return 0 (first mode)
	return 0
}

// GetPositionStep returns the current position step size
func (e *MonitorEditor) GetPositionStep() int {
	return e.positionStep
}

// SetPositionStep sets the position step size
func (e *MonitorEditor) SetPositionStep(step int) {
	if step > 0 {
		e.positionStep = step
	}
}

// snapPosition attempts to snap the monitor position to nearby monitors
func (e *MonitorEditor) snapPosition(monitorIndex int, newX, newY int) (int, int) {
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
		if abs(currentRight-otherLeft) <= e.snapDistance && abs(oldRight-otherLeft) > e.snapDistance/2 {
			snappedX = otherLeft - scaledWidth
		}
		if abs(currentLeft-otherRight) <= e.snapDistance && abs(oldLeft-otherRight) > e.snapDistance/2 {
			snappedX = otherRight
		}
		if abs(currentLeft-otherLeft) <= e.snapDistance && abs(oldLeft-otherLeft) > e.snapDistance/2 {
			snappedX = otherLeft
		}
		if abs(currentRight-otherRight) <= e.snapDistance && abs(oldRight-otherRight) > e.snapDistance/2 {
			snappedX = otherRight - scaledWidth
		}

		// Check vertical snapping - only snap if we weren't already close before
		if abs(currentBottom-otherTop) <= e.snapDistance && abs(oldBottom-otherTop) > e.snapDistance/2 {
			snappedY = otherTop - scaledHeight
		}
		if abs(currentTop-otherBottom) <= e.snapDistance && abs(oldTop-otherBottom) > e.snapDistance/2 {
			snappedY = otherBottom
		}
		if abs(currentTop-otherTop) <= e.snapDistance && abs(oldTop-otherTop) > e.snapDistance/2 {
			snappedY = otherTop
		}
		if abs(currentBottom-otherBottom) <= e.snapDistance && abs(oldBottom-otherBottom) > e.snapDistance/2 {
			snappedY = otherBottom - scaledHeight
		}
	}

	return snappedX, snappedY
}

// getMonitorEdges returns the edges of a monitor accounting for scale and rotation
func (e *MonitorEditor) getMonitorEdges(monitor *MonitorSpec) (left, right, top, bottom int) {
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
func (e *MonitorEditor) phaseThrough(movingIndex int, targetX, targetY, dx, dy int) (int, int) {
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
func (e *MonitorEditor) calculatePhaseDistance(movingIndex int, startX, startY, dx, dy int) int {
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
func (e *MonitorEditor) wouldOverlap(x1, y1, w1, h1, x2, y2, w2, h2 int) bool {
	return x1 < x2+w2 && x1+w1 > x2 && y1 < y2+h2 && y1+h1 > y2
}

// hasOverlapAt checks if a monitor would overlap with any other monitor at the given position
func (e *MonitorEditor) hasOverlapAt(movingIndex int, x, y int) bool {
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
func (e *MonitorEditor) rectanglesOverlap(x1, y1, x2, y2, x3, y3, x4, y4 int) bool {
	return x1 < x4 && x2 > x3 && y1 < y4 && y2 > y3
}

// hasConnectivity checks if a monitor at the given position touches at least one other monitor
func (e *MonitorEditor) hasConnectivity(movingIndex int, x, y int) bool {
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
		if (abs(movingBottom-otherTop) <= tolerance || abs(movingTop-otherBottom) <= tolerance) &&
			!(movingRight <= otherLeft-tolerance || movingLeft >= otherRight+tolerance) {
			return true
		}

		// Vertical edges (left/right close)
		if (abs(movingRight-otherLeft) <= tolerance || abs(movingLeft-otherRight) <= tolerance) &&
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
func (e *MonitorEditor) monitorsConnectedAtCorner(x1, y1, x2, y2, x3, y3, x4, y4, tolerance int) bool {
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
func (e *MonitorEditor) getMonitorDimensions(monitor *MonitorSpec) (int, int) {
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
func (e *MonitorEditor) snapMonitorToNearestEdge(monitorIndex int) {
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

// findNearestConnectedPosition finds the nearest position where a monitor would be connected
func (e *MonitorEditor) findNearestConnectedPosition(monitorIndex int) (int, int) {
	monitor := e.monitors[monitorIndex]
	currentX, currentY := monitor.X, monitor.Y

	// If there's only one active monitor, stay at current position
	activeMonitors := 0
	for _, m := range e.monitors {
		if !m.Disabled {
			activeMonitors++
		}
	}
	if activeMonitors <= 1 {
		return currentX, currentY
	}

	monitorWidth, _ := e.getMonitorDimensions(monitor)

	// Simple approach: just put it to the right of the first other monitor
	for i, otherMonitor := range e.monitors {
		if i == monitorIndex || otherMonitor.Disabled {
			continue
		}

		otherWidth, _ := e.getMonitorDimensions(otherMonitor)

		// Try right of other monitor first
		newX := otherMonitor.X + otherWidth
		newY := otherMonitor.Y

		if e.hasConnectivity(monitorIndex, newX, newY) {
			return newX, newY
		}

		// Try left of other monitor
		newX = otherMonitor.X - monitorWidth
		newY = otherMonitor.Y

		if e.hasConnectivity(monitorIndex, newX, newY) {
			return newX, newY
		}
	}

	// If nothing worked, return current position
	return currentX, currentY
}
