package tui

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/fiffeek/hyprdynamicmonitors/internal/config"
	"github.com/fiffeek/hyprdynamicmonitors/internal/matchers"
	"github.com/fiffeek/hyprdynamicmonitors/internal/power"
	"github.com/sirupsen/logrus"
)

type HDMProfilePreview struct {
	cfg              *config.Config
	matcher          *matchers.Matcher
	pulled           bool
	profile          *matchers.MatchedProfile
	monitors         []*MonitorSpec
	textarea         textarea.Model
	height           int
	width            int
	powerState       power.PowerState
	runningUnderTest bool
	lidState         power.LidState
	colors           *ColorsManager
}

func NewHDMProfilePreview(cfg *config.Config,
	matcher *matchers.Matcher, monitors []*MonitorSpec, powerState power.PowerState,
	runningUnderTest bool, lidState power.LidState, colors *ColorsManager,
) *HDMProfilePreview {
	ta := textarea.New()
	return &HDMProfilePreview{
		cfg:              cfg,
		powerState:       powerState,
		lidState:         lidState,
		matcher:          matcher,
		pulled:           false,
		monitors:         monitors,
		textarea:         ta,
		runningUnderTest: runningUnderTest,
		colors:           colors,
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

	switch msg := msg.(type) {
	case ConfigReloaded:
		logrus.Debug("Received config reloaded event")
		h.pulled = false
		h.profile = nil
		h.textarea.SetValue("")
	case PowerStateChanged:
		logrus.Debug("Overriding the current power state")
		h.pulled = false
		h.powerState = msg.state
		h.profile = nil
		h.textarea.SetValue("")
	case LidStateChanged:
		logrus.Debug("Overriding the current lid state")
		h.pulled = false
		h.lidState = msg.state
		h.profile = nil
		h.textarea.SetValue("")
	}

	if !h.pulled {
		h.pulled = true

		mons, err := ConvertToHyprMonitors(h.monitors)
		if err != nil {
			cmds = append(cmds, OperationStatusCmd(OperationNameMatchingProfile, err))
		} else {
			ok, profile, err := h.matcher.Match(h.cfg.Get(), mons, h.powerState, h.lidState)
			cmds = append(cmds, OperationStatusCmd(OperationNameMatchingProfile, err))
			if ok {
				h.profile = profile
			}
		}
		if h.profile != nil {
			contents, err := os.ReadFile(h.profile.Profile.ConfigFile)
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

	title := h.colors.TitleStyle().Margin(0, 0, 0, 0).Render(
		fmt.Sprintf("Profile Config Preview (%s)", h.profile.Profile.ConfigType.Value()))
	availableHeight -= lipgloss.Height(title)
	sections = append(sections, title)

	// if running under test then just show the filename.
	// Since all configs are generated in tmp with random prefix directories,
	// it is easier to compare golden fixtures this way.
	configFile := h.profile.Profile.ConfigFile
	if h.runningUnderTest {
		configFile = filepath.Base(configFile)
	}

	subtitle := h.colors.MutedStyle().Margin(0, 0, 1, 0).Width(h.width).Render(configFile)
	availableHeight -= lipgloss.Height(subtitle)
	sections = append(sections, subtitle)

	h.textarea.SetHeight(availableHeight)
	ta := h.textarea.View()
	sections = append(sections, ta)

	return lipgloss.JoinVertical(lipgloss.Top, sections...)
}

func (h *HDMProfilePreview) GetPowerState() power.PowerState {
	return h.powerState
}

func (h *HDMProfilePreview) GetLidState() power.LidState {
	return h.lidState
}

func (h *HDMProfilePreview) GetProfile() *matchers.MatchedProfile {
	return h.profile
}

func (h *HDMProfilePreview) GetText() string {
	return h.textarea.Value()
}
