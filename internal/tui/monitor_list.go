package tui

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/fiffeek/hyprdynamicmonitors/internal/hypr"
	"github.com/sirupsen/logrus"
)

type MonitorItem struct {
	monitor              *MonitorSpec
	isSelectedForEditing bool
	inScaleMode          bool
	inModeSelection      bool
}

func (m MonitorItem) Title() string {
	return m.monitor.Name
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
	selectMonitor key.Binding
	unselect      key.Binding
	rotate        key.Binding
	scale         key.Binding
	changeMode    key.Binding
}

func NewMonitorListKeyMap() *MonitorListKeyMap {
	return &MonitorListKeyMap{
		selectMonitor: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "edit a monitor"),
		),
		unselect: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "unselect a monitor"),
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

func (m MonitorListKeyMap) ShortHelp(selected bool) []key.Binding {
	if selected {
		return []key.Binding{
			m.unselect,
			m.rotate,
			m.scale,
			m.changeMode,
		}
	}
	return []key.Binding{
		m.selectMonitor,
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

	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, d.keymap.selectMonitor), key.Matches(msg, d.keymap.unselect):
			logrus.Debugf("List called with selection")
			item.ToggleSelect()
			if item.Editing() {
				cmd = func() tea.Msg {
					return MonitorSelected{
						Index: m.Index(),
					}
				}
			} else {
				cmd = func() tea.Msg {
					return MonitorUnselected{}
				}
			}
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

	return cmd
}

func (d MonitorDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	logrus.Debug("Render on the monitor list called")
	monitor, ok := item.(MonitorItem)
	if !ok {
		return
	}

	var style lipgloss.Style
	switch {
	case index == m.Index() && monitor.Editing():
		style = MonitorEditingMode
	case index == m.Index():
		style = MonitorListSelected
	default:
		style = MonitorListTitle
	}
	title := fmt.Sprintf("%s %s%s", style.Render(monitor.Title()), monitor.MonitorDescription(), monitor.Indicator())
	desc := MutedStyle.Render(monitor.Description())
	content := fmt.Sprintf("%s\n%s", title, desc)

	fmt.Fprintf(w, "%s", content)
}

type monitorListKeyMap struct {
	Up    key.Binding
	Down  key.Binding
	Left  key.Binding
	Right key.Binding
}

func (m *monitorListKeyMap) Help() []key.Binding {
	return []key.Binding{m.Left, m.Down, m.Up, m.Right}
}

type MonitorList struct {
	L               list.Model
	hijackArrows    bool
	keys            *monitorListKeyMap
	help            help.Model
	delegate        MonitorDelegate
	monitorSelected bool
}

func NewMonitorList(monitors hypr.MonitorSpecs) *MonitorList {
	monitorItems := make([]list.Item, len(monitors))
	for i, monitor := range monitors {
		monitorItems[i] = MonitorItem{monitor: NewMonitorSpec(monitor)}
	}

	delegate := NewMonitorDelegate()
	monitorsList := list.New(monitorItems, delegate, 0, 0)
	monitorsList.Title = "Connected Monitors"
	monitorsList.SetShowStatusBar(false)
	monitorsList.SetFilteringEnabled(false)
	monitorsList.SetShowHelp(false)

	monitorsList.SetShowHelp(false)

	return &MonitorList{
		L:            monitorsList,
		hijackArrows: false,
		keys: &monitorListKeyMap{
			Up:    rootKeyMap.Up,
			Down:  rootKeyMap.Down,
			Left:  rootKeyMap.Left,
			Right: rootKeyMap.Right,
		},
		help:     help.New(),
		delegate: delegate,
	}
}

func (c *MonitorList) SetHijackArrows(v bool) {
	logrus.Debugf("Setting hijacked arrows: %v", v)
	c.hijackArrows = v
}

func (c *MonitorList) Update(msg tea.Msg) tea.Cmd {
	logrus.Debugf("Update called on MonitorList: %v", msg)
	switch msg := msg.(type) {
	case MonitorSelected:
		c.monitorSelected = true
	case MonitorUnselected:
		c.monitorSelected = false
	case WindowResized:
		logrus.Debug("Window resized")
		c.L.SetWidth(msg.width)
		c.L.SetHeight(msg.height)
		return nil
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "down", "left", "right":
			if c.hijackArrows {
				logrus.Debug("Arrows hijacked, not forwarding")
				return nil
			}
			logrus.Debug("Arrows not hijacked")
		}
	}

	var cmd tea.Cmd
	c.L, cmd = c.L.Update(msg)
	return cmd
}

func (c *MonitorList) View() string {
	var (
		sections    []string
		availHeight = c.L.Height()
	)

	help := c.ShortHelp()
	availHeight -= lipgloss.Height(help)
	content := lipgloss.NewStyle().Height(availHeight).Render(c.L.View())

	sections = append(sections, content)
	sections = append(sections, help)

	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

func (c *MonitorList) ShortHelp() string {
	listHelp := HelpStyle.Render(c.help.ShortHelpView(c.keys.Help()))
	delegateHelp := HelpStyle.Render(c.help.ShortHelpView(c.delegate.keymap.ShortHelp(c.monitorSelected)))
	return lipgloss.JoinVertical(lipgloss.Left, listHelp, delegateHelp)
}
