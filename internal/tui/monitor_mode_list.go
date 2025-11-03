package tui

import (
	"fmt"
	"io"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/sirupsen/logrus"
)

type ModeItem struct {
	mode string
}

type MonitorModeList struct {
	L        list.Model
	monitors []*MonitorSpec
	help     help.Model
}

func (m ModeItem) FilterValue() string {
	return m.mode
}

func (m ModeItem) View() string {
	var width, height int
	var refreshRate float64

	n, err := fmt.Sscanf(m.mode, "%dx%d@%fHz", &width, &height, &refreshRate)
	if err != nil || n != 3 {
		return m.mode
	}

	suffix := m.getResolutionSuffix(width, height)
	if suffix == "" {
		return m.mode
	}

	isExactStandard := IsStandardResolution(width, height) && IsWholeNumber(refreshRate) &&
		refreshRate >= minDesirableRefreshHz && IsFHDOrHigher(width, height)

	var styledSuffix string
	if isExactStandard {
		styledSuffix = StandardModeSuffixStyle.Render(" " + suffix)
	} else {
		styledSuffix = NonStandardModeSuffixStyle.Render(" " + suffix)
	}

	resolutionPart := fmt.Sprintf("%dx%d", width, height)
	refreshPart := fmt.Sprintf("%.0fHz", refreshRate)

	styledRefresh := RefreshRateBoldStyle.Render(" @ " + refreshPart)

	return resolutionPart + styledRefresh + styledSuffix
}

// getResolutionSuffix returns a styled suffix with aspect ratio and resolution class
func (m ModeItem) getResolutionSuffix(width, height int) string {
	aspectRatio := GetAspectRatio(width, height)
	resClass := GetResolutionClass(width, height)

	parts := []string{}
	if resClass != "" {
		parts = append(parts, resClass)
	}
	if aspectRatio != "" {
		parts = append(parts, aspectRatio)
	}

	if len(parts) == 0 {
		return ""
	}

	result := "["
	for i, part := range parts {
		if i > 0 {
			result += ", "
		}
		result += part
	}
	result += "]"

	return result
}

type ModeDelegate struct{}

func NewModeDelegate() ModeDelegate {
	return ModeDelegate{}
}

func (d ModeDelegate) Height() int {
	return 1
}

func (d ModeDelegate) Spacing() int {
	return 0
}

func (d ModeDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd {
	cmds := []tea.Cmd{}
	item, ok := m.SelectedItem().(ModeItem)
	if !ok {
		logrus.Warning("Monitor delegate called with an item that is not a ModeItem")
		return nil
	}
	// nolint:gocritic
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k", "down", "j":
			logrus.Debugf("Setting mode to: %s", item.mode)
			cmds = append(cmds, ChangeModePreviewCmd(item.mode))
		case "enter":
			logrus.Debugf("Setting final to: %s", item.mode)
			cmds = append(cmds, ChangeModeCmd(item.mode))
		}
	}
	return tea.Batch(cmds...)
}

func (d ModeDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	modeItem, ok := item.(ModeItem)
	if !ok {
		return
	}

	var style lipgloss.Style
	var prefix string
	switch {
	case index == m.Index():
		style = MonitorListSelected
		prefix = "► "
	default:
		style = MonitorListTitle
	}
	title := style.Render(prefix + modeItem.View())

	fmt.Fprintf(w, "%s", title)
}

func NewMonitorModeList(monitors []*MonitorSpec) *MonitorModeList {
	modesItems := []list.Item{}
	delegate := NewModeDelegate()
	modesList := list.New(modesItems, delegate, 0, 0)
	modesList.SetShowStatusBar(false)
	modesList.SetFilteringEnabled(false)
	modesList.SetShowHelp(false)
	modesList.SetShowTitle(false)

	return &MonitorModeList{
		L:        modesList,
		monitors: monitors,
		help:     help.New(),
	}
}

func (m *MonitorModeList) SetItems(monitor *MonitorSpec) tea.Cmd {
	logrus.Debugf("Setting the items: %v", monitor.AvailableModes)
	modesItems := []list.Item{}
	selectedMode := -1
	for i, mode := range monitor.AvailableModes {
		modesItems = append(modesItems, ModeItem{mode: mode})
		logrus.Debugf("Comparing modes: %s with %s", mode, monitor.ModeForComparison())
		if mode == monitor.ModeForComparison() {
			selectedMode = i
		}
	}

	cmd := m.L.SetItems(modesItems)
	m.L.Select(selectedMode)
	return cmd
}

func (m *MonitorModeList) ClearItems() tea.Cmd {
	cmd := m.L.SetItems([]list.Item{})
	m.L.ResetSelected()
	return cmd
}

func (m *MonitorModeList) Update(msg tea.Msg) tea.Cmd {
	// nolint:gocritic
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// nolint:gocritic
		switch msg.String() {
		case "esc":
			return CloseMonitorModeListCmd()
		}
	}
	var cmd tea.Cmd
	m.L, cmd = m.L.Update(msg)
	return cmd
}

func (m *MonitorModeList) View() string {
	sections := []string{}
	availHeight := m.L.Height()

	title := TitleStyle.Margin(0, 0, 1, 0).Render("Select monitor mode")
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

func (m *MonitorModeList) SetHeight(height int) {
	m.L.SetHeight(height)
}
