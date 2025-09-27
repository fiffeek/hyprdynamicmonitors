package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/fiffeek/hyprdynamicmonitors/internal/utils"
	"github.com/sirupsen/logrus"
)

type Header struct {
	title   string
	warning *string
	mode    *string
	width   int
}

func NewHeader(title string) *Header {
	return &Header{
		title:   title,
		warning: nil,
		mode:    nil,
		width:   0,
	}
}

func (h *Header) Update(msg tea.Msg) tea.Cmd {
	logrus.Debugf("Header received message: %v", msg)
	switch msg := msg.(type) {
	case StateChanged:
		if msg.state == StateNavigating {
			h.mode = nil
			return nil
		}
		h.mode = utils.StringPtr(msg.state.String())
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
	if h.mode != nil {
		mode = HeaderIndicatorStyle.Render(*h.mode)
		availableSpace -= lipgloss.Width(mode)
	}

	spacer := lipgloss.NewStyle().Width(availableSpace).Render("")
	sections = append(sections, spacer)

	if h.mode != nil {
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
