package tui

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/fiffeek/hyprdynamicmonitors/internal/config"
	"github.com/fiffeek/hyprdynamicmonitors/internal/matchers"
	"github.com/fiffeek/hyprdynamicmonitors/internal/power"
	"github.com/sirupsen/logrus"
)

type hdmKeyMap struct {
	NewProfile    key.Binding
	ApplyProfile  key.Binding
	EditorEdit    key.Binding
	RenderProfile key.Binding
}

func (h *hdmKeyMap) Help() []key.Binding {
	return []key.Binding{
		h.NewProfile,
		h.ApplyProfile,
		h.EditorEdit,
		h.RenderProfile,
	}
}

type HDMConfigPane struct {
	cfg           *config.Config
	matcher       *matchers.Matcher
	monitors      []*MonitorSpec
	keymap        *hdmKeyMap
	profile       *matchers.MatchedProfile
	pulledProfile bool
	help          help.Model
	height        int
	width         int
	powerState    power.PowerState
	lidState      power.LidState
}

func NewHDMConfigPane(cfg *config.Config, matcher *matchers.Matcher, monitors []*MonitorSpec,
	powerState power.PowerState, lidState power.LidState,
) *HDMConfigPane {
	return &HDMConfigPane{
		cfg:        cfg,
		matcher:    matcher,
		monitors:   monitors,
		powerState: powerState,
		lidState:   lidState,
		help:       help.New(),
		keymap: &hdmKeyMap{
			RenderProfile: key.NewBinding(
				key.WithKeys("R"),
				key.WithHelp("R", "render profile to config.general.destination"),
			),
			NewProfile: key.NewBinding(
				key.WithKeys("n"),
				key.WithHelp("n", "new profile"),
			),
			ApplyProfile: key.NewBinding(
				key.WithKeys("a"),
				key.WithHelp("a", "apply monitors to existing profile"),
			),
			EditorEdit: key.NewBinding(
				key.WithKeys("e"),
				key.WithHelp("e", "edit manually"),
			),
		},
	}
}

func (h *HDMConfigPane) Update(msg tea.Msg) tea.Cmd {
	cmds := []tea.Cmd{}

	switch msg := msg.(type) {
	case ConfigReloaded:
		logrus.Debug("Received config reloaded event")
		h.pulledProfile = false
		h.profile = nil
	case PowerStateChanged:
		logrus.Debug("Overriding the current power state")
		h.powerState = msg.state
		h.pulledProfile = false
		h.profile = nil
	case LidStateChanged:
		logrus.Debug("Overriding the current lid state")
		h.pulledProfile = false
		h.lidState = msg.state
		h.profile = nil
	}

	if !h.pulledProfile {
		logrus.Debugf("Current power state: %d", h.powerState)
		h.pulledProfile = true
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
	}

	switch msg := msg.(type) {
	case CreateNewProfileCommand:
		logrus.Debug("Received create new profile")
		cmds = append(cmds, profileNameToogled())
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, h.keymap.NewProfile):
			logrus.Debug("Creating a new config")
			cmds = append(cmds, profileNameToogled())
		case key.Matches(msg, h.keymap.ApplyProfile):
			logrus.Debug("Editing existing config")
			cmds = append(cmds, editProfileConfirmationCmd(h.profile.Profile.Name))
		case key.Matches(msg, h.keymap.EditorEdit):
			cmds = append(cmds, openEditor(h.profile.Profile.ConfigFile))
		case key.Matches(msg, h.keymap.RenderProfile):
			cmds = append(cmds, RenderHDMConfigCmd(h.profile, h.lidState, h.powerState))
		}
	}
	return tea.Batch(cmds...)
}

func (h *HDMConfigPane) SetHeight(height int) {
	h.height = height
}

func (h *HDMConfigPane) SetWidth(width int) {
	h.width = width
}

func (h *HDMConfigPane) View() string {
	logrus.Debugf("Available height %d", h.height)
	if h.cfg == nil {
		return h.renderNoConfig()
	}
	availableHeight := h.height
	sections := []string{}

	title := TitleStyle.Margin(0, 0, 1, 0).Render("HyprDynamicMonitors Profile")

	var content string
	help := HelpStyle.Width(h.width).Render(h.help.ShortHelpView(h.keymap.Help()))

	if h.profile == nil {
		content = h.renderNoMatchingProfile()
	} else {
		content = h.renderMatchedProfile(h.profile.Profile)
	}
	availableHeight -= lipgloss.Height(content)
	logrus.Debugf("Height of content: %d", lipgloss.Height(content))
	availableHeight -= lipgloss.Height(help)
	logrus.Debugf("Height of help: %d", lipgloss.Height(help))
	availableHeight -= lipgloss.Height(title)
	logrus.Debugf("Height of title: %d", lipgloss.Height(title))

	spacer := lipgloss.NewStyle().Width(h.width).Height(availableHeight).Render("")
	logrus.Debugf("Height of spacer: %d", lipgloss.Height(spacer))

	sections = append(sections, title)
	sections = append(sections, content)
	sections = append(sections, spacer)
	sections = append(sections, help)

	return lipgloss.JoinVertical(lipgloss.Top, sections...)
}

func (h *HDMConfigPane) renderNoConfig() string {
	return "No config"
}

func (h *HDMConfigPane) renderNoMatchingProfile() string {
	sections := []string{}

	sections = append(sections, HyprConfigTitleStyle.Width(h.width).Margin(0, 0, 1, 0).Render("No Matching Profile"))

	content := "No configuration profile matches the current monitor setup."
	content = lipgloss.NewStyle().Margin(0, 0, 2, 0).Width(h.width).Render(content)
	sections = append(sections, content)

	actionContent := fmt.Sprintf("Press %s to create a new profile",
		HeaderIndicatorStyle.Render("'n'"))
	actionBox := configPaneActionStyle.Render(actionContent)
	sections = append(sections, actionBox)

	return lipgloss.JoinVertical(lipgloss.Top, sections...)
}

func (h *HDMConfigPane) renderMatchedProfile(profile *config.Profile) string {
	var result strings.Builder

	result.WriteString(HyprConfigTitleStyle.Render(
		"Profile: " + profile.Name))
	result.WriteString("\n\n")

	if h.hasMonitorCountMismatch(profile) {
		result.WriteString(h.renderMonitorMismatchWarning(profile))
		result.WriteString("\n\n")
	}

	result.WriteString(h.renderProfileDetails(profile))

	if !h.hasMonitorCountMismatch(profile) {
		result.WriteString("\n")
		statusContent := "New settings will be appended (or replaced depending on the `<<<<<` markers position) when applied (type `a`).\n\nIf you're not running the daemon you likely want to follow up by rendering the profile into the hypr settings destination (type `R`)."
		statusBox := SubtitleInfoStyle.Width(h.width).
			Render(statusContent)
		result.WriteString(statusBox)
	}

	return result.String()
}

func (h *HDMConfigPane) hasMonitorCountMismatch(profile *config.Profile) bool {
	requiredCount := h.getRequiredMonitorCount(profile)
	currentCount := len(h.monitors)
	return requiredCount != currentCount
}

func (h *HDMConfigPane) getRequiredMonitorCount(profile *config.Profile) int {
	if profile.Conditions != nil && profile.Conditions.RequiredMonitors != nil {
		return len(profile.Conditions.RequiredMonitors)
	}
	return 0
}

func (h *HDMConfigPane) renderMonitorMismatchWarning(profile *config.Profile) string {
	requiredCount := h.getRequiredMonitorCount(profile)
	currentCount := len(h.monitors)

	var content strings.Builder
	content.WriteString(ErrorStyle.Render("⚠ Monitor Count Mismatch"))
	content.WriteString("\n")
	content.WriteString(MutedStyle.Render(fmt.Sprintf("Expected: %d monitors, Connected: %d", requiredCount, currentCount)))
	content.WriteString(MutedStyle.Render("\nConsider creating a new profile (press 'n')"))

	return configPaneWarningStyle.Render(content.String())
}

func (h *HDMConfigPane) renderProfileDetails(profile *config.Profile) string {
	var content strings.Builder

	content.WriteString(HeaderStyle.Render("Profile Details"))
	content.WriteString("\n")

	content.WriteString(configDetailLabelStyle.Render("Config File: "))
	content.WriteString(configDetailStyle.Render(
		h.truncateString(filepath.Base(profile.ConfigFile), 40)))
	content.WriteString("\n")

	if profile.ConfigType != nil {
		content.WriteString(configDetailLabelStyle.Render("Config Type: "))
		content.WriteString(configDetailStyle.Render(
			profile.ConfigType.Value()))
		content.WriteString("\n")
	}

	if profile.Conditions != nil && profile.Conditions.PowerState != nil {
		content.WriteString(configDetailLabelStyle.Render("Power State: "))
		content.WriteString(configDetailStyle.Render(profile.Conditions.PowerState.Value()))
		content.WriteString("\n")
	}

	if profile.Conditions != nil && profile.Conditions.LidState != nil {
		content.WriteString(configDetailLabelStyle.Render("Lid State: "))
		content.WriteString(configDetailStyle.Render(profile.Conditions.LidState.Value()))
		content.WriteString("\n")
	}

	if profile.Conditions != nil && profile.Conditions.RequiredMonitors != nil {
		content.WriteString(configDetailLabelStyle.Render("Required Monitors:"))
		content.WriteString("\n")
		for _, monitor := range profile.Conditions.RequiredMonitors {
			content.WriteString(h.renderRequiredMonitor(monitor))
		}
	}

	return configPaneBorderStyle.Width(h.width).Render(content.String())
}

func (h *HDMConfigPane) renderRequiredMonitor(monitor *config.RequiredMonitor) string {
	var parts []string
	if monitor.Name != nil {
		parts = append(parts, "Name: "+*monitor.Name)
	}
	if monitor.Description != nil {
		parts = append(parts, "Desc: "+
			h.truncateString(*monitor.Description, 25))
	}
	if monitor.MonitorTag != nil {
		parts = append(parts, "Tag: "+*monitor.MonitorTag)
	}

	content := "(no constraints)"
	if len(parts) > 0 {
		content = strings.Join(parts, " | ")
	}

	return "  • " + content + "\n"
}

func (h *HDMConfigPane) truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func (h *HDMConfigPane) GetPowerState() power.PowerState {
	return h.powerState
}

func (h *HDMConfigPane) GetLidState() power.LidState {
	return h.lidState
}

func (h *HDMConfigPane) GetProfile() *matchers.MatchedProfile {
	return h.profile
}
