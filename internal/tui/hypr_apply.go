package tui

import (
	"fmt"
	"os/exec"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/fiffeek/hyprdynamicmonitors/internal/profilemaker"
	"github.com/sirupsen/logrus"
)

type HyprApply struct {
	profileMaker *profilemaker.Service
}

func NewHyprApply(profileMaker *profilemaker.Service) *HyprApply {
	return &HyprApply{
		profileMaker: profileMaker,
	}
}

func (h *HyprApply) ApplyCurrent(monitors []*MonitorSpec) tea.Cmd {
	var lastError error
	for _, monitor := range monitors {
		cmd := fmt.Sprintf("hyprctl keyword monitor \"%s\"", monitor.ToHypr())
		// nolint:gosec,noctx
		if err := exec.Command("sh", "-c", cmd).Run(); err != nil {
			lastError = err
			logrus.WithError(err).Error("cant apply hypr settings")
		}
	}

	return OperationStatusCmd(OperationNameEphemeralApply, lastError)
}

func (h *HyprApply) CreateProfile(monitors []*MonitorSpec, name, file string) tea.Cmd {
	hyprMonitors, err := ConvertToHyprMonitors(monitors)
	if err != nil {
		return OperationStatusCmd(OperationNameCreateProfile, err)
	}
	err = h.profileMaker.FreezeGivenAs(name, file, hyprMonitors)
	return OperationStatusCmd(OperationNameCreateProfile, err)
}

func (h *HyprApply) EditProfile(monitors []*MonitorSpec, name string) tea.Cmd {
	hyprMonitors, err := ConvertToHyprMonitors(monitors)
	if err != nil {
		return OperationStatusCmd(OperationNameEditProfile, err)
	}
	err = h.profileMaker.EditExisting(name, hyprMonitors)
	return OperationStatusCmd(OperationNameEditProfile, err)
}
