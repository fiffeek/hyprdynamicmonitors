package tui

import (
	"errors"
	"fmt"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/sirupsen/logrus"
)

type ProfileNamePicker struct {
	textInput textinput.Model
	width     int
	height    int
}

func NewProfileNamePicker() *ProfileNamePicker {
	ti := textinput.New()
	ti.Placeholder = "New Profile Name"
	ti.Focus()
	ti.CharLimit = 156
	ti.Width = 20

	return &ProfileNamePicker{
		textInput: ti,
	}
}

func (p *ProfileNamePicker) SetWidth(width int) {
	p.width = width
	p.textInput.Width = width - 5
}

func (p *ProfileNamePicker) SetHeight(height int) {
	p.height = height
}

func (p *ProfileNamePicker) Update(msg tea.Msg) tea.Cmd {
	cmds := []tea.Cmd{}
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			cmds = append(cmds, profileNameToogled())
		case "enter":
			text := p.textInput.Value()
			// todo validate profile name, no spaces etc, check other profiles here from cfg
			if text == "" {
				cmds = append(cmds, operationStatusCmd(OperationNameCreateProfile, errors.New("cant create empty name profile")))
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
	title := TitleStyle.Margin(0, 0, 1, 0).Render("Type the profile name")
	ti := p.textInput.View()
	sections = append(sections, title)
	sections = append(sections, ti)
	return lipgloss.JoinVertical(lipgloss.Top, sections...)
}
