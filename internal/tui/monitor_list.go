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
	inMirroringMode      bool
	inColorSelection     bool
}

func (m MonitorItem) Title() string {
	return m.monitor.Name
}

func (m MonitorItem) MonitorDescription(colors *ColorsManager) string {
	if m.monitor.Description == "" {
		return ""
	}

	return colors.ListItemSubtitle().Render(fmt.Sprintf("(%s)", m.monitor.Description))
}

func (m MonitorItem) MonitorDescriptionTrunc(colors *ColorsManager) string {
	if m.monitor.Description == "" {
		return ""
	}

	desc := m.monitor.Description
	if len(desc) > 15 {
		desc = desc[:15]
		desc += "..."
	}
	return colors.ListItemSubtitle().Render(fmt.Sprintf("(%s)", desc))
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
	m.inMirroringMode = false
}

func (m MonitorItem) Editing() bool {
	return m.isSelectedForEditing || m.inScaleMode || m.inModeSelection || m.inMirroringMode
}

func (m MonitorItem) Indicator(colors *ColorsManager) string {
	if !m.Editing() {
		return ""
	}

	if m.inModeSelection {
		return colors.MonitorModeSelectionMode().Render("[CHANGE MODE]")
	}

	if m.inScaleMode {
		return colors.MonitorScaleMode().Render("[SCALE MODE]")
	}

	if m.inColorSelection {
		return colors.MonitorColorMode().Render("[COLOR EDIT]")
	}

	if m.inMirroringMode {
		return colors.MonitorMirroringMode().Render("[MIRRORING]")
	}

	if m.isSelectedForEditing {
		return colors.MonitorEditingMode().Render("[EDITING]")
	}

	return ""
}

func (m MonitorItem) DescriptionLines() []string {
	return []string{
		m.monitor.StatusPretty(),
		m.monitor.ModePretty(),
		m.monitor.ScalePretty(true),
		m.monitor.VRRPretty(),
		m.monitor.RotationPretty(true),
		m.monitor.PositionPretty(),
		m.monitor.MirrorPretty(),
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
	color         key.Binding
	flip          key.Binding
	changeMode    key.Binding
	vrr           key.Binding
	toggle        key.Binding
	mirror        key.Binding
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
		vrr: key.NewBinding(
			key.WithKeys("v"),
			key.WithHelp("v", "toggle vrr"),
		),
		toggle: key.NewBinding(
			key.WithKeys("e"),
			key.WithHelp("e", "enable/disable"),
		),
		changeMode: key.NewBinding(
			key.WithKeys("m"),
			key.WithHelp("m", "change mode"),
		),
		mirror: key.NewBinding(
			key.WithKeys("i"),
			key.WithHelp("i", "mirror"),
		),
		color: key.NewBinding(
			key.WithKeys("C"),
			key.WithHelp("C", "color"),
		),
		flip: key.NewBinding(
			key.WithKeys("L"),
			key.WithHelp("L", "flip"),
		),
	}
}

func (m MonitorListKeyMap) ShortHelp(state AppState) []key.Binding {
	if state.EditingMonitor {
		return []key.Binding{
			m.unselect,
			m.rotate,
			m.scale,
			m.changeMode,
			m.vrr,
			m.toggle,
			m.mirror,
			m.color,
			m.flip,
		}
	}
	return []key.Binding{
		m.selectMonitor,
	}
}

type MonitorDelegate struct {
	keymap *MonitorListKeyMap
	width  int
	colors *ColorsManager
}

func NewMonitorDelegate(colors *ColorsManager) MonitorDelegate {
	return MonitorDelegate{
		keymap: NewMonitorListKeyMap(),
		colors: colors,
	}
}

func (d *MonitorDelegate) SetWidth(width int) {
	d.width = width
}

func (d MonitorDelegate) Height() int {
	return 9
}

func (d MonitorDelegate) Spacing() int {
	return 1
}

func (d MonitorDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd {
	var cmds []tea.Cmd
	logrus.Debug("Update called on MonitorDelegate")
	item, ok := m.SelectedItem().(MonitorItem)
	if !ok {
		logrus.Warning("Monitor delegate called with an item that is not a MonitorItem")
		return nil
	}
	logrus.Debugf("Selected item %v", item)

	sendMonitorSelection := false

	switch msg := msg.(type) {
	case ScaleMonitorCommand:
		logrus.Debugf("List called with scale cmd")
		if !item.Editing() {
			return nil
		}
		previous := item.inScaleMode
		item.RemoveSelectionModes()
		item.inScaleMode = !previous
		sendMonitorSelection = true
	case CloseColorPickerCommand, ChangeColorPresetFinalCommand:
		logrus.Debug("Received final change color command")
		if !item.Editing() {
			return nil
		}
		previous := item.inColorSelection
		item.RemoveSelectionModes()
		item.inColorSelection = !previous
		sendMonitorSelection = true
	case ChangeModeCommand, CloseMonitorModeListCommand:
		logrus.Debug("Received final change mode command")
		if !item.Editing() {
			return nil
		}
		previous := item.inModeSelection
		item.RemoveSelectionModes()
		item.inModeSelection = !previous
		sendMonitorSelection = true
	case ChangeMirrorCommand, CloseMonitorMirrorListCommand:
		logrus.Debug("Received final change mirror command")
		if !item.Editing() {
			return nil
		}
		previous := item.inMirroringMode
		item.RemoveSelectionModes()
		item.inMirroringMode = !previous
		sendMonitorSelection = true
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, d.keymap.selectMonitor), key.Matches(msg, d.keymap.unselect):
			logrus.Debugf("List called with selection")
			item.ToggleSelect()
			if item.Editing() {
				sendMonitorSelection = true
			} else {
				cmds = append(cmds, func() tea.Msg {
					return MonitorUnselected{}
				})
			}
		case key.Matches(msg, d.keymap.rotate):
			logrus.Debugf("List called with rotate")
			if !item.Editing() {
				return nil
			}
			cmds = append(cmds, rotateMonitorCmd(item.monitor))
		case key.Matches(msg, d.keymap.vrr):
			logrus.Debugf("List called with vrr")
			if !item.Editing() {
				return nil
			}
			cmds = append(cmds, toggleMonitorVRRCmd(item.monitor))
		case key.Matches(msg, d.keymap.toggle):
			logrus.Debugf("List called with toggle")
			if !item.Editing() {
				return nil
			}
			cmds = append(cmds, toggleMonitorCmd(item.monitor))
		case key.Matches(msg, d.keymap.flip):
			logrus.Debugf("List called with flip")
			if !item.Editing() {
				return nil
			}
			cmds = append(cmds, flipMonitorCmd(item.monitor))
		case key.Matches(msg, d.keymap.scale):
			logrus.Debugf("List called with scale")
			if !item.Editing() {
				return nil
			}
			previous := item.inScaleMode
			item.RemoveSelectionModes()
			item.inScaleMode = !previous
			sendMonitorSelection = true
		case key.Matches(msg, d.keymap.mirror):
			logrus.Debugf("List called with mirror")
			if !item.Editing() {
				return nil
			}
			previous := item.inMirroringMode
			item.RemoveSelectionModes()
			item.inMirroringMode = !previous
			logrus.Debugf("MirroringMode set to: %v", item.inMirroringMode)
			sendMonitorSelection = true
		case key.Matches(msg, d.keymap.changeMode):
			logrus.Debugf("List called with changeMode")
			if !item.Editing() {
				return nil
			}
			previous := item.inModeSelection
			item.RemoveSelectionModes()
			item.inModeSelection = !previous
			sendMonitorSelection = true
		case key.Matches(msg, d.keymap.color):
			logrus.Debugf("List called with color")
			if !item.Editing() {
				return nil
			}
			previous := item.inColorSelection
			item.RemoveSelectionModes()
			item.inColorSelection = !previous
			sendMonitorSelection = true
		}
	}
	cmds = append(cmds, m.SetItem(m.Index(), item))

	if sendMonitorSelection {
		cmds = append(cmds, func() tea.Msg {
			return MonitorBeingEdited{
				ListIndex:      m.Index(),
				Scaling:        item.inScaleMode,
				MonitorID:      *item.monitor.ID,
				ModesEditor:    item.inModeSelection,
				MirroringMode:  item.inMirroringMode,
				ColorSelection: item.inColorSelection,
			}
		})
	}

	return tea.Batch(cmds...)
}

func (d MonitorDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	logrus.Debug("Render on the monitor list called")
	monitor, ok := item.(MonitorItem)
	if !ok {
		return
	}

	var style lipgloss.Style
	var prefix string
	switch {
	case index == m.Index() && monitor.Editing():
		style = d.colors.MonitorEditingMode()
		prefix = "► "
	case index == m.Index():
		style = d.colors.ListItemSelected()
		prefix = "► "
	default:
		style = d.colors.ListItemUnselected()
	}
	title := fmt.Sprintf("%s%s %s %s", prefix, style.Render(monitor.Title()),
		monitor.MonitorDescriptionTrunc(d.colors), monitor.Indicator(d.colors))
	desc := d.colors.MutedStyle().Render(monitor.Description())
	content := fmt.Sprintf("%s\n%s", title, desc)

	fmt.Fprintf(w, "%s", content)
}

type monitorListKeyMap struct {
	Up    key.Binding
	Down  key.Binding
	Left  key.Binding
	Right key.Binding
}

func (m *monitorListKeyMap) Help(state AppState) []key.Binding {
	if !state.EditingMonitor {
		return []key.Binding{m.Left, m.Down, m.Up, m.Right}
	}
	if state.Scaling {
		return []key.Binding{m.Up, m.Down}
	}
	return []key.Binding{m.Left, m.Down, m.Up, m.Right}
}

type MonitorList struct {
	L                    list.Model
	state                AppState
	keys                 *monitorListKeyMap
	help                 *CustomHelp
	delegate             MonitorDelegate
	monitorSelected      bool
	selectedMonitorIndex int
	width                int
	colors               *ColorsManager
}

func NewMonitorList(monitors []*MonitorSpec, colors *ColorsManager) *MonitorList {
	monitorItems := make([]list.Item, len(monitors))
	for i, monitor := range monitors {
		monitorItems[i] = MonitorItem{monitor: monitor}
	}

	delegate := NewMonitorDelegate(colors)
	monitorsList := list.New(monitorItems, delegate, 0, 0)
	monitorsList.Title = "Connected Monitors"
	monitorsList.SetShowStatusBar(false)
	monitorsList.SetFilteringEnabled(false)
	monitorsList.SetShowHelp(false)
	monitorsList.SetShowTitle(false)

	return &MonitorList{
		L: monitorsList,
		keys: &monitorListKeyMap{
			Up:    rootKeyMap.Up,
			Down:  rootKeyMap.Down,
			Left:  rootKeyMap.Left,
			Right: rootKeyMap.Right,
		},
		help:     NewCustomHelp(colors),
		delegate: delegate,
		colors:   colors,
	}
}

func (c *MonitorList) Update(msg tea.Msg) tea.Cmd {
	logrus.Debugf("Update called on MonitorList: %v", msg)
	switch msg := msg.(type) {
	case MonitorBeingEdited:
		c.monitorSelected = true
		c.selectedMonitorIndex = msg.MonitorID
	case MonitorUnselected:
		c.monitorSelected = false
		c.selectedMonitorIndex = -1
	case StateChanged:
		c.state = msg.State
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, rootKeyMap.Down),
			key.Matches(msg, rootKeyMap.Up),
			key.Matches(msg, rootKeyMap.Left),
			key.Matches(msg, rootKeyMap.Right),
			key.Matches(msg, rootKeyMap.NextPage):
			if c.state.Panning {
				logrus.Debug("In panning mode, exiting")
				return nil
			}
			if c.state.Editing() {
				logrus.Debug("Arrows hijacked, not forwarding to the list")
				cmd := c.processArrows(msg)
				return cmd
			}
			logrus.Debug("Arrows not hijacked")
		}
	}

	var cmd tea.Cmd
	c.L, cmd = c.L.Update(msg)
	return cmd
}

func (c *MonitorList) processArrows(msg tea.KeyMsg) tea.Cmd {
	logrus.Debug("Processing arrows for list updates")
	if !c.state.EditingMonitor || c.selectedMonitorIndex == -1 {
		logrus.Debug("Not in editing mode, exit")
		return nil
	}
	return c.handleMove(msg)
}

func (c *MonitorList) handleMove(msg tea.KeyMsg) tea.Cmd {
	stepX := DeltaNone
	stepY := DeltaNone

	switch {
	case key.Matches(msg, rootKeyMap.Up):
		stepY = DeltaLess
	case key.Matches(msg, rootKeyMap.Down):
		stepY = DeltaMore
	case key.Matches(msg, rootKeyMap.Left):
		stepX = DeltaLess
	case key.Matches(msg, rootKeyMap.Right):
		stepX = DeltaMore
	}

	return MoveMonitorCmd(c.selectedMonitorIndex, stepX, stepY)
}

func (c *MonitorList) SetHeight(height int) {
	c.L.SetHeight(height)
}

func (c *MonitorList) SetWidth(width int) {
	c.L.SetWidth(width)
	c.width = width
	c.delegate.SetWidth(width)
}

func (c *MonitorList) View() string {
	var (
		sections    []string
		availHeight = c.L.Height()
	)

	title := c.colors.TitleStyle().Margin(0, 0, 1, 0).Render("Connected Monitors")
	help := c.ShortHelp()
	availHeight -= lipgloss.Height(help)
	availHeight -= lipgloss.Height(title)
	logrus.Debugf("Help height: %d", lipgloss.Height(help))
	c.L.SetHeight(availHeight)
	content := lipgloss.NewStyle().Height(availHeight).Width(c.width).Render(c.L.View())

	sections = append(sections, title)
	sections = append(sections, content)
	sections = append(sections, help)

	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

func (c *MonitorList) ShortHelp() string {
	sections := []string{}

	listHelp := c.colors.HelpStyle().Width(c.width).Render(
		c.help.ShortHelpView(c.keys.Help(c.state)))
	sections = append(sections, listHelp)

	delegateHelp := c.colors.HelpStyle().Width(c.width).Render(c.help.ShortHelpView(
		c.delegate.keymap.ShortHelp(c.state)))
	sections = append(sections, delegateHelp)

	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}
