package tui

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type MonitorItem struct {
	monitor              *MonitorSpec
	isSelectedForEditing bool
	inScaleMode          bool
	inModeSelection      bool
}

func (m MonitorItem) Title() string {
	return fmt.Sprintf("%s %s%s", m.monitor.Name, m.MonitorDescription(), m.Indicator())
}

func (m MonitorItem) MonitorDescription() string {
	if m.monitor.Description == "" {
		return ""
	}

	return ItemSubtitle.Render(fmt.Sprintf("(%s)", m.monitor.Description))
}

func (m MonitorItem) Editing() bool {
	return m.isSelectedForEditing || m.inScaleMode || m.inModeSelection
}

func (m MonitorItem) Indicator() string {
	if !m.Editing() {
		return ""
	}

	if m.inModeSelection {
		return MonitorModeSelectionMode.Render(" [CHANGE MODE]")
	}

	if m.inScaleMode {
		return MonitorScaleMode.Render(" [SCALE MODE]")
	}

	if m.isSelectedForEditing {
		return MonitorEditingMode.Render(" [EDITING]")
	}

	return ""
}

func (m MonitorItem) Description() string {
	lines := []string{
		m.monitor.StatusPretty(),
		m.monitor.Mode(),
		m.monitor.ScalePretty(),
		m.monitor.VRRPretty(),
		m.monitor.RotationPretty(),
		m.monitor.PositionPretty(),
	}
	return strings.Join(lines, "\n")
}

func (m MonitorItem) FilterValue() string {
	return m.monitor.Name + " " + m.monitor.Description
}

type MonitorDelegate struct{}

func NewMonitorDelegate() MonitorDelegate {
	return MonitorDelegate{}
}

func (d MonitorDelegate) Height() int {
	return 8
}

func (d MonitorDelegate) Spacing() int {
	return 1
}

func (d MonitorDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd {
	return nil
}

func (d MonitorDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	monitor, ok := item.(MonitorItem)
	if !ok {
		return
	}

	// Styles
	selectedStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("212")).
		Bold(true)

	normalStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("255"))

	descStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("243"))

	var style lipgloss.Style
	if index == m.Index() {
		style = selectedStyle
	} else {
		style = normalStyle
	}

	title := style.Render(monitor.Title())
	desc := descStyle.Render(monitor.Description())
	content := fmt.Sprintf("%s\n%s", title, desc)

	fmt.Fprintf(w, "%s", content)
}
