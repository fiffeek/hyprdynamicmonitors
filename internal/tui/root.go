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
	monitorsList *MonitorList
}

func NewModel(cfg *config.Config, monitors hypr.MonitorSpecs) Model {
	monitorList := NewMonitorList(monitors)

	model := Model{
		config:       cfg,
		keys:         rootKeyMap,
		rootState:    &RootState{},
		layout:       NewLayout(),
		monitorsList: monitorList,
	}

	return model
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) View() string {
	logrus.Debug("Rendering the root model")

	monitorView := InactiveStyle.Width(m.layout.LeftPanesWidth()).Render(m.monitorsList.View())

	return lipgloss.JoinVertical(
		lipgloss.Left,
		monitorView,
	)
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
		m.monitorsList.Update(WindowResized{
			width:  m.layout.LeftPanesWidth(),
			height: m.layout.LeftMonitorsHeight(),
		})
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
	}

	return m, tea.Batch(cmds...)
}
