package tui

import (
	"errors"
	"fmt"
	"os/exec"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/fiffeek/hyprdynamicmonitors/internal/config"
	"github.com/fiffeek/hyprdynamicmonitors/internal/generators"
	"github.com/fiffeek/hyprdynamicmonitors/internal/matchers"
	"github.com/fiffeek/hyprdynamicmonitors/internal/power"
	"github.com/fiffeek/hyprdynamicmonitors/internal/profilemaker"
	"github.com/sirupsen/logrus"
)

type HyprApply struct {
	profileMaker *profilemaker.Service
	generator    *generators.ConfigGenerator
}

func NewHyprApply(profileMaker *profilemaker.Service, generator *generators.ConfigGenerator) *HyprApply {
	return &HyprApply{
		profileMaker: profileMaker,
		generator:    generator,
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

func (h *HyprApply) GenerateThroughHDM(cfg *config.Config, profile *matchers.MatchedProfile,
	monitors []*MonitorSpec, powerState power.PowerState, lidState power.LidState,
) tea.Cmd {
	if profile == nil {
		return OperationStatusCmd(OperationNameHydrate, errors.New("profile is nil"))
	}
	hyprMonitors, err := ConvertToHyprMonitors(monitors)
	if err != nil {
		return OperationStatusCmd(OperationNameHydrate, err)
	}
	destination := *cfg.Get().General.Destination
	_, err = h.generator.GenerateConfig(cfg.Get(), profile, hyprMonitors, powerState, lidState, destination)
	return OperationStatusCmd(OperationNameHydrate, err)
}
