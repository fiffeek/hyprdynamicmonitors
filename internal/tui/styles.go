package tui

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/fiffeek/hyprdynamicmonitors/internal/config"
)

type ColorsManager struct {
	cfg *config.Config
}

func NewColorsManager(cfg *config.Config) *ColorsManager {
	return &ColorsManager{
		cfg: cfg,
	}
}

func (c *ColorsManager) colors() *config.TUIColors {
	return c.cfg.Get().TUISection.Colors
}

// HelpKeyStyle returns the style for help key text.
func (c *ColorsManager) HelpKeyStyle() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(lipgloss.Color(*c.colors().HelpKeyColor))
}

// HelpDescriptionStyle returns the style for help description text.
func (c *ColorsManager) HelpDescriptionStyle() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(lipgloss.Color(*c.colors().HelpDescriptionColor))
}

// HelpSeparatorStyle returns the style for help separator text.
func (c *ColorsManager) HelpSeparatorStyle() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(lipgloss.Color(*c.colors().HelpSeparatorColor))
}

// ActiveStyle returns the style for active panes.
func (c *ColorsManager) ActiveStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color(*c.colors().ActivePaneColor))
}

// InactiveStyle returns the style for inactive panes.
func (c *ColorsManager) InactiveStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color(*c.colors().InactivePaneColor))
}

// ProgramNameStyle returns the style for the program name.
func (c *ColorsManager) ProgramNameStyle() lipgloss.Style {
	return lipgloss.NewStyle().Bold(true).Foreground(
		lipgloss.Color(*c.colors().ProgramNameColor)).Background(
		lipgloss.Color(*c.colors().ProgramNameBg))
}

// HeaderStyle returns the style for headers.
func (c *ColorsManager) HeaderStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color(*c.colors().HeaderColor))
}

// HeaderIndicatorStyle returns the style for header indicators.
func (c *ColorsManager) HeaderIndicatorStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color(*c.colors().HeaderIndicatorColor)).
		Background(lipgloss.Color(*c.colors().HeaderIndicatorBg))
}

// ErrorStyle returns the style for error status messages.
func (c *ColorsManager) ErrorStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color(*c.colors().ErrorColor)).
		Background(lipgloss.Color(*c.colors().ErrorBg))
}

// SuccessStyle returns the style for success status messages.
func (c *ColorsManager) SuccessStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color(*c.colors().SuccessColor)).
		Background(lipgloss.Color(*c.colors().SuccessBg))
}

// TabActiveStyle returns the style for active tabs.
func (c *ColorsManager) TabActiveStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color(*c.colors().TabActiveColor)).
		Background(lipgloss.Color(*c.colors().TabActiveBg)).
		Padding(0, 1)
}

// TabInactiveStyle returns the style for inactive tabs.
func (c *ColorsManager) TabInactiveStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color(*c.colors().TabInactiveColor)).
		Background(lipgloss.Color(*c.colors().TabInactiveBg)).
		Padding(0, 1)
}

// ListItemSubtitle returns the style for item subtitles.
func (c *ColorsManager) ListItemSubtitle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color(*c.colors().ListItemSubtitleColor)).
		Italic(true)
}

// MonitorColorMode returns the style for monitor color mode.
func (c *ColorsManager) MonitorColorMode() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color(*c.colors().MonitorColorModeColor)).
		Bold(true)
}

// MonitorScaleMode returns the style for monitor scale mode.
func (c *ColorsManager) MonitorScaleMode() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color(*c.colors().MonitorScaleModeColor)).
		Bold(true)
}

// MonitorMirroringMode returns the style for monitor mirroring mode.
func (c *ColorsManager) MonitorMirroringMode() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color(*c.colors().MonitorMirroringModeColor)).
		Bold(true)
}

// MonitorEditingMode returns the style for monitor editing mode.
func (c *ColorsManager) MonitorEditingMode() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color(*c.colors().MonitorEditingModeColor)).
		Bold(true)
}

// MonitorModeSelectionMode returns the style for monitor mode selection.
func (c *ColorsManager) MonitorModeSelectionMode() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color(*c.colors().MonitorModeSelectionColor)).
		Bold(true)
}

// MutedStyle returns the style for muted text.
func (c *ColorsManager) MutedStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color(*c.colors().MutedColor))
}

// ListItemUnselected returns the style for lists.
func (c *ColorsManager) ListItemUnselected() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color(*c.colors().ListItemUnselectedColor))
}

// ListItemSelected returns the style for selected items in lists.
func (c *ColorsManager) ListItemSelected() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color(*c.colors().ListItemSelectedColor)).
		Bold(true)
}

// LegendStyle returns the style for legends.
func (c *ColorsManager) LegendStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color(*c.colors().LegendColor))
}

// LegendSelectedMonitorStyle returns the style for the selected monitor.
func (c *ColorsManager) LegendSelectedMonitorStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color(*c.colors().LegendSelectedMonitorColor)).
		Background(lipgloss.Color(*c.colors().LegendSelectedMonitorBg))
}

// HelpStyle returns the style for help text.
func (c *ColorsManager) HelpStyle() lipgloss.Style {
	return lipgloss.NewStyle().Padding(0, 0, 0, 0)
}

// TitleStyle returns the style for titles.
func (c *ColorsManager) TitleStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color(*c.colors().TitleColor))
}

// SubtitleStyle returns the style for subtitles.
func (c *ColorsManager) SubtitleStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color(*c.colors().SubtitleColor))
}

// InfoStyle returns the style for subtitle info text.
func (c *ColorsManager) InfoStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color(*c.colors().InfoColor))
}

// LinkStyle returns the style for links.
func (c *ColorsManager) LinkStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color(*c.colors().LinkColor)).
		Italic(true)
}

// GridDotColor returns the color for grid dots.
func (c *ColorsManager) GridDotColor() string {
	return *c.colors().GridDotColor
}

func (c *ColorsManager) GridSnappingLineColor() string {
	return *c.colors().GridSnappingLineColor
}

// MonitorEdgeColors returns the colors for monitor edges.
func (c *ColorsManager) MonitorEdgeColors() []string {
	return *c.colors().MonitorEdgeColors
}

// GetMonitorColorStyle returns the style for a monitor color by index.
func (c *ColorsManager) GetMonitorColorStyle(index int) lipgloss.Style {
	colors := c.MonitorEdgeColors()
	color := colors[index%len(colors)]
	return lipgloss.NewStyle().Foreground(lipgloss.Color(color))
}

// ConfigPaneBorderStyle returns the style for config pane borders.
func (c *ColorsManager) ConfigPaneBorderStyle() lipgloss.Style {
	return lipgloss.NewStyle()
}

// WarningStyle returns the style for warnings
func (c *ColorsManager) WarningStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color(*c.colors().WarningColor)).
		Italic(true)
}

func (c *ColorsManager) MonitorPreviewPaneLabelBackgroundColor() string {
	return *c.colors().MonitorPreviewPaneLabelBackgroundColor
}

// RefreshRateBoldStyle returns the style for bold refresh rates.
func (c *ColorsManager) RefreshRateBoldStyle() lipgloss.Style {
	return lipgloss.NewStyle().Bold(true)
}

// MutedItalicStyle returns the style for muted italic text.
func (c *ColorsManager) MutedItalicStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Italic(true).
		Foreground(lipgloss.Color(*c.colors().MutedColor))
}

// CustomInputLabelStyle returns the style for custom input labels.
func (c *ColorsManager) CustomInputLabelStyle() lipgloss.Style {
	return lipgloss.NewStyle().Margin(1, 0, 0, 0)
}

// GetMonitorFillForEdge returns the fill color for a monitor edge.
func (c *ColorsManager) GetMonitorFillForEdge(color string, selected bool) string {
	colors := c.colors()
	edgeColors := *colors.MonitorEdgeColors

	// Find the index of the edge color
	for i, edgeColor := range edgeColors {
		if edgeColor == color {
			if !selected {
				fillColors := *colors.MonitorFillColorsUnselected
				if i < len(fillColors) {
					return fillColors[i]
				}
			} else {
				fillColors := *colors.MonitorFillColorsSelected
				if i < len(fillColors) {
					return fillColors[i]
				}
			}
		}
	}

	// Fallback if color not found in edge colors
	if !selected {
		return "250" // Light gray fallback
	}
	return "237" // Much darker gray fallback for selected
}

// GetMonitorBottomColor returns the bottom color for a monitor.
func (c *ColorsManager) GetMonitorBottomColor(color string) string {
	colors := c.colors()
	edgeColors := *colors.MonitorEdgeColors
	bottomColors := *colors.MonitorBottomColors

	// Find the index of the edge color
	for i, edgeColor := range edgeColors {
		if edgeColor == color {
			if i < len(bottomColors) {
				return bottomColors[i]
			}
		}
	}

	// Fallback if color not found
	return "255" // White as fallback
}
