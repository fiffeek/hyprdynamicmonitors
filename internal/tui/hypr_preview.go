package tui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

type HyprPreviewPane struct {
	monitors []*MonitorSpec
}

func NewHyprPreviewPane(monitors []*MonitorSpec) *HyprPreviewPane {
	return &HyprPreviewPane{
		monitors,
	}
}

func (h *HyprPreviewPane) View() string {
	var sections []string

	if len(h.monitors) == 0 {
		sections = append(sections, HyprConfigTitleStyle.Render(
			"Hyprland Config Preview\n\nNo monitors to configure"))
	} else {
		sections = append(sections, HyprConfigTitleStyle.Render("Hyprland Config Preview"))
	}

	for _, monitor := range h.monitors {
		hyprCommand := fmt.Sprintf("monitor = %s", monitor.ToHypr())
		sections = append(sections, HyprCommandStyle.Render(hyprCommand))
	}
	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}
