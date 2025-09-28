package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/sirupsen/logrus"
)

type Header struct {
	title   string
	warning string
	mode    string
	width   int
	err     string
}

func NewHeader(title string) *Header {
	return &Header{
		title:   title,
		warning: "",
		mode:    "",
		width:   0,
	}
}

func (h *Header) Update(msg tea.Msg) tea.Cmd {
	logrus.Debugf("Header received message: %v", msg)
	switch msg := msg.(type) {
	case StateChanged:
		h.mode = msg.state.String()
	case OperationStatus:
		if msg.IsError() {
			h.err = msg.String()
		} else {
			h.err = ""
		}
	}
	return nil
}

func (h *Header) View() string {
	sections := []string{}
	availableSpace := h.width
	logrus.Debugf("Available header space: %d", availableSpace)

	header := HeaderStyle.Render(h.title)
	availableSpace -= lipgloss.Width(header)
	sections = append(sections, header)

	var mode string
	if h.mode != "" {
		mode = HeaderIndicatorStyle.Render(h.mode)
		availableSpace -= lipgloss.Width(mode)
	}

	var status string
	if h.err != "" {
		status = ErrorStyle.Render(h.err)
		availableSpace -= lipgloss.Width(status)
	}

	spacer := lipgloss.NewStyle().Width(availableSpace).Render("")
	sections = append(sections, spacer)

	if h.err != "" {
		sections = append(sections, status)
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
