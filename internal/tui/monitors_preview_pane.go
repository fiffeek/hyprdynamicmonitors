package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/fiffeek/hyprdynamicmonitors/internal/utils"
	"github.com/sirupsen/logrus"
)

type gridCell struct {
	color           string
	backgroundColor string
	char            rune
}

type MonitorsPreviewPane struct {
	monitors              []*MonitorSpec
	selectedIndex         int
	panX, panY            int
	originalVirtualWidth  int
	originalVirtualHeight int
	virtualWidth          int
	virtualHeight         int
	width                 int
	height                int
	baseSpacing           int
	minSpacingY           int
	minSpacingX           int
	panning               bool
	panStep               int
	snapping              bool
	followMonitor         bool
	snapGridX             *int
	snapGridY             *int
	zoomStep              float64
}

func NewMonitorsPreviewPane(monitors []*MonitorSpec) *MonitorsPreviewPane {
	pane := &MonitorsPreviewPane{
		monitors:              monitors,
		selectedIndex:         -1,
		panX:                  0,
		panY:                  0,
		virtualWidth:          8000,
		virtualHeight:         8000,
		originalVirtualWidth:  8000,
		originalVirtualHeight: 8000,
		baseSpacing:           200,
		minSpacingY:           2,
		minSpacingX:           4,
		panStep:               100,
		snapping:              true,
		zoomStep:              1.1,
	}

	pane.autoFitMonitors()

	return pane
}

// calculateMonitorBounds returns the bounding box of all non-disabled monitors
// Returns (left, right, top, bottom, hasMonitors)
func (p *MonitorsPreviewPane) calculateMonitorBounds() (int, int, int, int, bool) {
	var minX, maxX, minY, maxY int
	hasMonitors := false

	for _, monitor := range p.monitors {
		if monitor.Disabled {
			continue
		}

		scaledWidth := int(float64(monitor.Width) / monitor.Scale)
		scaledHeight := int(float64(monitor.Height) / monitor.Scale)
		visualWidth := scaledWidth
		visualHeight := scaledHeight

		if monitor.NeedsDimensionsSwap() {
			visualWidth = scaledHeight
			visualHeight = scaledWidth
		}

		monitorLeft := monitor.X
		monitorRight := monitor.X + visualWidth
		monitorTop := monitor.Y
		monitorBottom := monitor.Y + visualHeight

		if !hasMonitors {
			minX = monitorLeft
			maxX = monitorRight
			minY = monitorTop
			maxY = monitorBottom
			hasMonitors = true
			continue
		}

		if monitorLeft < minX {
			minX = monitorLeft
		}
		if monitorRight > maxX {
			maxX = monitorRight
		}
		if monitorTop < minY {
			minY = monitorTop
		}
		if monitorBottom > maxY {
			maxY = monitorBottom
		}
	}

	return minX, maxX, minY, maxY, hasMonitors
}

// autoFitMonitors centers the view on all monitors and adjusts zoom to fit
func (p *MonitorsPreviewPane) autoFitMonitors() {
	minX, maxX, minY, maxY, hasMonitors := p.calculateMonitorBounds()

	if !hasMonitors {
		return
	}

	p.panX = (minX + maxX) / 2
	p.panY = (minY + maxY) / 2
	monitorWidth := maxX - minX
	monitorHeight := maxY - minY
	// Add padding on both sides
	withPadding := 1.6

	if monitorWidth > 0 && monitorHeight > 0 {
		paddedWidth := int(float64(monitorWidth) * withPadding)
		paddedHeight := int(float64(monitorHeight) * withPadding)
		aspectRatio := p.AspectRatio()
		paddedHeight = int(float64(paddedHeight) * aspectRatio)

		// Lock both to the same value
		p.virtualWidth = max(paddedWidth, paddedHeight, 1000)
		p.virtualHeight = max(paddedWidth, paddedHeight, 1000)

		p.originalVirtualWidth = p.virtualWidth
		p.originalVirtualHeight = p.virtualHeight
	}
}

func (p *MonitorsPreviewPane) panToMonitorCenter() {
	monitor := p.monitors[p.selectedIndex]
	x, y := monitor.Center()
	p.panX = x
	p.panY = y
}

func (p *MonitorsPreviewPane) Update(msg tea.Msg) tea.Cmd {
	logrus.Debugf("Update called on MonitorsPreviewPane: %v", msg)
	switch msg := msg.(type) {
	case ShowGridLineCommand:
		logrus.Debugf("Show grid line: x=%v, y=%v", msg.x, msg.y)
		p.snapGridX = msg.x
		p.snapGridY = msg.y
	case MoveMonitorCommand:
		p.snapGridX = nil
		p.snapGridY = nil
		if p.followMonitor {
			p.panToMonitorCenter()
		}
	case MonitorBeingEdited:
		p.selectedIndex = msg.ListIndex
		p.panToMonitorCenter()
	case MonitorUnselected:
		p.selectedIndex = -1
	case StateChanged:
		p.panning = msg.State.IsPanning()
		p.snapping = msg.State.Snapping
		p.followMonitor = msg.State.MonitorFollowMode
		if msg.State.MirrorSelection || msg.State.ModeSelection || msg.State.Scaling {
			p.snapGridX = nil
			p.snapGridY = nil
		}

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, rootKeyMap.Up):
			if p.panning {
				p.panY -= p.panStep
			}
		case key.Matches(msg, rootKeyMap.Down):
			if p.panning {
				p.panY += p.panStep
			}
		case key.Matches(msg, rootKeyMap.Left):
			if p.panning {
				p.panX -= p.panStep
			}
		case key.Matches(msg, rootKeyMap.Right):
			if p.panning {
				p.panX += p.panStep
			}
		case key.Matches(msg, rootKeyMap.Center):
			p.panX = 0
			p.panY = 0
		case key.Matches(msg, rootKeyMap.ZoomIn):
			p.ZoomIn()
		case key.Matches(msg, rootKeyMap.ZoomOut):
			p.ZoomOut()
		case key.Matches(msg, rootKeyMap.ResetZoom):
			p.ResetZoom()
		case key.Matches(msg, rootKeyMap.FitMonitors):
			p.autoFitMonitors()
		}
	}

	return nil
}

func (p *MonitorsPreviewPane) SetHeight(h int) {
	p.height = h
}

func (p *MonitorsPreviewPane) SetWidth(w int) {
	logrus.Debugf("Setting with to %d", w)
	p.width = w
}

func (p *MonitorsPreviewPane) View() string {
	logrus.Debugf("Rendering with %d, %d", p.width, p.height)
	return p.Render(p.width, p.height)
}

func (p *MonitorsPreviewPane) Render(availableWidth, availableHeight int) string {
	if len(p.monitors) == 0 {
		return "No monitors detected"
	}

	title := TitleStyle.Render("Monitor Preview")
	legend := p.renderLegend()

	titleHeight := lipgloss.Height(title)
	legendHeight := lipgloss.Height(legend)
	titleWidth := lipgloss.Width(title)

	gridHeight := availableHeight - legendHeight - titleHeight
	grid := p.renderGrid(availableWidth-4, gridHeight)

	scaleInfo := p.ScaleInfo()
	scaleInfoWidth := lipgloss.Width(scaleInfo)

	return lipgloss.JoinVertical(
		lipgloss.Top,
		lipgloss.JoinHorizontal(
			lipgloss.Left,
			title,
			lipgloss.NewStyle().Width(availableWidth-titleWidth-scaleInfoWidth).Render(""),
			SubtitleInfoStyle.Render(scaleInfo),
		),
		grid,
		legend,
	)
}

func (p *MonitorsPreviewPane) ScaleInfo() string {
	scaleInfo := fmt.Sprintf("Virtual Area: %dx%d", p.virtualWidth, p.virtualHeight)
	if p.panX != 0 || p.panY != 0 {
		scaleInfo += fmt.Sprintf(" | Center: (%d,%d)", p.panX, p.panY)
	}
	if p.snapping {
		scaleInfo += " | Snapping"
	}
	if p.followMonitor {
		scaleInfo += " | Follow ON"
	}
	return scaleInfo
}

func (p *MonitorsPreviewPane) AspectRatio() float64 {
	ratio := 2.0
	logrus.Debugf("Aspect ratio %f", ratio)
	// Account for character aspect ratio: terminal characters are typically ~1:2 (width:height)
	// So we need to expand the Y scale to compensate for characters being taller than wide
	return ratio // height/width of a terminal character
}

func (p *MonitorsPreviewPane) renderGrid(gridWidth, gridHeight int) string {
	if gridWidth <= 0 || gridHeight <= 0 {
		return "Area too small for visualization"
	}

	scaleX := float64(gridWidth) / float64(p.virtualWidth)
	scaleY := float64(gridHeight) / float64(p.virtualHeight) * p.AspectRatio()
	grid := p.initializeGrid(gridHeight, gridWidth, scaleX, scaleY)

	// draw all monitors except the selected one
	for i, monitor := range p.monitors {
		if i != p.selectedIndex {
			p.DrawMonitor(i, monitor, scaleX, scaleY, grid)
		}
	}

	// draw the selected monitor on top
	if p.selectedIndex >= 0 && p.selectedIndex < len(p.monitors) {
		p.DrawMonitor(p.selectedIndex, p.monitors[p.selectedIndex], scaleX, scaleY, grid)
	}

	// draw snap grid lines if any
	p.drawSnapGridLines(grid, scaleX, scaleY, gridWidth, gridHeight)

	lines := p.ColorGrid(grid)
	return strings.Join(lines, "\n")
}

func (p *MonitorsPreviewPane) drawSnapGridLines(grid [][]gridCell, scaleX, scaleY float64, gridWidth, gridHeight int) {
	if p.snapGridX == nil && p.snapGridY == nil {
		return
	}

	centerX := gridWidth / 2
	centerY := gridHeight / 2

	if p.snapGridX != nil {
		lineX := centerX + int(float64(*p.snapGridX-p.panX)*scaleX)
		if lineX >= 0 && lineX < gridWidth {
			for y := 0; y < gridHeight; y++ {
				if grid[y][lineX].char == ' ' || grid[y][lineX].char == '·' {
					grid[y][lineX] = gridCell{char: '│', color: "243"} // Muted color
				}
			}
		}
	}

	if p.snapGridY != nil {
		lineY := centerY + int(float64(*p.snapGridY-p.panY)*scaleY)
		if lineY >= 0 && lineY < gridHeight {
			for x := 0; x < gridWidth; x++ {
				if grid[lineY][x].char == ' ' || grid[lineY][x].char == '·' {
					grid[lineY][x] = gridCell{char: '─', color: "243"} // Muted color
				}
			}
		}
	}
}

// ColorGrid converts the grid cells to the colorized representation
func (*MonitorsPreviewPane) ColorGrid(grid [][]gridCell) []string {
	var lines []string
	for _, row := range grid {
		var line strings.Builder
		var currentColor string
		var currentBgColor string
		var currentSegment strings.Builder

		for _, cell := range row {
			if cell.color != currentColor || cell.backgroundColor != currentBgColor {
				// Flush previous segment with its color
				if currentSegment.Len() > 0 {
					style := lipgloss.NewStyle()
					if currentColor != "" {
						style = style.Foreground(lipgloss.Color(currentColor))
					}
					if currentBgColor != "" {
						style = style.Background(lipgloss.Color(currentBgColor))
					}
					line.WriteString(style.Render(currentSegment.String()))
					currentSegment.Reset()
				}
				currentColor = cell.color
				currentBgColor = cell.backgroundColor
			}
			currentSegment.WriteRune(cell.char)
		}

		// Flush remaining segment
		if currentSegment.Len() > 0 {
			style := lipgloss.NewStyle()
			if currentColor != "" {
				style = style.Foreground(lipgloss.Color(currentColor))
			}
			if currentBgColor != "" {
				style = style.Background(lipgloss.Color(currentBgColor))
			}
			line.WriteString(style.Render(currentSegment.String()))
		}

		lines = append(lines, line.String())
	}
	return lines
}

func (p *MonitorsPreviewPane) DrawMonitor(i int, monitor *MonitorSpec, scaleX, scaleY float64, grid [][]gridCell) {
	if monitor.Disabled {
		return
	}

	gridWidth := len(grid[0])
	gridHeight := len(grid)

	scaledWidth := int(float64(monitor.Width) / monitor.Scale)
	scaledHeight := int(float64(monitor.Height) / monitor.Scale)
	visualWidth := scaledWidth
	visualHeight := scaledHeight

	if monitor.NeedsDimensionsSwap() {
		visualWidth = scaledHeight
		visualHeight = scaledWidth
	}

	// Convert monitor coordinates to grid coordinates
	// Center the coordinate system (0,0 becomes center of grid)
	centerX := gridWidth / 2
	centerY := gridHeight / 2

	// Apply pan offsets to shift the view
	startX := centerX + int(float64(monitor.X-p.panX)*scaleX)
	startY := centerY + int(float64(monitor.Y-p.panY)*scaleY)
	endX := startX + int(float64(visualWidth)*scaleX)
	endY := startY + int(float64(visualHeight)*scaleY)
	rectangle := NewMonitorRectangle(
		startX, startY, endX, endY, monitor,
	)
	rectangle.Clamp(gridWidth, gridHeight)

	// Skip if monitor is outside visible area
	if startX >= gridWidth || startY >= gridHeight || endX < 0 || endY < 0 {
		return
	}

	color := MonitorEdgeColors[i%len(MonitorEdgeColors)]
	isSelected := i == p.selectedIndex
	p.drawMonitorRectangle(isSelected, rectangle, color, grid)
	p.drawMonitorLabel(isSelected, monitor, rectangle, grid, color)
}

func (*MonitorsPreviewPane) drawMonitorLabel(isSelected bool, monitor *MonitorSpec,
	rectangle *MonitorRectangle, grid [][]gridCell, color string,
) {
	gridWidth := len(grid[0])
	gridHeight := len(grid)

	label := monitor.Name
	if len(label) > 6 {
		label = label[:6] // Truncate long names
	}
	if monitor.Flipped {
		label = utils.Reverse(label)
	}
	if isSelected {
		label = "*" + label
	}

	// Determine positioning arrow based on monitor's relative position
	arrow := monitor.PositionArrowView()
	fullLabel := label + arrow

	labelY := (rectangle.startY + rectangle.endY) / 2
	labelStartX := (rectangle.startX + rectangle.endX + 1 - len(fullLabel)) / 2

	if labelY >= 0 && labelY < gridHeight && labelStartX >= rectangle.startX && labelStartX <= rectangle.endX {
		for j, char := range fullLabel {
			labelX := labelStartX + j
			if labelX >= 0 && labelX < gridWidth && labelX >= rectangle.startX && labelX <= rectangle.endX {
				grid[labelY][labelX] = gridCell{char: char, color: color}
			}
		}
	}
}

func (*MonitorsPreviewPane) drawMonitorRectangle(isSelected bool, rectangle *MonitorRectangle,
	monitorEdgeColor string, grid [][]gridCell,
) {
	gridWidth := len(grid[0])
	gridHeight := len(grid)

	borderChar := '█'
	fillChar := '█'

	for y := rectangle.startY; y <= rectangle.endY; y++ {
		for x := rectangle.startX; x <= rectangle.endX; x++ {
			if y >= 0 && y < gridHeight && x >= 0 && x < gridWidth {

				edgeColor := monitorEdgeColor

				// Use a brighter/different shade for the bottom edge
				if rectangle.isBottomEdge(x, y) {
					edgeColor = GetMonitorBottomColor(monitorEdgeColor)
				}

				switch {
				case y == rectangle.startY && (x == rectangle.startX || x == rectangle.endX):
					grid[y][x].char = '▀'
					grid[y][x].color = monitorEdgeColor
					grid[y][x].backgroundColor = monitorEdgeColor
				case y == rectangle.startY && x > rectangle.startX && x < rectangle.endX:
					grid[y][x].char = '▄'
					grid[y][x].color = GetMonitorFillForEdge(monitorEdgeColor, isSelected)
					grid[y][x].backgroundColor = edgeColor
				case y == rectangle.endY && (x == rectangle.startX || x == rectangle.endX):
					grid[y][x].char = '▄'
					grid[y][x].color = monitorEdgeColor
					grid[y][x].backgroundColor = monitorEdgeColor
				case y == rectangle.endY && x > rectangle.startX && x < rectangle.endX:
					grid[y][x].char = '▀'
					grid[y][x].color = GetMonitorFillForEdge(monitorEdgeColor, isSelected)
					grid[y][x].backgroundColor = edgeColor
				case y == rectangle.startY || y == rectangle.endY ||
					x == rectangle.startX || x == rectangle.endX:
					grid[y][x] = gridCell{char: borderChar, color: edgeColor, backgroundColor: ""}
				default:
					grid[y][x] = gridCell{char: fillChar, color: GetMonitorFillForEdge(
						monitorEdgeColor, isSelected), backgroundColor: ""}
				}
			}
		}
	}
}

// initializeGrid creates a grid with dot character that react to scaling
func (p *MonitorsPreviewPane) initializeGrid(gridHeight, gridWidth int, scaleX, scaleY float64) [][]gridCell {
	grid := make([][]gridCell, gridHeight)
	for i := range grid {
		grid[i] = make([]gridCell, gridWidth)
		for j := range grid[i] {
			virtualSpacingX := int(float64(p.baseSpacing) * scaleX)
			virtualSpacingY := int(float64(p.baseSpacing) * scaleY)

			if virtualSpacingX < p.minSpacingX {
				virtualSpacingX = p.minSpacingX
			}
			if virtualSpacingY < p.minSpacingY {
				virtualSpacingY = p.minSpacingY
			}

			if i%virtualSpacingY == 0 && j%virtualSpacingX == 0 {
				grid[i][j] = gridCell{char: '·', color: GridDotColor}
			} else {
				grid[i][j] = gridCell{char: ' ', color: ""}
			}
		}
	}
	return grid
}

func (p *MonitorsPreviewPane) renderLegend() string {
	var legendItems []string

	for i, monitor := range p.monitors {
		if monitor.Disabled {
			continue
		}

		coloredPattern := GetMonitorColorStyle(i).Render("██")
		item := fmt.Sprintf("%s %s - %s, %s, %s, %s",
			coloredPattern, monitor.Name, monitor.Mode(), monitor.PositionPretty(),
			monitor.ScalePretty(false), monitor.RotationPretty(false))

		if i == p.selectedIndex {
			item = SelectedMonitorStyle.Render("► " + item + " ◄")
		}

		legendItems = append(legendItems, item)
	}

	if len(legendItems) == 0 {
		return "No active monitors"
	}

	return LegendStyle.Render(strings.Join(legendItems, "\n"))
}

func (p *MonitorsPreviewPane) ZoomIn() {
	p.virtualWidth = int(float64(p.virtualWidth) / p.zoomStep)
	p.virtualHeight = int(float64(p.virtualHeight) / p.zoomStep)
	// Set minimum bounds
	if p.virtualWidth < 1000 {
		p.virtualWidth = 1000
	}
	if p.virtualHeight < 750 {
		p.virtualHeight = 750
	}
}

func (p *MonitorsPreviewPane) ResetZoom() {
	p.virtualHeight = p.originalVirtualHeight
	p.virtualWidth = p.originalVirtualWidth
}

func (p *MonitorsPreviewPane) ZoomOut() {
	p.virtualWidth = int(float64(p.virtualWidth) * p.zoomStep)
	p.virtualHeight = int(float64(p.virtualHeight) * p.zoomStep)
	// Set maximum bounds
	if p.virtualWidth > 32000 {
		p.virtualWidth = 32000
	}
	if p.virtualHeight > 24000 {
		p.virtualHeight = 24000
	}
}

func (p *MonitorsPreviewPane) GetSelectedIndex() int {
	return p.selectedIndex
}

func (p *MonitorsPreviewPane) GetSnapGridX() *int {
	return p.snapGridX
}

func (p *MonitorsPreviewPane) GetSnapGridY() *int {
	return p.snapGridY
}

func (p *MonitorsPreviewPane) GetPanning() bool {
	return p.panning
}

func (p *MonitorsPreviewPane) GetSnapping() bool {
	return p.snapping
}

func (p *MonitorsPreviewPane) GetFollowMonitor() bool {
	return p.followMonitor
}

func (p *MonitorsPreviewPane) GetPanX() int {
	return p.panX
}

func (p *MonitorsPreviewPane) GetPanY() int {
	return p.panY
}

func (p *MonitorsPreviewPane) GetVirtualWidth() int {
	return p.virtualWidth
}

func (p *MonitorsPreviewPane) GetVirtualHeight() int {
	return p.virtualHeight
}
