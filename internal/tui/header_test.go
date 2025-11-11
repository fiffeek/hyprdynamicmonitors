package tui_test

import (
	"errors"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/fiffeek/hyprdynamicmonitors/internal/testutils"
	"github.com/fiffeek/hyprdynamicmonitors/internal/tui"
	"github.com/stretchr/testify/assert"
)

func TestHeader_Update(t *testing.T) {
	testErr := errors.New("test error")

	tests := []struct {
		name            string
		setupMsg        tea.Cmd
		msg             tea.Msg
		expectedMode    string
		expectedView    tui.ViewMode
		expectedErr     string
		expectedSuccess string
	}{
		{
			name:            "StateChanged updates mode",
			setupMsg:        tui.StateChangedCmd(tui.AppState{Panning: true}),
			expectedMode:    "Panning",
			expectedView:    tui.ViewMode(0),
			expectedErr:     "",
			expectedSuccess: "",
		},
		{
			name:            "ViewChanged updates current view",
			setupMsg:        tui.ViewChangedCmd(tui.ProfileView),
			expectedMode:    "",
			expectedView:    tui.ProfileView,
			expectedErr:     "",
			expectedSuccess: "",
		},
		{
			name:            "OperationStatus with error updates error",
			setupMsg:        tui.OperationStatusCmd(tui.OperationNameMatchingProfile, testErr),
			expectedMode:    "",
			expectedView:    tui.ViewMode(0),
			expectedErr:     "Reload Profile: test error",
			expectedSuccess: "",
		},
		{
			name:            "OperationStatus success with showSuccessToUser",
			setupMsg:        tui.OperationStatusCmd(tui.OperationNameCreateProfile, nil),
			expectedMode:    "",
			expectedView:    tui.ViewMode(0),
			expectedErr:     "",
			expectedSuccess: "Create Profile: success",
		},
		{
			name:            "OperationStatus success without showSuccessToUser",
			setupMsg:        tui.OperationStatusCmd(tui.OperationNameMove, nil),
			expectedMode:    "",
			expectedView:    tui.ViewMode(0),
			expectedErr:     "",
			expectedSuccess: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := testutils.NewTestConfig(t).Get()
			colors := tui.NewColorsManager(cfg)
			header := tui.NewHeader("test", []tui.ViewMode{tui.MonitorsListView}, "1.0.0", colors)

			if tt.setupMsg != nil {
				tt.msg = tt.setupMsg()
			}

			header.Update(tt.msg)

			assert.Equal(t, tt.expectedMode, header.GetMode())
			assert.Equal(t, tt.expectedView, header.GetCurrentView())
			assert.Equal(t, tt.expectedErr, header.GetError())
			assert.Equal(t, tt.expectedSuccess, header.GetSuccess())
		})
	}
}
