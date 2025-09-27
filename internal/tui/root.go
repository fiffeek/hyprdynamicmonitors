// Package tui provides a TUI implementation to interact with hyprdynamicmonitors (and hyprland) monitors config
package tui

import (
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
	}

	return model
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) View() string {
	logrus.Debug("Rendering the root model")

	monitorView := InactiveStyle.Width(m.layout.LeftPanesWidth()).Render(m.monitorsList.View())
	previewPane := InactiveStyle.Width(m.layout.RightPanesWidth()).Render(m.monitorsPreviewPane.View())

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

	return lipgloss.Place(m.layout.visibleWidth, m.layout.visibleHeight, lipgloss.Center, lipgloss.Center, view)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	logrus.Debugf("Received a message in root: %v", msg)
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case MonitorSelected:
		logrus.Debug("Monitor selected event in root")
		m.rootState.ChangeState(StateEditingMonitor)
	case MonitorUnselected:
		logrus.Debug("Monitor unselected event in root")
		m.rootState.ChangeState(StateNavigating)
	case tea.WindowSizeMsg:
		m.layout.SetHeight(msg.Height)
		m.layout.SetWidth(msg.Width)
		cmd := m.monitorsList.Update(WindowResized{
			width:  m.layout.LeftPanesWidth(),
			height: m.layout.LeftMonitorsHeight(),
		})
		cmds = append(cmds, cmd)
		cmd = m.monitorsPreviewPane.Update(WindowResized{
			width:  m.layout.RightPanesWidth(),
			height: m.layout.RightPreviewHeight(),
		})
		cmds = append(cmds, cmd)
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Quit):
			return m, tea.Quit
		}
	}

	m.monitorsList.SetHijackArrows(m.rootState.State != StateNavigating)

	switch m.rootState.CurrentView {
	case MonitorsListView:
		cmd := m.monitorsList.Update(msg)
		cmds = append(cmds, cmd)
		cmd = m.monitorsPreviewPane.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}
