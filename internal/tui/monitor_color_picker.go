package tui

import (
	"fmt"
	"io"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/sirupsen/logrus"
)

type colorPresetItem struct {
	preset ColorPreset
}

func (m colorPresetItem) FilterValue() string {
	return m.preset.Value()
}

func (m colorPresetItem) View() string {
	return m.preset.Value()
}

type colorPresetDelegate struct {
	colors *ColorsManager
}

func NewColorPresetDelegate(colors *ColorsManager) colorPresetDelegate {
	return colorPresetDelegate{
		colors: colors,
	}
}

func (d colorPresetDelegate) Height() int {
	return 1
}

func (d colorPresetDelegate) Spacing() int {
	return 0
}

func (d colorPresetDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd {
	cmds := []tea.Cmd{}
	item, ok := m.SelectedItem().(colorPresetItem)
	if !ok {
		logrus.Warning("color preset delegate called with an item that is not a colorPresetItem")
		return nil
	}
	// nolint:gocritic
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k", "down", "j":
			logrus.Debugf("Setting color preset to: %s", item.preset.Value())
			cmds = append(cmds, ChangeColorPresetCmd(item.preset))
		case "enter":
			logrus.Debugf("Setting final preset to: %s", item.preset.Value())
			cmds = append(cmds, ChangeColorPresetFinalCmd(item.preset))
		}
	}
	return tea.Batch(cmds...)
}

func (d colorPresetDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	modeItem, ok := item.(colorPresetItem)
	if !ok {
		return
	}

	var style lipgloss.Style
	var prefix string
	switch {
	case index == m.Index():
		// todo
		style = d.colors.ListItemSelected()
		prefix = "► "
	default:
		style = d.colors.ListItemUnselected()
	}
	title := style.Render(prefix + modeItem.View())

	fmt.Fprintf(w, "%s", title)
}

type colorPickerKepMap struct {
	FlipBitdepth        key.Binding
	Back                key.Binding
	Accept              key.Binding
	Up                  key.Binding
	Down                key.Binding
	AdjustSdrBrightness key.Binding
	AdjustSdrSaturation key.Binding
}

func (s *colorPickerKepMap) Help(sdr bool) []key.Binding {
	keys := []key.Binding{
		s.FlipBitdepth,
		s.Back,
		s.Accept,
		s.Up,
		s.Down,
	}
	if sdr {
		keys = append(keys, s.AdjustSdrBrightness)
		keys = append(keys, s.AdjustSdrSaturation)
	}
	return keys
}

type ColorPicker struct {
	colorPreset   list.Model
	help          *CustomHelp
	width         int
	height        int
	monitor       *MonitorSpec
	keyMap        *colorPickerKepMap
	sdrBrightness *NumberAdjuster
	sdrSaturation *NumberAdjuster
	colors        *ColorsManager
}

func NewColorPicker(colors *ColorsManager) *ColorPicker {
	items := []list.Item{}
	delegate := NewColorPresetDelegate(colors)
	list := list.New(items, delegate, 0, 0)
	list.SetShowStatusBar(false)
	list.SetFilteringEnabled(false)
	list.SetShowHelp(false)
	list.SetShowTitle(false)

	return &ColorPicker{
		colorPreset:   list,
		help:          NewCustomHelp(colors),
		sdrBrightness: NewNumberAdjuster(0.0, 2.0, 0.5, 0.01),
		sdrSaturation: NewNumberAdjuster(0.0, 2.0, 0.5, 0.01),
		colors:        colors,
		keyMap: &colorPickerKepMap{
			Back: key.NewBinding(
				key.WithKeys("esc"),
				key.WithHelp("esc", "back"),
			),
			Accept: key.NewBinding(
				key.WithKeys("enter"),
				key.WithHelp("enter", "select"),
			),
			FlipBitdepth: key.NewBinding(
				key.WithKeys("b"),
				key.WithHelp("b", "next bitdepth"),
			),
			Up: key.NewBinding(
				key.WithKeys("up", "k"),
				key.WithHelp("up/k", "previous preset"),
			),
			Down: key.NewBinding(
				key.WithKeys("down", "j"),
				key.WithHelp("down/j", "next preset"),
			),
			AdjustSdrBrightness: key.NewBinding(
				key.WithKeys("r", "R"),
				key.WithHelp("r/R", "inc/dec sdr brightness"),
			),
			AdjustSdrSaturation: key.NewBinding(
				key.WithKeys("t", "T"),
				key.WithHelp("t/T", "inc/dec sdr saturation"),
			),
		},
	}
}

func (m *ColorPicker) Update(msg tea.Msg) tea.Cmd {
	cmds := []tea.Cmd{}
	// nolint:gocritic
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// nolint:gocritic
		switch {
		case key.Matches(msg, m.keyMap.FlipBitdepth):
			logrus.Debug("FlipBitdepth requested")
			cmds = append(cmds, nextBitdepthCmd(*m.monitor.ID))
		}

		// nolint:gocritic
		switch msg.String() {
		case "esc":
			logrus.Debug("Close color picker")
			return CloseColorPickerCmd()
		case "r":
			if m.monitor.ColorPreset.CanAdjustSdr() {
				m.sdrBrightness.Increase()
				cmds = append(cmds, AdjustSdrBrightnessCmd(*m.monitor.ID, m.sdrBrightness.Value()))
			}
		case "R":
			if m.monitor.ColorPreset.CanAdjustSdr() {
				m.sdrBrightness.Decrease()
				cmds = append(cmds, AdjustSdrBrightnessCmd(*m.monitor.ID, m.sdrBrightness.Value()))
			}
		case "t":
			if m.monitor.ColorPreset.CanAdjustSdr() {
				m.sdrSaturation.Increase()
				cmds = append(cmds, AdjustSdrSaturationCmd(*m.monitor.ID, m.sdrSaturation.Value()))
			}
		case "T":
			if m.monitor.ColorPreset.CanAdjustSdr() {
				m.sdrSaturation.Decrease()
				cmds = append(cmds, AdjustSdrSaturationCmd(*m.monitor.ID, m.sdrSaturation.Value()))
			}
		}
	}
	var cmd tea.Cmd
	m.colorPreset, cmd = m.colorPreset.Update(msg)
	cmds = append(cmds, cmd)

	return tea.Batch(cmds...)
}

func (m *ColorPicker) View() string {
	sdr := m.monitor.ColorPreset.CanAdjustSdr()
	sections := []string{}
	availableSpace := m.height
	logrus.Debugf("availableSpace for color picker: %d", availableSpace)

	titleText := m.colors.TitleStyle().Render("Adjust Colors: ")
	monitorName := m.monitor.Name
	title := lipgloss.NewStyle().Width(m.width).Margin(0, 0, 1, 0).Render(titleText + monitorName)
	sections = append(sections, title)
	availableSpace -= lipgloss.Height(title)
	logrus.Debugf("availableSpace for color picker: %d", availableSpace)

	subtitle := m.colors.InfoStyle().Width(m.width).Margin(0, 0, 1, 0).Render(
		"Note: Hyprctl does not support exposing a few of the following values, thus, the defaults shown might not reflect the current monitor state. See: https://github.com/fiffeek/hyprdynamicmonitors/issues/34")
	sections = append(sections, subtitle)
	availableSpace -= lipgloss.Height(subtitle)

	bitdepthtitle := m.colors.SubtitleStyle().Render("Bitdepth: ")
	bitdepthvalue := m.monitor.Bitdepth.Value()
	bitdepthline := lipgloss.NewStyle().Margin(0, 0, 1, 0).Width(m.width).Render(bitdepthtitle + bitdepthvalue)
	sections = append(sections, bitdepthline)
	availableSpace -= lipgloss.Height(bitdepthline)
	logrus.Debugf("availableSpace for color picker: %d", availableSpace)

	help := lipgloss.NewStyle().Width(m.width).Render(m.help.ShortHelpView(m.keyMap.Help(sdr)))
	availableSpace -= lipgloss.Height(help)
	logrus.Debugf("availableSpace for color picker: %d", availableSpace)

	presettitle := m.colors.SubtitleStyle().Render("Color Preset:")
	availableSpace -= lipgloss.Height(presettitle)
	sections = append(sections, presettitle)

	var sdrBrightnessLine string
	var sdrSaturationLine string
	if sdr {
		sdrBrightnessTitle := m.colors.SubtitleStyle().Render("SDR Brightness: ")
		sdrBrightnessValue := fmt.Sprintf("%.2f", m.monitor.SdrBrightness)
		sdrBrightnessLine = lipgloss.NewStyle().Width(m.width).Render(
			sdrBrightnessTitle + sdrBrightnessValue)
		availableSpace -= lipgloss.Height(sdrBrightnessLine)

		saturationTitle := m.colors.SubtitleStyle().Render("SDR Saturation: ")
		saturationValue := fmt.Sprintf("%.2f", m.monitor.SdrSaturation)
		sdrSaturationLine = lipgloss.NewStyle().Width(m.width).Render(saturationTitle + saturationValue)
		availableSpace -= lipgloss.Height(sdrSaturationLine)
	}

	colorPresetsHeight := min(availableSpace, 11)
	m.colorPreset.SetHeight(colorPresetsHeight)
	presets := lipgloss.NewStyle().Height(colorPresetsHeight).Render(m.colorPreset.View())
	availableSpace -= lipgloss.Height(presets)
	sections = append(sections, presets)
	logrus.Debugf("availableSpace for color picker: %d", availableSpace)

	if sdr {
		sections = append(sections, sdrBrightnessLine)
		sections = append(sections, sdrSaturationLine)
	}

	if availableSpace > 0 {
		spacer := lipgloss.NewStyle().Height(availableSpace).Render("")
		sections = append(sections, spacer)
	}
	sections = append(sections, help)

	return lipgloss.JoinVertical(lipgloss.Top, sections...)
}

func (m *ColorPicker) SetWidth(width int) {
	m.width = width
}

func (m *ColorPicker) SetHeight(height int) {
	m.height = height
}

func (m *ColorPicker) SetMonitor(monitor *MonitorSpec) tea.Cmd {
	m.monitor = monitor
	m.sdrBrightness.SetCurrent(m.monitor.SdrBrightness)
	m.sdrSaturation.SetCurrent(m.monitor.SdrSaturation)
	return m.SetItems(monitor)
}

func (m *ColorPicker) Unset() tea.Cmd {
	m.monitor = nil
	return nil
}

func (m *ColorPicker) SetItems(monitor *MonitorSpec) tea.Cmd {
	items := []list.Item{}
	selectedMode := -1
	for i, preset := range allColorPresets {
		items = append(items, colorPresetItem{preset: preset})
		if preset == monitor.ColorPreset {
			selectedMode = i
		}
	}

	cmd := m.colorPreset.SetItems(items)
	m.colorPreset.Select(selectedMode)
	return cmd
}
