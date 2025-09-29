package tui

import (
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/fiffeek/hyprdynamicmonitors/internal/config"
	"github.com/fiffeek/hyprdynamicmonitors/internal/matchers"
	"github.com/fiffeek/hyprdynamicmonitors/internal/power"
	"github.com/sirupsen/logrus"
)

type HDMProfilePreview struct {
	cfg      *config.Config
	matcher  *matchers.Matcher
	pulled   bool
	profile  *config.Profile
	monitors []*MonitorSpec
	textarea textarea.Model
	height   int
	width    int
}

func NewHDMProfilePreview(cfg *config.Config, matcher *matchers.Matcher, monitors []*MonitorSpec) *HDMProfilePreview {
	ta := textarea.New()
	return &HDMProfilePreview{
		cfg:      cfg,
		matcher:  matcher,
		pulled:   false,
		monitors: monitors,
		textarea: ta,
	}
}

func (h *HDMProfilePreview) SetHeight(height int) {
	h.height = height
	h.textarea.SetHeight(height)
}

func (h *HDMProfilePreview) SetWidth(width int) {
	h.width = width
	h.textarea.SetWidth(width)
}

func (h *HDMProfilePreview) Update(msg tea.Msg) tea.Cmd {
	cmds := []tea.Cmd{}

	switch msg.(type) {
	case ConfigReloaded:
		logrus.Debug("Received config reloaded event")
		h.pulled = false
		h.profile = nil
		h.textarea.SetValue("")
	}

	if !h.pulled {
		h.pulled = true
		_, profile, err := h.matcher.Match(h.cfg.Get(), ConvertToHyprMonitors(h.monitors), power.ACPowerState)
		cmds = append(cmds, operationStatusCmd(OperationNameMatchingProfile, err))
		h.profile = profile
		if h.profile != nil {
			contents, err := os.ReadFile(h.profile.ConfigFile)
			text := "Can't pull config"
			if err == nil {
				text = string(contents)
			}
			logrus.Debugf("Textarea text: %s", text)
			h.textarea.SetValue(text)
		}

	}

	return tea.Batch(cmds...)
}

func (h *HDMProfilePreview) View() string {
	if h.profile == nil {
		return "No profile config"
	}

	sections := []string{}
	availableHeight := h.height

	title := TitleStyle.Margin(0, 0, 0, 0).Render(
		fmt.Sprintf("Profile Config Preview (%s)", h.profile.ConfigType.Value()))
	availableHeight -= lipgloss.Height(title)
	sections = append(sections, title)

	subtitle := SubtitleInfoStyle.Margin(0, 0, 1, 0).Render(h.profile.ConfigFile)
	availableHeight -= lipgloss.Height(subtitle)
	sections = append(sections, subtitle)

	h.textarea.SetHeight(availableHeight)
	ta := h.textarea.View()
	sections = append(sections, ta)

	return lipgloss.JoinVertical(lipgloss.Top, sections...)
}
