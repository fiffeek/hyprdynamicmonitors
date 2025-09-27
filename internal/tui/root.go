// Package tui provides a TUI implementation to interact with hyprdynamicmonitors (and hyprland) monitors config
package tui

import (
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/fiffeek/hyprdynamicmonitors/internal/config"
	"github.com/fiffeek/hyprdynamicmonitors/internal/hypr"
	"github.com/sirupsen/logrus"
)

type Model struct {
	config    *config.Config
	keys      keyMap
	layout    *Layout
	rootState *RootState

	// components
	monitorsList        *MonitorList
	monitorsPreviewPane *MonitorsPreviewPane
	help                help.Model
	header              *Header
}

func NewModel(cfg *config.Config, hyprMonitors hypr.MonitorSpecs) Model {
	monitors := make([]*MonitorSpec, len(hyprMonitors))
	for i, monitor := range hyprMonitors {
		monitors[i] = NewMonitorSpec(monitor)
	}

	model := Model{
		config:              cfg,
		keys:                rootKeyMap,
		rootState:           &RootState{},
		layout:              NewLayout(),
		monitorsList:        NewMonitorList(monitors),
		monitorsPreviewPane: NewMonitorsPreviewPane(monitors),
		help:                help.New(),
		header:              NewHeader("HyprDynamicMonitors"),
	}

	return model
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) View() string {
	logrus.Debug("Rendering the root model")

	m.header.SetWidth(m.layout.visibleWidth)
	header := m.header.View()
	headerHeight := lipgloss.Height(header)

	globalHelp := HelpStyle.Width(m.layout.visibleWidth).Render(m.help.ShortHelpView(m.GlobalHelp()))
	globalHelpHeight := lipgloss.Height(globalHelp)

	m.layout.SetReservedTop(globalHelpHeight + headerHeight)

	m.monitorsList.SetHeight(m.layout.LeftMonitorsHeight())
	monitorView := InactiveStyle.Width(m.layout.LeftPanesWidth()).Height(m.layout.LeftMonitorsHeight()).Render(m.monitorsList.View())

	m.monitorsPreviewPane.SetHeight(m.layout.RightPreviewHeight())
	m.monitorsPreviewPane.SetWidth(m.layout.RightPanesWidth())
	previewPane := InactiveStyle.Width(m.layout.RightPanesWidth()).Height(m.layout.RightPreviewHeight()).Render(m.monitorsPreviewPane.View())

	if m.rootState.State == StateFullscreen {
		m.monitorsPreviewPane.SetHeight(m.layout.AvailableHeight())
		m.monitorsPreviewPane.SetWidth(m.layout.AvailableWidth())
		previewPane = InactiveStyle.Width(m.layout.AvailableWidth()).Height(m.layout.AvailableHeight()).Render(m.monitorsPreviewPane.View())
	}

	leftPanels := lipgloss.JoinVertical(
		lipgloss.Left,
		monitorView,
	)
	rightPanels := lipgloss.JoinVertical(
		lipgloss.Left,
		previewPane,
	)

	view := lipgloss.JoinHorizontal(
		lipgloss.Left,
		leftPanels,
		rightPanels,
	)

	if m.rootState.State == StateFullscreen {
		view = previewPane
	}

	screen := lipgloss.JoinVertical(lipgloss.Top, header, globalHelp, view)

	return lipgloss.Place(m.layout.visibleWidth, m.layout.visibleHeight, lipgloss.Center, lipgloss.Center, screen)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	logrus.Debugf("Received a message in root: %v", msg)
	var cmds []tea.Cmd
	stateChanged := false

	switch msg := msg.(type) {
	case MonitorSelected:
		logrus.Debug("Monitor selected event in root")
		m.rootState.ChangeState(StateEditingMonitor)
		stateChanged = true
	case MonitorUnselected:
		logrus.Debug("Monitor unselected event in root")
		m.rootState.ChangeState(StateNavigating)
		stateChanged = true
	case tea.WindowSizeMsg:
		m.layout.SetHeight(msg.Height)
		m.layout.SetWidth(msg.Width)
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Quit):
			return m, tea.Quit
		case key.Matches(msg, m.keys.Pan):
			logrus.Debug("Toggling pane mode")
			m.rootState.ToggleState(StatePanning)
			stateChanged = true
		case key.Matches(msg, m.keys.Fullscreen):
			logrus.Debug("Toggling fullscreen mode")
			m.rootState.ToggleState(StateFullscreen)
			stateChanged = true
		}
	}

	if stateChanged {
		cmds = append(cmds, func() tea.Msg {
			return StateChanged{
				state: m.rootState.State,
			}
		})
	}

	switch m.rootState.CurrentView {
	case MonitorsListView:
		cmd := m.monitorsList.Update(msg)
		cmds = append(cmds, cmd)
		cmd = m.monitorsPreviewPane.Update(msg)
		cmds = append(cmds, cmd)
	}

	cmd := m.header.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m *Model) GlobalHelp() []key.Binding {
	return []key.Binding{rootKeyMap.Pan, rootKeyMap.Fullscreen}
}
