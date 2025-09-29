package tui

import (
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/sirupsen/logrus"
)

type confirmKeyMap struct {
	Accept key.Binding
	Reject key.Binding
	Back   key.Binding
}

func (c *confirmKeyMap) Help() []key.Binding {
	return []key.Binding{
		c.Accept,
		c.Reject,
		c.Back,
	}
}

type ConfirmationPrompt struct {
	accepted tea.Cmd
	rejected tea.Cmd
	keys     confirmKeyMap
	title    string
	help     help.Model
	width    int
	height   int
}

func NewConfirmationPrompt(title string, accepted, rejected tea.Cmd) *ConfirmationPrompt {
	return &ConfirmationPrompt{
		title:    title,
		accepted: accepted,
		rejected: rejected,
		keys: confirmKeyMap{
			Accept: key.NewBinding(
				key.WithKeys("Y"),
				key.WithHelp("Y", "yes"),
			),
			Reject: key.NewBinding(
				key.WithKeys("n", "N"),
				key.WithHelp("n/N", "no"),
			),
			Back: key.NewBinding(
				key.WithKeys("esc"),
				key.WithHelp("esc", "back"),
			),
		},
		help: help.New(),
	}
}

func (c *ConfirmationPrompt) Update(msg tea.Msg) tea.Cmd {
	cmds := []tea.Cmd{}
	// nolint:gocritic
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, c.keys.Accept):
			logrus.Debug("Confirmation prompt accepted")
			cmds = append(cmds, c.accepted)

		case key.Matches(msg, c.keys.Reject):
			logrus.Debug("Confirmation prompt rejected")
			cmds = append(cmds, c.rejected)

		case key.Matches(msg, c.keys.Back):
			logrus.Debug("Confirmation prompt back")
			cmds = append(cmds, c.rejected)
		}
	}
	return tea.Batch(cmds...)
}

func (c *ConfirmationPrompt) SetHeight(height int) {
	c.height = height
}

func (c *ConfirmationPrompt) SetWidth(width int) {
	c.width = width
}

func (c *ConfirmationPrompt) View() string {
	title := c.title
	help := c.help.ShortHelpView(c.keys.Help())
	view := lipgloss.JoinVertical(lipgloss.Top, title, help)

	return lipgloss.Place(c.width, c.height, lipgloss.Center, lipgloss.Center, view)
}
