package tui

import (
	"errors"
	"fmt"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/sirupsen/logrus"
)

type profileNamePickerKeyMap struct {
	Create key.Binding
	Back   key.Binding
}

func (c *profileNamePickerKeyMap) Help() []key.Binding {
	return []key.Binding{
		c.Create,
		c.Back,
	}
}

type ProfileNamePicker struct {
	textInput textinput.Model
	width     int
	height    int
	help      help.Model
	keyMap    *profileNamePickerKeyMap
}

func NewProfileNamePicker() *ProfileNamePicker {
	ti := textinput.New()
	ti.Placeholder = "New Profile Name"
	ti.Focus()
	ti.CharLimit = 20
	ti.Width = 20

	return &ProfileNamePicker{
		textInput: ti,
		help:      help.New(),
		keyMap: &profileNamePickerKeyMap{
			Create: key.NewBinding(
				key.WithKeys("enter"),
				key.WithHelp("enter", "create profile"),
			),
			Back: key.NewBinding(
				key.WithKeys("esc"),
				key.WithHelp("esc", "return/back"),
			),
		},
	}
}

func (p *ProfileNamePicker) SetWidth(width int) {
	p.width = width
	p.textInput.Width = width - 5
	p.textInput.CharLimit = width - 5
}

func (p *ProfileNamePicker) SetHeight(height int) {
	p.height = height
}

func (p *ProfileNamePicker) Update(msg tea.Msg) tea.Cmd {
	cmds := []tea.Cmd{}
	// nolint:gocritic
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, p.keyMap.Back):
			cmds = append(cmds, profileNameToogled())
		case key.Matches(msg, p.keyMap.Create):
			text := p.textInput.Value()
			// todo validate profile name, no spaces etc, check other profiles here from cfg
			if text == "" {
				cmds = append(cmds, operationStatusCmd(OperationNameCreateProfile,
					errors.New("cant create empty name profile")))
			} else {
				p.textInput.SetValue("")
				logrus.Debugf("Setting name to: %s", text)
				file := fmt.Sprintf("hyprconfigs/%s.go.tmpl", text)
				cmds = append(cmds, createNewProfileCmd(text, file))
			}
		}
	}

	textInput, cmd := p.textInput.Update(msg)
	p.textInput = textInput
	cmds = append(cmds, cmd)

	return tea.Batch(cmds...)
}

func (p *ProfileNamePicker) View() string {
	sections := []string{}
	availableSpace := p.height
	logrus.Debugf("availableSpace for ProfileNamePicker: %d", availableSpace)

	title := TitleStyle.Margin(0, 0, 1, 0).Render("Type the profile name")
	sections = append(sections, title)
	availableSpace -= lipgloss.Height(title)

	ti := p.textInput.View()
	sections = append(sections, ti)
	availableSpace -= lipgloss.Height(ti)

	help := p.help.ShortHelpView(p.keyMap.Help())
	availableSpace -= lipgloss.Height(help)

	logrus.Debugf("spacer height: %d", availableSpace)

	spacer := lipgloss.NewStyle().Height(availableSpace).Render("")
	sections = append(sections, spacer)
	sections = append(sections, help)

	return lipgloss.JoinVertical(lipgloss.Top, sections...)
}
