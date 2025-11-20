package tui

import (
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
)

type CustomHelp struct {
	colors *ColorsManager
}

func NewCustomHelp(colors *ColorsManager) *CustomHelp {
	return &CustomHelp{
		colors: colors,
	}
}

func (m *CustomHelp) ShortHelpView(bindings []key.Binding) string {
	model := help.New()
	model.Styles.ShortKey = m.colors.HelpKeyStyle()
	model.Styles.ShortDesc = m.colors.HelpDescriptionStyle()
	model.Styles.ShortSeparator = m.colors.HelpSeparatorStyle()
	return model.ShortHelpView(bindings)
}

func (m *CustomHelp) FullHelpView(groups [][]key.Binding) string {
	model := help.New()
	model.Styles.FullKey = m.colors.HelpKeyStyle()
	model.Styles.FullDesc = m.colors.HelpDescriptionStyle()
	model.Styles.FullSeparator = m.colors.HelpSeparatorStyle()
	return model.FullHelpView(groups)
}
