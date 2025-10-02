package tui

import (
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/sirupsen/logrus"
)

type scaleSelectorKeyMap struct {
	Up     key.Binding
	Down   key.Binding
	Select key.Binding
	Back   key.Binding
}

func (s *scaleSelectorKeyMap) Help() []key.Binding {
	return []key.Binding{
		s.Up,
		s.Down,
		s.Select,
		s.Back,
	}
}

type ScaleSelector struct {
	help                 help.Model
	minScale             float64
	maxScale             float64
	keyMap               *scaleSelectorKeyMap
	width                int
	height               int
	selectedMonitorIndex int
	monitor              *MonitorSpec
	currentScale         float64
	lastKeyTime          time.Time
	repeatCount          int
	step                 float64
}

func NewScaleSelector() *ScaleSelector {
	return &ScaleSelector{
		help:     help.New(),
		minScale: 0.1,
		maxScale: 10.0,
		step:     0.01,
		keyMap: &scaleSelectorKeyMap{
			Up: key.NewBinding(
				key.WithKeys("up", "k"),
				key.WithHelp("up/k", "increase"),
			),
			Down: key.NewBinding(
				key.WithKeys("down", "j"),
				key.WithHelp("down/j", "decrease"),
			),
			Select: key.NewBinding(
				key.WithKeys("enter"),
				key.WithHelp("enter", "select"),
			),
			Back: key.NewBinding(
				key.WithKeys("esc"),
				key.WithHelp("esc", "back"),
			),
		},
	}
}

func (s *ScaleSelector) SetWidth(width int) {
	s.width = width
}

func (s *ScaleSelector) SetHeight(height int) {
	s.height = height
}

func (s *ScaleSelector) Unset() tea.Cmd {
	s.currentScale = 0
	return nil
}

func (s *ScaleSelector) Set(monitor *MonitorSpec) tea.Cmd {
	s.monitor = monitor
	s.currentScale = monitor.Scale
	return nil
}

func (s *ScaleSelector) View() string {
	if s.monitor == nil {
		return ""
	}

	sections := []string{}
	availableSpace := s.height
	logrus.Debugf("availableSpace for ScaleSelector: %d", availableSpace)

	title := TitleStyle.Width(s.width).Render("Adjust scale")
	sections = append(sections, title)
	availableSpace -= lipgloss.Height(title)

	subtitle := SubtitleInfoStyle.Margin(0, 0, 1, 0).Width(s.width).Render(
		fmt.Sprintf("Tip: Hold for acceleration, single tick is %.2f", s.step))
	sections = append(sections, subtitle)
	availableSpace -= lipgloss.Height(subtitle)

	content := fmt.Sprintf("Scale: %.2f", s.monitor.Scale)
	sections = append(sections, content)
	availableSpace -= lipgloss.Height(content)

	help := lipgloss.NewStyle().Width(s.width).Render(s.help.ShortHelpView(s.keyMap.Help()))
	logrus.Debug("HELP", lipgloss.Height(help), lipgloss.Width(help), s.width)
	availableSpace -= lipgloss.Height(help)

	logrus.Debugf("spacer height: %d", availableSpace)

	spacer := lipgloss.NewStyle().Height(availableSpace).Render("")
	sections = append(sections, spacer)
	sections = append(sections, help)

	return lipgloss.JoinVertical(lipgloss.Top, sections...)
}

func (s *ScaleSelector) Update(msg tea.Msg) tea.Cmd {
	cmds := []tea.Cmd{}
	switch msg := msg.(type) {
	case MonitorBeingEdited:
		s.selectedMonitorIndex = msg.MonitorID
	case MonitorUnselected:
		s.selectedMonitorIndex = -1
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, s.keyMap.Up):
			s.adjustScale(s.step)
			cmds = append(cmds, previewScaleMonitorCmd(*s.monitor.ID, s.currentScale))
		case key.Matches(msg, s.keyMap.Down):
			s.adjustScale(-s.step)
			cmds = append(cmds, previewScaleMonitorCmd(*s.monitor.ID, s.currentScale))
		case key.Matches(msg, s.keyMap.Select), key.Matches(msg, s.keyMap.Back):
			logrus.Debug("Select in scale")
			cmds = append(cmds, scaleMonitorCmd(*s.monitor.ID, s.currentScale))
		}
	}
	return tea.Batch(cmds...)
}

func (s *ScaleSelector) GetCurrentScale() float64 {
	return s.currentScale
}

func (s *ScaleSelector) GetSelectedMonitorIndex() int {
	return s.selectedMonitorIndex
}

func (s *ScaleSelector) adjustScale(baseIncrement float64) {
	now := time.Now()

	if now.Sub(s.lastKeyTime) < 200*time.Millisecond {
		s.repeatCount++
	} else {
		s.repeatCount = 0
	}

	s.lastKeyTime = now

	multiplier := 1.0
	switch {
	case s.repeatCount > 5:
		multiplier = 10.0
	case s.repeatCount > 3:
		multiplier = 5.0
	case s.repeatCount > 1:
		multiplier = 2.0
	}

	increment := baseIncrement * multiplier
	newScale := s.currentScale + increment

	if newScale < s.minScale {
		newScale = s.minScale
	}
	if newScale > s.maxScale {
		newScale = s.maxScale
	}

	s.currentScale = newScale
}
