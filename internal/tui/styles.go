package tui

import "github.com/charmbracelet/lipgloss"

var (
	ActiveStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("62"))

	InactiveStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("240"))
	HeaderStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("205"))
	HeaderIndicatorStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("226")).
				Background(lipgloss.Color("235"))
	ErrorStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("196")).
			Background(lipgloss.Color("235"))
	SuccessStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("82")).
			Background(lipgloss.Color("235"))
	TabActiveStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("255")).
			Background(lipgloss.Color("105")).
			Padding(0, 1)
	TabInactiveStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("243")).
				Background(lipgloss.Color("235")).
				Padding(0, 1)
)

var (
	HyprConfigTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("180"))

	HyprCommandStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("255")).
				Background(lipgloss.Color("235"))
)

var (
	ItemSubtitle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("180")).
			Italic(true)
	MonitorColorMode = lipgloss.NewStyle().
				Foreground(lipgloss.Color("15")).
				Bold(true)
	MonitorScaleMode = lipgloss.NewStyle().
				Foreground(lipgloss.Color("39")).
				Bold(true)
	MonitorMirroringMode = lipgloss.NewStyle().
				Foreground(lipgloss.Color("140")).
				Bold(true)
	MonitorEditingMode = lipgloss.NewStyle().
				Foreground(lipgloss.Color("226")).
				Bold(true)
	MonitorModeSelectionMode = lipgloss.NewStyle().
					Foreground(lipgloss.Color("208")).
					Bold(true)
	MutedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("243"))
	MonitorListTitle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("255"))
	MonitorListSelected = lipgloss.NewStyle().
				Foreground(lipgloss.Color("212")).
				Bold(true)
)

var (
	LegendStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("243"))

	SelectedMonitorStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("226")).
				Background(lipgloss.Color("235"))
)

var (
	HelpStyle  = lipgloss.NewStyle().Padding(0, 0, 0, 2)
	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("105"))
	SubtitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("180"))
	SubtitleInfoStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("108")).
				Italic(true)
	LinkStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Italic(true)
)

var GridDotColor = "240" // Grey color for grid dots

var MonitorEdgeColors = []string{"105", "208", "39", "226", "196", "99"}

func GetMonitorColorStyle(index int) lipgloss.Style {
	color := MonitorEdgeColors[index%len(MonitorEdgeColors)]
	return lipgloss.NewStyle().Foreground(lipgloss.Color(color))
}

func GetMonitorFillForEdge(color string, selected bool) string {
	if !selected {
		switch color {
		case "105":
			return "99" // Bright purple -> darker purple
		case "208":
			return "166" // Bright orange -> darker orange
		case "39":
			return "25" // Bright blue -> darker blue
		case "226":
			return "220" // Bright yellow -> darker golden yellow
		case "196":
			return "124" // Bright red -> darker red
		case "99":
			return "60" // Bright pink -> darker magenta
		default:
			return "250" // Light gray fallback instead of white
		}
	}

	switch color {
	case "105":
		return "98" // Bright purple -> darker purple
	case "208":
		return "94" // Bright orange -> much darker orange/brown
	case "39":
		return "17" // Bright blue -> much darker blue
	case "226":
		return "100" // Bright yellow -> much darker gold
	case "196":
		return "52" // Bright red -> much darker red/maroon
	case "99":
		return "53" // Bright pink -> much darker magenta/purple
	default:
		return "237" // Much darker gray fallback
	}
}

func GetMonitorBottomColor(color string) string {
	switch color {
	case "105":
		return "141" // Purple -> Bright purple
	case "208":
		return "214" // Orange -> Bright orange
	case "39":
		return "45" // Blue -> Bright blue
	case "226":
		return "228" // Yellow -> Bright yellow
	case "196":
		return "197" // Red -> Bright red
	case "99":
		return "105" // Pink -> Bright pink
	default:
		return "255" // White as fallback
	}
}

var (
	configPaneBorderStyle = lipgloss.NewStyle()

	configPaneWarningStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("196")).
				Italic(true)

	configPaneActionStyle = lipgloss.NewStyle().
				Background(lipgloss.Color("235"))

	configDetailStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("180"))

	configDetailLabelStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("105")).
				Bold(true)
)

var (
	// Monitor mode list styles
	StandardModeSuffixStyle = lipgloss.NewStyle().
				Italic(true).
				Foreground(lipgloss.Color("120"))

	NonStandardModeSuffixStyle = lipgloss.NewStyle().
					Italic(true).
					Foreground(lipgloss.Color("243"))

	RefreshRateBoldStyle = lipgloss.NewStyle().
				Bold(true)

	// Scale selector styles
	MutedItalicStyle = lipgloss.NewStyle().
				Italic(true).
				Foreground(lipgloss.Color("243"))

	CustomInputLabelStyle = lipgloss.NewStyle().
				Margin(1, 0, 0, 0)
)
