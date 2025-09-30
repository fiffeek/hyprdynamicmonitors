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
	NewProfile   key.Binding
	ApplyProfile key.Binding
	EditorEdit   key.Binding
}

func (h *hdmKeyMap) Help() []key.Binding {
	return []key.Binding{
		h.NewProfile,
		h.ApplyProfile,
		h.EditorEdit,
	}
}

type HDMConfigPane struct {
	cfg           *config.Config
	matcher       *matchers.Matcher
	monitors      []*MonitorSpec
	keymap        *hdmKeyMap
	profile       *config.Profile
	pulledProfile bool
	help          help.Model
	height        int
	width         int
	powerState    power.PowerState
}

func NewHDMConfigPane(cfg *config.Config, matcher *matchers.Matcher, monitors []*MonitorSpec, powerState power.PowerState) *HDMConfigPane {
	return &HDMConfigPane{
		cfg:        cfg,
		matcher:    matcher,
		monitors:   monitors,
		powerState: powerState,
		help:       help.New(),
		keymap: &hdmKeyMap{
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
	}

	if !h.pulledProfile {
		logrus.Debugf("Current power state: %d", h.powerState)
		h.pulledProfile = true
		_, profile, err := h.matcher.Match(h.cfg.Get(), ConvertToHyprMonitors(h.monitors), h.powerState)
		cmds = append(cmds, operationStatusCmd(OperationNameMatchingProfile, err))
		h.profile = profile
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
			cmds = append(cmds, editProfileConfirmationCmd(h.profile.Name))
		case key.Matches(msg, h.keymap.EditorEdit):
			cmds = append(cmds, openEditor(h.profile.ConfigFile))
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
		content = h.renderMatchedProfile(h.profile)
	}
	availableHeight -= lipgloss.Height(content)
	availableHeight -= lipgloss.Height(help)
	availableHeight -= lipgloss.Height(title)

	spacer := lipgloss.NewStyle().Width(h.width).Height(availableHeight).Render("")

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
	var result strings.Builder

	result.WriteString(HyprConfigTitleStyle.Render("No Matching Profile"))
	result.WriteString("\n\n")

	result.WriteString("No configuration profile matches the current monitor setup.\n\n")

	actionContent := fmt.Sprintf("Press %s to create a new profile",
		HeaderIndicatorStyle.Render("'n'"))
	actionBox := configPaneActionStyle.Render(actionContent)
	result.WriteString(actionBox)

	return result.String()
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
		statusContent := "New settings will be appended when saved"
		statusBox := configPaneBorderStyle.
			BorderForeground(lipgloss.Color("39")).
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

	if profile.Conditions != nil && profile.Conditions.RequiredMonitors != nil {
		content.WriteString(configDetailLabelStyle.Render("Required Monitors:"))
		content.WriteString("\n")
		for _, monitor := range profile.Conditions.RequiredMonitors {
			content.WriteString(h.renderRequiredMonitor(monitor))
		}
	}

	return configPaneBorderStyle.Render(content.String())
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

func (h *HDMConfigPane) GetProfile() *config.Profile {
	return h.profile
}
