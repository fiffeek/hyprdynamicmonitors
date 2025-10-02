package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
)

func TestConfirmationPrompt_Update(t *testing.T) {
	acceptedCalled := false
	rejectedCalled := false

	acceptedCmd := func() tea.Msg { acceptedCalled = true; return nil }
	rejectedCmd := func() tea.Msg { rejectedCalled = true; return nil }

	tests := []struct {
		name           string
		msg            tea.Msg
		expectAccepted bool
		expectRejected bool
	}{
		{
			name:           "Y key accepts",
			msg:            tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'Y'}},
			expectAccepted: true,
			expectRejected: false,
		},
		{
			name:           "n key rejects",
			msg:            tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}},
			expectAccepted: false,
			expectRejected: true,
		},
		{
			name:           "esc key rejects",
			msg:            tea.KeyMsg{Type: tea.KeyEsc},
			expectAccepted: false,
			expectRejected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			acceptedCalled = false
			rejectedCalled = false

			prompt := NewConfirmationPrompt("Test", acceptedCmd, rejectedCmd)
			cmd := prompt.Update(tt.msg)

			if cmd != nil {
				cmd()
			}

			assert.Equal(t, tt.expectAccepted, acceptedCalled)
			assert.Equal(t, tt.expectRejected, rejectedCalled)
		})
	}
}
