package tui

import (
	"fmt"
	"io"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/sirupsen/logrus"
)

type MirrorItem struct {
	mirrorName string
}

type MirrorList struct {
	L        list.Model
	monitors []*MonitorSpec
	help     *CustomHelp
	colors   *ColorsManager
}

func (m MirrorItem) FilterValue() string {
	return m.mirrorName
}

func (m MirrorItem) View() string {
	return m.mirrorName
}

type MirrorDelegate struct {
	colors *ColorsManager
}

func NewMirrorDelegate(colors *ColorsManager) MirrorDelegate {
	return MirrorDelegate{
		colors: colors,
	}
}

func (d MirrorDelegate) Height() int {
	return 1
}

func (d MirrorDelegate) Spacing() int {
	return 0
}

func (d MirrorDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd {
	cmds := []tea.Cmd{}
	item, ok := m.SelectedItem().(MirrorItem)
	if !ok {
		logrus.Warning("Mirror delegate called with an item that is not a MirrorItem")
		return nil
	}
	// nolint:gocritic
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k", "down", "j":
			logrus.Debugf("Setting mode to: %s", item.mirrorName)
			cmds = append(cmds, ChangeMirrorPreviewCmd(item.mirrorName))
		case "enter":
			logrus.Debugf("Setting final to: %s", item.mirrorName)
			cmds = append(cmds, ChangeMirrorCmd(item.mirrorName))
		}
	}
	return tea.Batch(cmds...)
}

func (d MirrorDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	modeItem, ok := item.(MirrorItem)
	if !ok {
		return
	}

	var style lipgloss.Style
	var prefix string
	switch {
	case index == m.Index():
		style = d.colors.ListItemSelected()
		prefix = "► "
	default:
		style = d.colors.ListItemUnselected()
	}
	title := style.Render(prefix + modeItem.View())
	content := title

	fmt.Fprintf(w, "%s", content)
}

func NewMirrorList(monitors []*MonitorSpec, colors *ColorsManager) *MirrorList {
	modesItems := []list.Item{}
	delegate := NewMirrorDelegate(colors)
	modesList := list.New(modesItems, delegate, 0, 0)
	modesList.SetShowStatusBar(false)
	modesList.SetFilteringEnabled(false)
	modesList.SetShowHelp(false)
	modesList.SetShowTitle(false)

	return &MirrorList{
		L:        modesList,
		monitors: monitors,
		help:     NewCustomHelp(colors),
		colors:   colors,
	}
}

func (m *MirrorList) SetItems(monitor *MonitorSpec) tea.Cmd {
	logrus.Debugf("Setting the items: %v", monitor.AvailableModes)
	items := []list.Item{}
	items = append(items, MirrorItem{mirrorName: "none"})
	selectedItem := 0
	for _, other := range m.monitors {
		if other.Disabled {
			logrus.Debugf("Skipping a disabled monitor: %s", other.Name)
			continue
		}

		if monitor.Name != other.Name {
			items = append(items, MirrorItem{mirrorName: other.Name})
		}

		if other.Name == monitor.Mirror {
			selectedItem = len(items) - 1
		}
	}

	cmd := m.L.SetItems(items)
	m.L.Select(selectedItem)
	return cmd
}

func (m *MirrorList) ClearItems() tea.Cmd {
	cmd := m.L.SetItems([]list.Item{})
	m.L.ResetSelected()
	return cmd
}

func (m *MirrorList) Update(msg tea.Msg) tea.Cmd {
	// nolint:gocritic
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// nolint:gocritic
		switch msg.String() {
		case "esc":
			logrus.Debug("Close monitor mirror list")
			return CloseMonitorMirrorListCmd()
		}
	}
	var cmd tea.Cmd
	m.L, cmd = m.L.Update(msg)
	return cmd
}

func (m *MirrorList) View() string {
	sections := []string{}
	availHeight := m.L.Height()

	title := m.colors.TitleStyle().Margin(0, 0, 1, 0).Render("Select a monitor mirror")
	availHeight -= lipgloss.Height(title)
	sections = append(sections, title)

	logrus.Debugf("Items: %v", m.L.Items())

	help := m.help.ShortHelpView([]key.Binding{
		rootKeyMap.Up, rootKeyMap.Down, rootKeyMap.Enter, rootKeyMap.Back,
	})
	availHeight -= lipgloss.Height(help)

	m.L.SetHeight(availHeight)
	content := lipgloss.NewStyle().Height(availHeight).Render(m.L.View())
	sections = append(sections, content)

	sections = append(sections, help)

	return lipgloss.JoinVertical(lipgloss.Top, sections...)
}

func (m *MirrorList) SetHeight(height int) {
	m.L.SetHeight(height)
}
