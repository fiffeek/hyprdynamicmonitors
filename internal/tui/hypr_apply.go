package tui

import (
	"fmt"
	"os/exec"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/sirupsen/logrus"
)

type HyprApply struct{}

func NewHyprApply() *HyprApply {
	return &HyprApply{}
}

func (h *HyprApply) ApplyCurrent(monitors []*MonitorSpec) tea.Cmd {
	var lastError error
	for _, monitor := range monitors {
		cmd := fmt.Sprintf("hyprctl keyword monitor \"%s\"", monitor.ToHypr())
		if err := exec.Command("sh", "-c", cmd).Run(); err != nil {
			lastError = err
			logrus.WithError(err).Error("cant apply hypr settings")
		}
	}

	return operationStatusCmd(OperationNameEphemeralApply, lastError)
}
