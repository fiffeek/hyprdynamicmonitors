package tui

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/sirupsen/logrus"
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

func (m *MonitorItem) Unselect() {
	m.inScaleMode = false
	m.inModeSelection = false
	m.isSelectedForEditing = false
}

func (m *MonitorItem) ToggleSelect() {
	if m.isSelectedForEditing {
		m.Unselect()
		return
	}
	m.isSelectedForEditing = true
}

func (m *MonitorItem) RemoveSelectionModes() {
	m.inScaleMode = false
	m.inModeSelection = false
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

func (m MonitorItem) DescriptionLines() []string {
	return []string{
		m.monitor.StatusPretty(),
		m.monitor.Mode(),
		m.monitor.ScalePretty(),
		m.monitor.VRRPretty(),
		m.monitor.RotationPretty(),
		m.monitor.PositionPretty(),
	}
}

func (m MonitorItem) Description() string {
	return strings.Join(m.DescriptionLines(), "\n")
}

func (m MonitorItem) FilterValue() string {
	return m.monitor.Name + " " + m.monitor.Description
}

type MonitorListKeyMap struct {
	choose     key.Binding
	rotate     key.Binding
	scale      key.Binding
	changeMode key.Binding
}

func NewMonitorListKeyMap() *MonitorListKeyMap {
	return &MonitorListKeyMap{
		choose: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "edit monitor"),
		),
		rotate: key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "rotate"),
		),
		scale: key.NewBinding(
			key.WithKeys("s"),
			key.WithHelp("s", "scale"),
		),
		changeMode: key.NewBinding(
			key.WithKeys("m"),
			key.WithHelp("m", "change mode"),
		),
	}
}

func (m MonitorListKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{
		m.choose,
	}
}

func (m MonitorListKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{
			m.choose,
			m.rotate,
			m.scale,
			m.changeMode,
		},
	}
}

type MonitorDelegate struct {
	keymap *MonitorListKeyMap
}

func NewMonitorDelegate() MonitorDelegate {
	return MonitorDelegate{
		keymap: NewMonitorListKeyMap(),
	}
}

func (d MonitorDelegate) Height() int {
	return 8
}

func (d MonitorDelegate) Spacing() int {
	return 1
}

func (d MonitorDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd {
	logrus.Debug("Update called on MonitorDelegate")
	item, ok := m.SelectedItem().(MonitorItem)
	if !ok {
		logrus.Warning("Monitor delegate called with an item that is not a MonitorItem")
		return nil
	}
	logrus.Debugf("Selected item %v", item)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, d.keymap.choose):
			logrus.Debugf("List called with choose")
			item.ToggleSelect()
		case key.Matches(msg, d.keymap.scale):
			logrus.Debugf("List called with scale")
			if !item.Editing() {
				return nil
			}
			previous := item.inScaleMode
			item.RemoveSelectionModes()
			item.inScaleMode = !previous
		case key.Matches(msg, d.keymap.changeMode):
			logrus.Debugf("List called with changeMode")
			if !item.Editing() {
				return nil
			}
			previous := item.inModeSelection
			item.RemoveSelectionModes()
			item.inModeSelection = !previous
		}
	}
	m.SetItem(m.Index(), item)

	return nil
}

func (d MonitorDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	logrus.Debug("Render on the monitor list called")
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

func (d MonitorDelegate) ShortHelp() []key.Binding {
	return []key.Binding{d.keymap.choose}
}

func (d MonitorDelegate) FullHelp() []key.Binding {
	return []key.Binding{d.keymap.choose}
}
