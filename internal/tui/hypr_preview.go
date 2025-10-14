package tui

import (
	"github.com/charmbracelet/lipgloss"
)

type HyprPreviewPane struct {
	monitors []*MonitorSpec
	clamp    bool
	width    int
}

func NewHyprPreviewPane(monitors []*MonitorSpec) *HyprPreviewPane {
	return &HyprPreviewPane{
		monitors: monitors,
		width:    0,
		clamp:    true,
	}
}

func (h *HyprPreviewPane) SetWidth(width int) {
	h.width = width
}

func (h *HyprPreviewPane) SetClamp(clamp bool) {
	h.clamp = clamp
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
		hyprCommand := "monitor = " + monitor.ToHypr()
		if h.clamp {
			hyprCommand = ClampTextTo(hyprCommand, h.width)
		}
		sections = append(sections, HyprCommandStyle.Render(hyprCommand))
	}
	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

func ClampTextTo(text string, maxLen int) string {
	if maxLen < 0 {
		return text
	}
	if len(text) <= maxLen {
		return text
	}
	dots := "..."
	lenWithoutDots := maxLen - len(dots)
	return dots + text[len(text)-lenWithoutDots:]
}
