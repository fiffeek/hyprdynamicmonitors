// Package tui provides a TUI implementation to interact with hyprdynamicmonitors (and hyprland) monitors config
package tui

import (
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/fiffeek/hyprdynamicmonitors/internal/config"
	"github.com/fiffeek/hyprdynamicmonitors/internal/hypr"
	"github.com/fiffeek/hyprdynamicmonitors/internal/matchers"
	"github.com/fiffeek/hyprdynamicmonitors/internal/power"
	"github.com/fiffeek/hyprdynamicmonitors/internal/profilemaker"
	"github.com/sirupsen/logrus"
)

type Model struct {
	config    *config.Config
	keys      keyMap
	layout    *Layout
	rootState *RootState

	// universal components
	confirmationPrompt *ConfirmationPrompt

	// components for monitor view
	monitorsList        *MonitorList
	monitorsPreviewPane *MonitorsPreviewPane
	monitorModes        *MonitorModeList
	monitorMirrors      *MirrorList
	help                help.Model
	header              *Header
	hyprPreviewPane     *HyprPreviewPane
	scaleSelector       *ScaleSelector

	// components for config view
	hdm               *HDMConfigPane
	profileNamePicker *ProfileNamePicker
	hdmProfilePreview *HDMProfilePreview

	// stores
	monitorEditor *MonitorEditorStore

	// actions
	hyprApply    *HyprApply
	profileMaker *profilemaker.Service

	// for tests
	duration *time.Duration
	start    time.Time
}

func NewModel(cfg *config.Config, hyprMonitors hypr.MonitorSpecs,
	profileMaker *profilemaker.Service, version string, powerState power.PowerState, duration *time.Duration,
) Model {
	monitors := make([]*MonitorSpec, len(hyprMonitors))
	for i, monitor := range hyprMonitors {
		monitors[i] = NewMonitorSpec(monitor)
	}

	state := NewState(monitors, cfg)
	matcher := matchers.NewMatcher()

	model := Model{
		config:              cfg,
		keys:                rootKeyMap,
		rootState:           state,
		layout:              NewLayout(),
		monitorsList:        NewMonitorList(monitors),
		monitorsPreviewPane: NewMonitorsPreviewPane(monitors),
		help:                help.New(),
		header:              NewHeader("HyprDynamicMonitors", state.viewModes, version),
		hyprPreviewPane:     NewHyprPreviewPane(monitors),
		monitorEditor:       NewMonitorEditor(monitors),
		monitorModes:        NewMonitorModeList(monitors),
		monitorMirrors:      NewMirrorList(monitors),
		hyprApply:           NewHyprApply(profileMaker),
		hdm:                 NewHDMConfigPane(cfg, matcher, monitors, powerState),
		profileMaker:        profileMaker,
		profileNamePicker:   NewProfileNamePicker(),
		confirmationPrompt:  nil,
		scaleSelector:       NewScaleSelector(),
		hdmProfilePreview:   NewHDMProfilePreview(cfg, matcher, monitors, powerState),
		start:               time.Now(),
		duration:            duration,
	}

	return model
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) timeLeft() time.Duration {
	// infinite time effectively when no duration is passed
	if m.duration == nil {
		return time.Minute
	}
	return *m.duration - time.Since(m.start)
}

func (m Model) View() string {
	logrus.Debug("Rendering the root model")

	if m.rootState.State.ShowConfirmationPrompt {
		m.confirmationPrompt.SetWidth(m.layout.PromptWidth())
		m.confirmationPrompt.SetHeight(m.layout.PromptHeight())
		prompt := ActiveStyle.Width(m.layout.PromptWidth()).Height(
			m.layout.PromptHeight()).Render(m.confirmationPrompt.View())

		return lipgloss.Place(m.layout.AvailableWidth(), m.layout.AvailableHeight(),
			lipgloss.Center, lipgloss.Center, prompt)
	}

	logrus.Debugf("Visible height: %d", m.layout.visibleHeight)

	m.header.SetWidth(m.layout.visibleWidth)
	header := m.header.View()
	headerHeight := lipgloss.Height(header)

	globalHelp := HelpStyle.Margin(0, 0, 1, 0).Width(m.layout.visibleWidth).Render(m.help.ShortHelpView(m.GlobalHelp()))
	globalHelpHeight := lipgloss.Height(globalHelp)

	m.layout.SetReservedTop(globalHelpHeight + headerHeight + 2)
	logrus.Debugf("Available height: %d", m.layout.AvailableHeight())

	left := m.leftPanels()
	right := m.rightPanels()

	leftPanels := lipgloss.JoinVertical(
		lipgloss.Left,
		left...,
	)
	rightPanels := lipgloss.JoinVertical(
		lipgloss.Left,
		right...,
	)

	view := lipgloss.JoinHorizontal(
		lipgloss.Left,
		leftPanels,
		rightPanels,
	)

	if m.rootState.State.Fullscreen {
		m.monitorsPreviewPane.SetHeight(m.layout.AvailableHeight() + 2)
		m.monitorsPreviewPane.SetWidth(m.layout.AvailableWidth())
		previewPane := ActiveStyle.Width(m.layout.AvailableWidth()).Height(
			m.layout.AvailableHeight() + 2).Render(m.monitorsPreviewPane.View())
		view = previewPane
	}

	return lipgloss.JoinVertical(lipgloss.Top, header, view, globalHelp)
}

func (m Model) rightPanels() []string {
	rightSections := []string{}

	if m.rootState.CurrentView() == ProfileView {
		m.hdmProfilePreview.SetHeight(m.layout.RightPreviewHeight())
		m.hdmProfilePreview.SetWidth(m.layout.RightPanesWidth())
		profileCfg := InactiveStyle.Height(m.layout.AvailableHeight() + 2).Width(
			m.layout.RightPanesWidth()).Render(m.hdmProfilePreview.View())
		rightSections = append(rightSections, profileCfg)
	}

	if m.rootState.CurrentView() == MonitorsListView {
		previewStyle := InactiveStyle
		if m.rootState.State.Panning {
			previewStyle = ActiveStyle
		}
		m.monitorsPreviewPane.SetHeight(m.layout.RightPreviewHeight())
		m.monitorsPreviewPane.SetWidth(m.layout.RightPanesWidth())
		previewPane := previewStyle.Width(m.layout.RightPanesWidth()).Height(
			m.layout.RightPreviewHeight()).Render(m.monitorsPreviewPane.View())

		if m.rootState.State.Fullscreen {
			m.monitorsPreviewPane.SetHeight(m.layout.AvailableHeight() + 2)
			m.monitorsPreviewPane.SetWidth(m.layout.AvailableWidth())
			previewPane = previewStyle.Width(m.layout.AvailableWidth()).Height(
				m.layout.AvailableHeight() + 2).Render(m.monitorsPreviewPane.View())
		}

		rightSections = append(rightSections, previewPane)
		if !m.rootState.State.Fullscreen {
			hyprPreview := InactiveStyle.Width(m.layout.RightPanesWidth()).Height(
				m.layout.RightHyprHeight()).Render(m.hyprPreviewPane.View())
			rightSections = append(rightSections, hyprPreview)
		}
	}

	return rightSections
}

func (m Model) leftPanels() []string {
	left := []string{}
	leftMainPanelSize := m.layout.AvailableHeight() + 2

	// config view, different panels
	if m.rootState.CurrentView() == ProfileView {
		if m.rootState.State.ProfileNameRequested {
			leftMainPanelSize = m.layout.LeftMonitorsHeight()
		}
		m.hdm.SetHeight(leftMainPanelSize)
		m.hdm.SetWidth(m.layout.LeftPanesWidth())
		mainPanelStyle := ActiveStyle
		if m.rootState.State.ProfileNameRequested {
			mainPanelStyle = InactiveStyle
		}
		view := mainPanelStyle.Width(m.layout.LeftPanesWidth()).Height(
			leftMainPanelSize).Render(m.hdm.View())
		left = append(left, view)

		if m.rootState.State.ProfileNameRequested {
			m.profileNamePicker.SetHeight(m.layout.LeftSubpaneHeight())
			m.profileNamePicker.SetWidth(m.layout.LeftPanesWidth())
			namePicker := ActiveStyle.Width(m.layout.LeftPanesWidth()).Height(
				m.layout.LeftSubpaneHeight()).Render(m.profileNamePicker.View())
			left = append(left, namePicker)
		}
		return left
	}

	// monitor view, different panels
	if m.rootState.CurrentView() == MonitorsListView {
		if m.rootState.State.ModeSelection || m.rootState.State.MirrorSelection || m.rootState.State.Scaling {
			leftMainPanelSize = m.layout.LeftMonitorsHeight()
		}
		m.monitorsList.SetHeight(leftMainPanelSize)
		m.monitorsList.SetWidth(m.layout.LeftPanesWidth())
		logrus.Debugf("Monitors list height: %d", leftMainPanelSize)
		monitorViewStyle := ActiveStyle
		if m.rootState.State.ModeSelection || m.rootState.State.MirrorSelection ||
			m.rootState.State.Scaling || m.rootState.State.Panning {
			monitorViewStyle = InactiveStyle
		}
		monitorView := monitorViewStyle.Width(m.layout.LeftPanesWidth()).Height(
			leftMainPanelSize).Render(m.monitorsList.View())
		left = append(left, monitorView)

		subpaneStyle := InactiveStyle
		if !m.rootState.State.Panning {
			subpaneStyle = ActiveStyle
		}

		if m.rootState.State.ModeSelection {
			m.monitorModes.SetHeight(m.layout.LeftSubpaneHeight())
			modeSelectionPane := subpaneStyle.Width(m.layout.LeftPanesWidth()).Height(
				m.layout.LeftSubpaneHeight()).Render(m.monitorModes.View())
			logrus.Debugf("Mode selection pane height: %d", m.layout.LeftSubpaneHeight())
			left = append(left, modeSelectionPane)
		}

		if m.rootState.State.MirrorSelection {
			m.monitorMirrors.SetHeight(m.layout.LeftSubpaneHeight())
			pane := subpaneStyle.Width(m.layout.LeftPanesWidth()).Height(
				m.layout.LeftSubpaneHeight()).Render(m.monitorMirrors.View())
			logrus.Debugf("Mirrors pane height: %d", m.layout.LeftSubpaneHeight())
			left = append(left, pane)
		}

		if m.rootState.State.Scaling {
			m.scaleSelector.SetHeight(m.layout.LeftSubpaneHeight())
			m.scaleSelector.SetWidth(m.layout.LeftPanesWidth())
			scalingSelector := subpaneStyle.Width(m.layout.LeftPanesWidth()).Height(
				m.layout.LeftSubpaneHeight()).Render(m.scaleSelector.View())
			logrus.Debugf("Scaling selection pane height: %d", m.layout.LeftSubpaneHeight())
			left = append(left, scalingSelector)
		}
	}

	return left
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.timeLeft() <= 0 {
		return m, tea.Quit
	}

	logrus.Debugf("Received a message in root: %v", msg)
	var cmds []tea.Cmd
	stateChanged := false

	switch msg := msg.(type) {
	case ConfigReloaded:
		logrus.Debug("Received config reloaded event in root")
		cmds = append(cmds, OperationStatusCmd(OperationNameHDMConfigReloadRequested, nil))
	case MonitorBeingEdited:
		logrus.Debug("Monitor selected event in root")
		m.rootState.SetMonitorEditState(msg)
		stateChanged = true
		// todo move this to the components, let them rely on the index, easier testing ?
		cmds = append(cmds, m.monitorModes.SetItems(m.rootState.monitors[msg.ListIndex]))
		cmds = append(cmds, m.monitorMirrors.SetItems(m.rootState.monitors[msg.ListIndex]))
		cmds = append(cmds, m.scaleSelector.Set(m.rootState.monitors[msg.ListIndex]))
	case MonitorUnselected:
		logrus.Debug("Monitor unselected event in root")
		m.rootState.ClearMonitorEditState()
		cmds = append(cmds, m.monitorModes.ClearItems())
		cmds = append(cmds, m.monitorMirrors.ClearItems())
		cmds = append(cmds, m.scaleSelector.Unset())
		stateChanged = true
	case MoveMonitorCommand:
		logrus.Debug("Received a monitor move command")
		cmds = append(cmds, m.monitorEditor.MoveMonitor(msg.MonitorID, msg.StepX, msg.StepY))
	case ToggleMonitorCommand:
		logrus.Debug("Received a monitor toggle command")
		cmds = append(cmds, m.monitorEditor.ToggleDisable(msg.MonitorID))
	case ToggleMonitorVRRCommand:
		logrus.Debug("Received a monitor vrr command")
		cmds = append(cmds, m.monitorEditor.ToggleVRR(msg.MonitorID))
	case RotateMonitorCommand:
		logrus.Debug("Received a monitor rotate command")
		cmds = append(cmds, m.monitorEditor.RotateMonitor(msg.MonitorID))
	case PreviewScaleMonitorCommand:
		logrus.Debug("Received a monitor scale command")
		cmds = append(cmds, m.monitorEditor.ScaleMonitor(msg.monitorID, msg.scale))
	case ScaleMonitorCommand:
		logrus.Debug("Received a monitor scale command")
		cmds = append(cmds, m.monitorEditor.ScaleMonitor(msg.monitorID, msg.scale))
		cmds = append(cmds, m.monitorsList.Update(msg))
	case ChangeMirrorPreviewCommand:
		logrus.Debug("Received preview change for monitor mirror")
		cmds = append(cmds, m.monitorEditor.SetMirror(
			m.rootState.State.MonitorEditedListIndex, msg.MirrorOf))
	case ChangeMirrorCommand:
		logrus.Debug("Received change for monitor mirror")
		cmds = append(cmds, m.monitorEditor.SetMirror(
			m.rootState.State.MonitorEditedListIndex, msg.MirrorOf))
		cmds = append(cmds, m.monitorsList.Update(msg))
	case ChangeModePreviewCommand:
		logrus.Debug("Received preview change for monitor mode")
		cmds = append(cmds, m.monitorEditor.SetMode(
			m.rootState.State.MonitorEditedListIndex, msg.Mode))
	case CloseMonitorModeListCommand, CloseMonitorMirrorListCommand:
		cmds = append(cmds, m.monitorsList.Update(msg))
	case ChangeModeCommand:
		logrus.Debug("Received change for monitor mode")
		cmds = append(cmds, m.monitorEditor.SetMode(
			m.rootState.State.MonitorEditedListIndex, msg.Mode))
		cmds = append(cmds, m.monitorsList.Update(msg))
	case CreateNewProfileCommand:
		logrus.Debug("Received create new profile")
		cmds = append(cmds, m.hyprApply.CreateProfile(m.monitorEditor.GetMonitors(), msg.name, msg.file))
		cmds = append(cmds, m.hdm.Update(msg))
	case EditProfileConfirmationCommand:
		logrus.Debug("Received edit existing profile confirm")
		m.confirmationPrompt = NewConfirmationPrompt(
			fmt.Sprintf("Apply edited settings to %s profile?", msg.name),
			tea.Batch(toggleConfirmationPromptCmd(), editProfileCmd(msg.name)),
			toggleConfirmationPromptCmd())
		m.rootState.ToggleConfirmationPrompt()
		stateChanged = true
	case EditProfileCommand:
		logrus.Debug("Received edit existing profile")
		cmds = append(cmds, m.hyprApply.EditProfile(m.monitorEditor.GetMonitors(), msg.name))
	case ProfileNameToggled:
		logrus.Debug("Profile name requested")
		m.rootState.ToggleProfileNameRequested()
		stateChanged = true
	case ApplyEphemeralCommand:
		logrus.Debug("Applying hypr settings")
		cmds = append(cmds, m.hyprApply.ApplyCurrent(m.monitorEditor.GetMonitors()))
	case ToggleConfirmationPromptCommand:
		logrus.Debug("Toggling confirmation prompt")
		m.rootState.ToggleConfirmationPrompt()
		stateChanged = true
	case tea.WindowSizeMsg:
		m.layout.SetHeight(msg.Height)
		m.layout.SetWidth(msg.Width)
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Quit):
			return m, tea.Quit
		case key.Matches(msg, m.keys.Tab):
			if !m.rootState.State.ShowConfirmationPrompt {
				m.rootState.NextView()
				cmds = append(cmds, ViewChangedCmd(m.rootState.CurrentView()))
			}
		}
	}

	if m.rootState.CurrentView() == ProfileView {
		// nolint:gocritic
		switch msg := msg.(type) {
		case tea.KeyMsg:
			// nolint:gocritic
			switch {
			case key.Matches(msg, m.keys.EditHDMConfig):
				cmds = append(cmds, openEditor(m.config.Get().ConfigPath))
			}
		}
	}

	if m.rootState.CurrentView() == MonitorsListView {
		// nolint:gocritic
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch {
			case key.Matches(msg, m.keys.FollowMonitor):
				m.rootState.ToggleFollowMonitorMode()
				stateChanged = true
			case key.Matches(msg, m.keys.Pan):
				logrus.Debug("Toggling pane mode")
				m.rootState.TogglePanning()
				stateChanged = true
			case key.Matches(msg, m.keys.Fullscreen):
				logrus.Debug("Toggling fullscreen mode")
				m.rootState.ToggleFullscreen()
				stateChanged = true
			case key.Matches(msg, m.keys.ToggleSnapping):
				logrus.Debug("Toggling snapping")
				m.rootState.ToggleSnapping()
				m.monitorEditor.SetSnapping(m.rootState.State.Snapping)
				stateChanged = true
			case key.Matches(msg, m.keys.ApplyHypr):
				logrus.Debug("Toggling confirmationPrompt for hypr apply")
				m.confirmationPrompt = NewConfirmationPrompt(
					"Apply hypr settings now?",
					tea.Batch(toggleConfirmationPromptCmd(), applyEphemeralCmd()),
					toggleConfirmationPromptCmd())
				m.rootState.ToggleConfirmationPrompt()
				stateChanged = true
			}
		}
	}
	if stateChanged {
		cmds = append(cmds, func() tea.Msg {
			return StateChanged{
				State: m.rootState.State,
			}
		})
	}

	if !m.rootState.State.ShowConfirmationPrompt {
		switch m.rootState.CurrentView() {
		case MonitorsListView:
			if !m.rootState.State.Panning {
				switch {
				case m.rootState.State.ModeSelection:
					cmd := m.monitorModes.Update(msg)
					cmds = append(cmds, cmd)
				case m.rootState.State.MirrorSelection:
					cmd := m.monitorMirrors.Update(msg)
					cmds = append(cmds, cmd)
				case m.rootState.State.Scaling:
					cmd := m.scaleSelector.Update(msg)
					cmds = append(cmds, cmd)
				default:
					cmd := m.monitorsList.Update(msg)
					cmds = append(cmds, cmd)
				}
			}

			cmd := m.monitorsPreviewPane.Update(msg)
			cmds = append(cmds, cmd)
		case ProfileView:
			logrus.Debug("Update for config view")
			if m.rootState.State.ProfileNameRequested {
				cmds = append(cmds, m.profileNamePicker.Update(msg))
			} else {
				cmds = append(cmds, m.hdm.Update(msg))
				cmds = append(cmds, m.hdmProfilePreview.Update(msg))
			}
		}
	} else {
		cmds = append(cmds, m.confirmationPrompt.Update(msg))
	}

	cmd := m.header.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m *Model) GlobalHelp() []key.Binding {
	bindings := []key.Binding{}
	if m.rootState.HasMoreThanOneView() {
		bindings = append(bindings, rootKeyMap.Tab)
	}
	if m.rootState.CurrentView() == MonitorsListView {
		monitors := []key.Binding{
			rootKeyMap.Pan,
			rootKeyMap.Fullscreen, rootKeyMap.FollowMonitor, rootKeyMap.Center, rootKeyMap.ZoomIn, rootKeyMap.ZoomOut,
			rootKeyMap.ToggleSnapping, rootKeyMap.ApplyHypr,
		}
		bindings = append(bindings, monitors...)

	}
	if m.rootState.CurrentView() == ProfileView {
		profile := []key.Binding{
			rootKeyMap.EditHDMConfig,
		}
		bindings = append(bindings, profile...)
	}
	return bindings
}
