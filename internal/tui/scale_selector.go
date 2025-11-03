package tui

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/sirupsen/logrus"
)

type scaleSelectorKeyMap struct {
	Up                  key.Binding
	Down                key.Binding
	Select              key.Binding
	Back                key.Binding
	EnableScaleSnapping key.Binding
	One                 key.Binding
	Two                 key.Binding
	Custom              key.Binding
}

func (s *scaleSelectorKeyMap) Help() []key.Binding {
	return []key.Binding{
		s.Up,
		s.Down,
		s.Select,
		s.Back,
		s.EnableScaleSnapping,
		s.One,
		s.Two,
		s.Custom,
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
	enableScaleSnapping  bool
	customInput          textinput.Model
	customInputMode      bool
}

func NewScaleSelector() *ScaleSelector {
	ti := textinput.New()
	ti.Placeholder = "Enter scale (e.g., 1.5)"
	ti.CharLimit = 12
	ti.Width = 30

	return &ScaleSelector{
		help:                help.New(),
		minScale:            0.1,
		maxScale:            10.0,
		step:                0.005,
		enableScaleSnapping: true,
		customInput:         ti,
		customInputMode:     false,
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
			EnableScaleSnapping: key.NewBinding(
				key.WithKeys("e"),
				key.WithHelp("e", "enable/disable scale snapping"),
			),
			One: key.NewBinding(
				key.WithKeys("1"),
				key.WithHelp("1", "1.00x"),
			),
			Two: key.NewBinding(
				key.WithKeys("2"),
				key.WithHelp("2", "2.00x"),
			),
			Custom: key.NewBinding(
				key.WithKeys("C"),
				key.WithHelp("C", "custom"),
			),
		},
	}
}

func (s *ScaleSelector) SetWidth(width int) {
	s.width = width
	s.customInput.Width = width - 2
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

	titleString := "Adjust scale"
	if s.enableScaleSnapping {
		titleString += " (snapping)"
	}

	title := TitleStyle.Width(s.width).Render(titleString)
	sections = append(sections, title)
	availableSpace -= lipgloss.Height(title)

	subtitle := SubtitleInfoStyle.Margin(0, 0, 1, 0).Width(s.width).Render(
		fmt.Sprintf("Tip: Hold for acceleration, single tick is %.8f", s.step))
	sections = append(sections, subtitle)
	availableSpace -= lipgloss.Height(subtitle)

	if s.customInputMode {
		inputView := s.customInput.View()
		sections = append(sections, inputView)
		availableSpace -= lipgloss.Height(inputView)

		inputHint := MutedItalicStyle.Render("Press Enter to apply, Esc to cancel")
		availableSpace -= lipgloss.Height(inputHint)

		spacer := lipgloss.NewStyle().Height(availableSpace).Render("")
		sections = append(sections, spacer)
		sections = append(sections, inputHint)

		return lipgloss.JoinVertical(lipgloss.Top, sections...)
	}

	dpi := s.calculateEffectiveDPI()
	scaleText := fmt.Sprintf("Scale: %.8f ", s.monitor.Scale)
	dpiText := MutedItalicStyle.Render(fmt.Sprintf("(%d DPI)", dpi))
	content := scaleText + dpiText
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
		if s.customInputMode {
			switch msg.Type {
			case tea.KeyEnter:
				value := s.customInput.Value()
				if scale, err := strconv.ParseFloat(value, 64); err == nil {
					s.currentScale = scale
					s.monitor.Scale = scale
					cmds = append(cmds, OperationStatusCmd(OperationNameTypeScale, nil))
					cmds = append(cmds, previewScaleMonitorCmd(*s.monitor.ID, s.currentScale))
					s.customInputMode = false
					s.customInput.SetValue("")
					s.customInput.Blur()
				} else {
					cmds = append(cmds, OperationStatusCmd(
						OperationNameTypeScale, errors.New("can't parse")))
				}
			case tea.KeyEsc:
				s.customInputMode = false
				s.customInput.SetValue("")
				s.customInput.Blur()
			default:
				var cmd tea.Cmd
				s.customInput, cmd = s.customInput.Update(msg)
				cmds = append(cmds, cmd)
			}
			return tea.Batch(cmds...)
		}

		// Normal mode key handling
		switch {
		case key.Matches(msg, s.keyMap.Up):
			err := s.adjustScale(s.step)
			cmds = append(cmds, OperationStatusCmd(OperationNameFindClosestScale, err))
			cmds = append(cmds, previewScaleMonitorCmd(*s.monitor.ID, s.currentScale))
		case key.Matches(msg, s.keyMap.Down):
			err := s.adjustScale(-s.step)
			cmds = append(cmds, OperationStatusCmd(OperationNameFindClosestScale, err))
			cmds = append(cmds, previewScaleMonitorCmd(*s.monitor.ID, s.currentScale))
		case key.Matches(msg, s.keyMap.Select), key.Matches(msg, s.keyMap.Back):
			logrus.Debug("Select in scale")
			cmds = append(cmds, scaleMonitorCmd(*s.monitor.ID, s.currentScale))
		case key.Matches(msg, s.keyMap.EnableScaleSnapping):
			s.enableScaleSnapping = !s.enableScaleSnapping
			if s.enableScaleSnapping {
				s.step = 0.005
			} else {
				s.step = 0.0001
			}
		case key.Matches(msg, s.keyMap.One):
			s.currentScale = 1.0
			cmds = append(cmds, previewScaleMonitorCmd(*s.monitor.ID, s.currentScale))
		case key.Matches(msg, s.keyMap.Two):
			s.currentScale = 2.0
			cmds = append(cmds, previewScaleMonitorCmd(*s.monitor.ID, s.currentScale))
		case key.Matches(msg, s.keyMap.Custom):
			// Enter custom input mode
			s.customInputMode = true
			s.customInput.SetValue(fmt.Sprintf("%.8f", s.currentScale))
			s.customInput.Focus()
			cmds = append(cmds, textinput.Blink)
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

func (s *ScaleSelector) calculateEffectiveDPI() int {
	if s.monitor == nil {
		return 0
	}

	// Calculate DPI based on standard 96 DPI reference
	return int(96 * s.monitor.Scale)
}

func (s *ScaleSelector) adjustScale(baseIncrement float64) error {
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

	logrus.Debugf("Scale set to %f", newScale)

	if !s.enableScaleSnapping {
		s.currentScale = newScale
		return nil
	}

	validScale, err := s.monitor.ClosestValidScale(newScale, increment > 0, increment < 0)
	if err != nil {
		s.currentScale = newScale
		return fmt.Errorf("cant find closest valid scale: %w", err)
	}

	s.currentScale = validScale
	return nil
}
