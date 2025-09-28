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
	monitorModes        *MonitorModeList
	help                help.Model
	header              *Header
	hyprPreviewPane     *HyprPreviewPane

	// stores
	monitorEditor *MonitorEditorStore
}

func NewModel(cfg *config.Config, hyprMonitors hypr.MonitorSpecs) Model {
	monitors := make([]*MonitorSpec, len(hyprMonitors))
	for i, monitor := range hyprMonitors {
		monitors[i] = NewMonitorSpec(monitor)
	}

	model := Model{
		config:              cfg,
		keys:                rootKeyMap,
		rootState:           NewState(monitors),
		layout:              NewLayout(),
		monitorsList:        NewMonitorList(monitors),
		monitorsPreviewPane: NewMonitorsPreviewPane(monitors),
		help:                help.New(),
		header:              NewHeader("HyprDynamicMonitors"),
		hyprPreviewPane:     NewHyprPreviewPane(monitors),
		monitorEditor:       NewMonitorEditor(monitors),
		monitorModes:        NewMonitorModeList(monitors),
	}

	return model
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) View() string {
	logrus.Debug("Rendering the root model")
	logrus.Debugf("Visible height: %d", m.layout.visibleHeight)
	rightSections := []string{}

	m.header.SetWidth(m.layout.visibleWidth)
	header := m.header.View()
	headerHeight := lipgloss.Height(header)

	globalHelp := HelpStyle.Width(m.layout.visibleWidth).Render(m.help.ShortHelpView(m.GlobalHelp()))
	globalHelpHeight := lipgloss.Height(globalHelp)

	m.layout.SetReservedTop(globalHelpHeight + headerHeight + 2)
	logrus.Debugf("Available height: %d", m.layout.AvailableHeight())

	leftMonitorsHeight := m.layout.AvailableHeight() + 2
	if m.rootState.State.ModeSelection {
		leftMonitorsHeight = m.layout.LeftMonitorsHeight()
	}
	m.monitorsList.SetHeight(leftMonitorsHeight)
	m.monitorsList.SetWidth(m.layout.LeftPanesWidth())
	logrus.Debugf("Monitors list height: %d", leftMonitorsHeight)
	monitorViewStyle := ActiveStyle
	if m.rootState.State.ModeSelection {
		monitorViewStyle = InactiveStyle
	}
	monitorView := monitorViewStyle.Width(m.layout.LeftPanesWidth()).Height(
		leftMonitorsHeight).Render(m.monitorsList.View())

	m.monitorsPreviewPane.SetHeight(m.layout.RightPreviewHeight())
	m.monitorsPreviewPane.SetWidth(m.layout.RightPanesWidth())
	previewPane := InactiveStyle.Width(m.layout.RightPanesWidth()).Height(
		m.layout.RightPreviewHeight()).Render(m.monitorsPreviewPane.View())

	if m.rootState.State.Fullscreen {
		m.monitorsPreviewPane.SetHeight(m.layout.AvailableHeight())
		m.monitorsPreviewPane.SetWidth(m.layout.AvailableWidth())
		previewPane = InactiveStyle.Width(m.layout.AvailableWidth()).Height(
			m.layout.AvailableHeight()).Render(m.monitorsPreviewPane.View())
	}

	rightSections = append(rightSections, previewPane)
	if !m.rootState.State.Fullscreen {
		hyprPreview := InactiveStyle.Width(m.layout.RightPanesWidth()).Height(
			m.layout.RightHyprHeight()).Render(m.hyprPreviewPane.View())
		rightSections = append(rightSections, hyprPreview)
	}

	left := []string{monitorView}

	if m.rootState.State.ModeSelection {
		m.monitorModes.SetHeight(m.layout.LeftSubpaneHeight())
		modeSelectionPane := ActiveStyle.Width(m.layout.LeftPanesWidth()).Height(m.layout.LeftSubpaneHeight()).Render(m.monitorModes.View())
		logrus.Debugf("Mode selection pane height: %d", m.layout.LeftSubpaneHeight())
		left = append(left, modeSelectionPane)
	}

	leftPanels := lipgloss.JoinVertical(
		lipgloss.Left,
		left...,
	)
	rightPanels := lipgloss.JoinVertical(
		lipgloss.Left,
		rightSections...,
	)

	view := lipgloss.JoinHorizontal(
		lipgloss.Left,
		leftPanels,
		rightPanels,
	)

	if m.rootState.State.Fullscreen {
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
	case MonitorBeingEdited:
		logrus.Debug("Monitor selected event in root")
		m.rootState.SetMonitorEditState(msg)
		stateChanged = true
		cmds = append(cmds, m.monitorModes.SetItems(m.rootState.monitors[msg.ListIndex]))
	case MonitorUnselected:
		logrus.Debug("Monitor unselected event in root")
		m.rootState.ClearMonitorEditState()
		cmds = append(cmds, m.monitorModes.ClearItems())
		stateChanged = true
	case MoveMonitorCommand:
		logrus.Debug("Received a monitor move command")
		cmds = append(cmds, m.monitorEditor.MoveMonitor(msg.monitorID, msg.stepX, msg.stepY))
	case ToggleMonitorCommand:
		logrus.Debug("Received a monitor toggle command")
		cmds = append(cmds, m.monitorEditor.ToggleDisable(msg.monitorID))
	case ToggleMonitorVRRCommand:
		logrus.Debug("Received a monitor vrr command")
		cmds = append(cmds, m.monitorEditor.ToggleVRR(msg.monitorID))
	case RotateMonitorCommand:
		logrus.Debug("Received a monitor rotate command")
		cmds = append(cmds, m.monitorEditor.RotateMonitor(msg.monitorID))
	case ScaleMonitorCommand:
		logrus.Debug("Received a monitor scale command")
		cmds = append(cmds, m.monitorEditor.ScaleMonitor(msg.monitorID, msg.delta))
	case ChangeModePreviewCommand:
		logrus.Debug("Received preview change for monitor mode")
		cmds = append(cmds, m.monitorEditor.SetMode(m.rootState.State.MonitorEditedListIndex, msg.mode))
	case ChangeModeCommand:
		logrus.Debug("Received preview change for monitor mode")
		cmds = append(cmds, m.monitorEditor.SetMode(m.rootState.State.MonitorEditedListIndex, msg.mode))
		cmds = append(cmds, m.monitorsList.Update(msg))
	case tea.WindowSizeMsg:
		m.layout.SetHeight(msg.Height)
		m.layout.SetWidth(msg.Width)
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Quit):
			return m, tea.Quit
		case key.Matches(msg, m.keys.Pan):
			logrus.Debug("Toggling pane mode")
			m.rootState.TogglePanning()
			stateChanged = true
		case key.Matches(msg, m.keys.Fullscreen):
			logrus.Debug("Toggling fullscreen mode")
			m.rootState.ToggleFullscreen()
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
		if !m.rootState.State.Panning {
			if m.rootState.State.ModeSelection {
				cmd := m.monitorModes.Update(msg)
				cmds = append(cmds, cmd)
			} else {
				cmd := m.monitorsList.Update(msg)
				cmds = append(cmds, cmd)
			}
		}

		cmd := m.monitorsPreviewPane.Update(msg)
		cmds = append(cmds, cmd)
	}

	cmd := m.header.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m *Model) GlobalHelp() []key.Binding {
	return []key.Binding{rootKeyMap.Pan, rootKeyMap.Fullscreen, rootKeyMap.Center, rootKeyMap.ZoomIn, rootKeyMap.ZoomOut}
}
