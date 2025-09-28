package tui

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/fiffeek/hyprdynamicmonitors/internal/config"
	"github.com/fiffeek/hyprdynamicmonitors/internal/hypr"
	"github.com/fiffeek/hyprdynamicmonitors/internal/matchers"
	"github.com/fiffeek/hyprdynamicmonitors/internal/power"
	"github.com/sirupsen/logrus"
)

type hdmKeyMap struct {
	NewProfile   key.Binding
	ApplyProfile key.Binding
}

type HDMConfigPane struct {
	cfg          *config.Config
	matcher      *matchers.Matcher
	monitors     []*MonitorSpec
	keymap       *hdmKeyMap
	profile      *config.Profile
	pulldProfile bool
}

func NewHDMConfigPane(cfg *config.Config, matcher *matchers.Matcher, monitors []*MonitorSpec) *HDMConfigPane {
	return &HDMConfigPane{
		cfg:      cfg,
		matcher:  matcher,
		monitors: monitors,
		keymap: &hdmKeyMap{
			NewProfile: key.NewBinding(
				key.WithKeys("n"),
				key.WithHelp("n", "new config"),
			),
			ApplyProfile: key.NewBinding(
				key.WithKeys("a"),
				key.WithHelp("e", "apply to existing config"),
			),
		},
	}
}

func (h *HDMConfigPane) Update(msg tea.Msg) tea.Cmd {
	cmds := []tea.Cmd{}

	if !h.pulldProfile {
		h.pulldProfile = true
		_, profile, err := h.matcher.Match(h.cfg.Get(), h.convertToHyprMonitors(), power.ACPowerState)
		cmds = append(cmds, operationStatusCmd(OperationNameMatchingProfile, err))
		h.profile = profile
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, h.keymap.NewProfile):
			logrus.Debug("Creating a new config")
			cmds = append(cmds, createNewProfileCmd("hello", "hyprconfigs/test.go.tmpl"))
		case key.Matches(msg, h.keymap.ApplyProfile):
			logrus.Debug("Editing existing config")
			cmds = append(cmds, editProfileCmd(h.profile.Name))
		}
	}
	return tea.Batch(cmds...)
}

func (h *HDMConfigPane) View() string {
	if h.cfg == nil {
		return h.renderNoConfig()
	}

	if h.profile == nil {
		return h.renderNoMatchingProfile()
	}

	return h.renderMatchedProfile(h.profile)
}

func (h *HDMConfigPane) renderNoConfig() string {
	return "No config"
}

func (h *HDMConfigPane) convertToHyprMonitors() []*hypr.MonitorSpec {
	var hyprMonitors []*hypr.MonitorSpec
	for _, monitor := range h.monitors {
		hyprMonitors = append(hyprMonitors, monitor.ToHyprMonitors())
	}
	return hyprMonitors
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
		fmt.Sprintf("Matched Profile: %s", profile.Name)))
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
			fmt.Sprintf("%v", profile.ConfigType.Value())))
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
		parts = append(parts, fmt.Sprintf("Name: %s", *monitor.Name))
	}
	if monitor.Description != nil {
		parts = append(parts, fmt.Sprintf("Desc: %s",
			h.truncateString(*monitor.Description, 25)))
	}
	if monitor.MonitorTag != nil {
		parts = append(parts, fmt.Sprintf("Tag: %s", *monitor.MonitorTag))
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
