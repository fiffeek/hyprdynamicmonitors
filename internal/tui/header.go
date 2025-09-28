package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/sirupsen/logrus"
)

type Header struct {
	title          string
	warning        string
	mode           string
	width          int
	err            string
	success        string
	availableViews []ViewMode
	currentView    ViewMode
}

func NewHeader(title string, availableViews []ViewMode) *Header {
	return &Header{
		title:          title,
		warning:        "",
		mode:           "",
		width:          0,
		success:        "",
		availableViews: availableViews,
	}
}

func (h *Header) Update(msg tea.Msg) tea.Cmd {
	logrus.Debugf("Header received message: %v", msg)
	switch msg := msg.(type) {
	case StateChanged:
		h.mode = msg.state.String()
	case ViewChanged:
		h.currentView = msg.view
	case OperationStatus:
		if msg.IsError() {
			h.err = msg.String()
			h.success = ""
		} else {
			h.err = ""
			if msg.showSuccessToUser {
				h.success = msg.String()
			} else {
				h.success = ""
			}
		}
	}
	return nil
}

func (h *Header) View() string {
	sections := []string{}
	availableSpace := h.width
	logrus.Debugf("Available header space: %d", availableSpace)

	programName := HeaderStyle.Foreground(lipgloss.Color("0")).Background(lipgloss.Color("250")).Padding(0, 1).Render(h.title)
	availableSpace -= lipgloss.Width(programName)
	sections = append(sections, programName)

	if len(h.availableViews) > 1 {
		tabs := h.renderTabs()
		availableSpace -= lipgloss.Width(tabs)
		sections = append(sections, tabs)
	}

	var mode string
	if h.mode != "" {
		mode = HeaderIndicatorStyle.Render(h.mode)
		availableSpace -= lipgloss.Width(mode)
	}

	var statusError string
	if h.err != "" {
		statusError = ErrorStyle.Render(h.err)
		availableSpace -= lipgloss.Width(statusError)
	}

	var statusSuccess string
	if h.success != "" {
		statusSuccess = SuccessStyle.Render(h.success)
		availableSpace -= lipgloss.Width(statusSuccess)
	}

	spacer := lipgloss.NewStyle().Width(availableSpace).Render("")
	sections = append(sections, spacer)

	if h.success != "" {
		sections = append(sections, statusSuccess)
	}

	if h.err != "" {
		sections = append(sections, statusError)
	}

	if h.mode != "" {
		sections = append(sections, mode)
	}

	return lipgloss.JoinHorizontal(
		lipgloss.Left,
		sections...,
	)
}

func (h *Header) SetWidth(width int) {
	h.width = width
}

func (h *Header) renderTabs() string {
	tabs := []string{}
	for _, view := range h.availableViews {
		var tab string
		if view == h.currentView {
			tab = TabActiveStyle.Render(view.String())
		} else {
			tab = TabInactiveStyle.Render(view.String())
		}
		tabs = append(tabs, tab)
	}
	return lipgloss.JoinHorizontal(lipgloss.Left, tabs...)
}
