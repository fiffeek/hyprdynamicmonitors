package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/sirupsen/logrus"
)

type gridCell struct {
	char  rune
	color string
}

type MonitorsPreviewPane struct {
	monitors      []*MonitorSpec
	selectedIndex int
	panX, panY    int
	virtualWidth  int
	virtualHeight int
	width         int
	height        int
	baseSpacing   int
	minSpacingY   int
	minSpacingX   int
}

func NewMonitorsPreviewPane(monitors []*MonitorSpec) *MonitorsPreviewPane {
	return &MonitorsPreviewPane{
		monitors:      monitors,
		selectedIndex: -1,
		panX:          0,
		panY:          0,
		virtualWidth:  8000,
		virtualHeight: 8000,
		baseSpacing:   200,
		minSpacingY:   2,
		minSpacingX:   4,
	}
}

func (p *MonitorsPreviewPane) Update(msg tea.Msg) tea.Cmd {
	logrus.Debugf("Update called on MonitorsPreviewPane: %v", msg)
	switch msg := msg.(type) {
	case MonitorSelected:
		p.selectedIndex = msg.Index
	case MonitorUnselected:
		p.selectedIndex = -1
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
			ScaleInfoStyle.Render(scaleInfo),
		),
		grid,
		legend,
	)
}

func (p *MonitorsPreviewPane) ScaleInfo() string {
	scaleInfo := fmt.Sprintf("Virtual Area: %dx%d", p.virtualWidth, p.virtualHeight)
	if p.panX != 0 || p.panY != 0 {
		scaleInfo += fmt.Sprintf(" | Pan: (%d,%d)", p.panX, p.panY)
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

	lines := p.ColorGrid(grid)
	return strings.Join(lines, "\n")
}

// ColorGrid converts the grid cells to the colorized representation
func (*MonitorsPreviewPane) ColorGrid(grid [][]gridCell) []string {
	var lines []string
	for _, row := range grid {
		var line strings.Builder
		var currentColor string
		var currentSegment strings.Builder

		for _, cell := range row {
			if cell.color != currentColor {
				// Flush previous segment with its color
				if currentSegment.Len() > 0 {
					if currentColor != "" {
						style := lipgloss.NewStyle().Foreground(lipgloss.Color(currentColor))
						line.WriteString(style.Render(currentSegment.String()))
					} else {
						line.WriteString(currentSegment.String())
					}
					currentSegment.Reset()
				}
				currentColor = cell.color
			}
			currentSegment.WriteRune(cell.char)
		}

		// Flush remaining segment
		if currentSegment.Len() > 0 {
			if currentColor != "" {
				style := lipgloss.NewStyle().Foreground(lipgloss.Color(currentColor))
				line.WriteString(style.Render(currentSegment.String()))
			} else {
				line.WriteString(currentSegment.String())
			}
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

	color := MonitorColors[i%len(MonitorColors)]
	isSelected := i == p.selectedIndex
	p.drawMonitorRectangle(isSelected, rectangle, color, grid)
	p.drawMonitorLabel(monitor, rectangle, grid, color)
}

func (*MonitorsPreviewPane) drawMonitorLabel(monitor *MonitorSpec, rectangle *MonitorRectangle, grid [][]gridCell, color string) {
	gridWidth := len(grid[0])
	gridHeight := len(grid)

	label := monitor.Name
	if len(label) > 6 {
		label = label[:6] // Truncate long names
	}

	// Determine positioning arrow based on monitor's relative position
	arrow := monitor.PositionArrowView()
	fullLabel := label + arrow

	labelY := (rectangle.startY + rectangle.endY) / 2
	labelStartX := (rectangle.startX + rectangle.endX - len(fullLabel)) / 2

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
	color string, grid [][]gridCell,
) {
	gridWidth := len(grid[0])
	gridHeight := len(grid)

	borderChar := '█'
	fillChar := '▓'
	if isSelected {
		borderChar = '▓'
		fillChar = '▒'
	}

	for y := rectangle.startY; y <= rectangle.endY; y++ {
		for x := rectangle.startX; x <= rectangle.endX; x++ {
			if y >= 0 && y < gridHeight && x >= 0 && x < gridWidth {
				// check bounds
				if y == rectangle.startY || y == rectangle.endY ||
					x == rectangle.startX || x == rectangle.endX {
					// Use different color for the "bottom" edge based on rotation
					edgeColor := color

					if rectangle.isBottomEdge(x, y) {
						// Use a brighter/different shade for the bottom edge
						edgeColor = GetBrightMonitorColor(color)
					}

					grid[y][x] = gridCell{char: borderChar, color: edgeColor}
					continue
				}

				grid[y][x] = gridCell{char: fillChar, color: color}
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

		coloredPattern := GetMonitorColorStyle(i).Render("▓▓")
		item := fmt.Sprintf("%s %s - %s, %s, %s, %s",
			coloredPattern, monitor.Name, monitor.Mode(), monitor.PositionPretty(),
			monitor.ScalePretty(), monitor.RotationPretty())

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
