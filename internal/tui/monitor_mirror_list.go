package tui

import (
	"fmt"
	"io"

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
}

func (m MirrorItem) FilterValue() string {
	return m.mirrorName
}

func (m MirrorItem) View() string {
	return m.mirrorName
}

type MirrorDelegate struct{}

func NewMirrorDelegate() MirrorDelegate {
	return MirrorDelegate{}
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
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		// todo all keymaps here
		case "up", "k", "down", "j":
			logrus.Debugf("Setting mode to: %s", item.mirrorName)
			cmds = append(cmds, changeMirrorPreviewCmd(item.mirrorName))
		case "enter":
			logrus.Debugf("Setting final to: %s", item.mirrorName)
			cmds = append(cmds, changeMirrorCmd(item.mirrorName))
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
	switch {
	case index == m.Index():
		style = MonitorListSelected
	default:
		style = MonitorListTitle
	}
	title := style.Render(modeItem.View())
	content := title

	fmt.Fprintf(w, "%s", content)
}

func NewMirrorList(monitors []*MonitorSpec) *MirrorList {
	modesItems := []list.Item{}
	delegate := NewMirrorDelegate()
	modesList := list.New(modesItems, delegate, 0, 0)
	modesList.SetShowStatusBar(false)
	modesList.SetFilteringEnabled(false)
	modesList.SetShowHelp(true)
	modesList.SetShowTitle(false)

	return &MirrorList{
		L:        modesList,
		monitors: monitors,
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
	var cmd tea.Cmd
	m.L, cmd = m.L.Update(msg)
	return cmd
}

func (m *MirrorList) View() string {
	logrus.Debugf("Items: %v", m.L.Items())
	availHeight := m.L.Height()
	content := lipgloss.NewStyle().Height(availHeight).Render(m.L.View())
	return content
}

func (m *MirrorList) SetHeight(height int) {
	m.L.SetHeight(height)
}
