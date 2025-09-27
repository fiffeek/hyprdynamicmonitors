// Package tui provides a TUI implementation to interact with hyprdynamicmonitors (and hyprland) monitors config
package tui

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
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
	monitorsList list.Model
}

type keyMap struct {
	Tab   key.Binding
	Enter key.Binding
	Quit  key.Binding
}

var rootKeyMap = keyMap{
	Tab: key.NewBinding(
		key.WithKeys("tab"),
		key.WithHelp("tab", "switch view"),
	),
	Enter: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "select"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "ctrl+c"),
		key.WithHelp("q", "quit"),
	),
}

func NewModel(cfg *config.Config, monitors hypr.MonitorSpecs) Model {
	monitorItems := make([]list.Item, len(monitors))
	for i, monitor := range monitors {
		monitorItems[i] = MonitorItem{monitor: NewMonitorSpec(monitor)}
	}

	monitorsList := list.New(monitorItems, NewMonitorDelegate(), 0, 0)
	monitorsList.Title = "Connected Monitors"
	monitorsList.SetShowStatusBar(false)
	monitorsList.SetFilteringEnabled(false)

	model := Model{
		config:       cfg,
		keys:         rootKeyMap,
		rootState:    &RootState{},
		layout:       NewLayout(),
		monitorsList: monitorsList,
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
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.layout.SetHeight(msg.Height)
		m.layout.SetWidth(msg.Width)
		m.monitorsList.SetWidth(m.layout.LeftPanesWidth())
		m.monitorsList.SetHeight(m.layout.LeftMonitorsHeight())
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Quit):
			return m, tea.Quit
		case key.Matches(msg, m.keys.Tab):
		}
	}

	switch m.rootState.CurrentView {
	case MonitorsListView:
		monitorsList, cmd := m.monitorsList.Update(msg)
		m.monitorsList = monitorsList
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}
