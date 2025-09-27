package tui

import "github.com/charmbracelet/lipgloss"

var (
	ActiveStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("62"))

	InactiveStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("240"))
)

var (
	ItemSubtitle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("180")).
			Italic(true)
	MonitorScaleMode = lipgloss.NewStyle().
				Foreground(lipgloss.Color("39")).
				Bold(true)
	MonitorEditingMode = lipgloss.NewStyle().
				Foreground(lipgloss.Color("226")).
				Bold(true)
	MonitorModeSelectionMode = lipgloss.NewStyle().
					Foreground(lipgloss.Color("208")).
					Bold(true)
)
