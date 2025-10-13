package tui

import (
	"os"
	"os/exec"
	"slices"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/fiffeek/hyprdynamicmonitors/internal/power"
)

type MonitorBeingEdited struct {
	ListIndex      int
	Scaling        bool
	MonitorID      int
	ModesEditor    bool
	MirroringMode  bool
	ColorSelection bool
}

type MonitorUnselected struct{}

type StateChanged struct {
	State AppState
}

type ViewChanged struct {
	view ViewMode
}

func StateChangedCmd(state AppState) tea.Cmd {
	return func() tea.Msg {
		return StateChanged{
			State: state,
		}
	}
}

func ViewChangedCmd(view ViewMode) tea.Cmd {
	return func() tea.Msg {
		return ViewChanged{
			view: view,
		}
	}
}

type LidStateChanged struct {
	state power.LidState
}

func LidStateChangedCmd(state power.LidState) tea.Msg {
	return LidStateChanged{
		state,
	}
}

type PowerStateChanged struct {
	state power.PowerState
}

func PowerStateChangedCmd(state power.PowerState) tea.Msg {
	return PowerStateChanged{
		state,
	}
}

type ConfigReloaded struct{}

type ProfileNameToggled struct{}

func profileNameToogled() tea.Cmd {
	return func() tea.Msg {
		return ProfileNameToggled{}
	}
}

type clearStatusMsg struct{}

func clearStatusAfter(d time.Duration) tea.Cmd {
	return tea.Tick(d, func(t time.Time) tea.Msg {
		return clearStatusMsg{}
	})
}

type OperationName int

const (
	OperationNameNone = iota
	OperationNameScale
	OperationNameRotate
	OperationNameToggleVRR
	OperationNameToggleMonitor
	OperationNameMove
	OperationNamePreviewMode
	OperationNamePreviewMirror
	OperationNameEphemeralApply
	OperationNameCreateProfile
	OperationNameMatchingProfile
	OperationNameEditProfile
	OperationNameHDMConfigReloadRequested
	OperationNameNextBitdepth
	OperationNameSetColorPreset
	OperationNameAdjustSdrBrightness
	OperationNameAdjustSdrSaturation
)

type OperationStatus struct {
	name              OperationName
	err               error
	showSuccessToUser bool
}

func (o OperationStatus) IsError() bool {
	return o.err != nil
}

func (o OperationStatus) String() string {
	operationName := ""

	switch o.name {
	case OperationNameNone:
		operationName = "None"
	case OperationNameScale:
		operationName = "Scale Apply"
	case OperationNameRotate:
		operationName = "Rotate Apply"
	case OperationNameToggleVRR:
		operationName = "Toggle VRR"
	case OperationNameToggleMonitor:
		operationName = "Toggle Monitor"
	case OperationNameMove:
		operationName = "Move"
	case OperationNamePreviewMode:
		operationName = "Preview Mode"
	case OperationNamePreviewMirror:
		operationName = "Preview Mirror"
	case OperationNameEphemeralApply:
		operationName = "Apply"
	case OperationNameCreateProfile:
		operationName = "Create Profile"
	case OperationNameMatchingProfile:
		operationName = "Matching Profile"
	case OperationNameEditProfile:
		operationName = "Edit Profile"
	case OperationNameHDMConfigReloadRequested:
		operationName = "HDM Config Reload"
	case OperationNameNextBitdepth:
		operationName = "Next Bitdepth"
	case OperationNameSetColorPreset:
		operationName = "Set Color Preset"
	case OperationNameAdjustSdrBrightness:
		operationName = "Set SDR Brightness"
	case OperationNameAdjustSdrSaturation:
		operationName = "Set SDR Saturation"
	default:
		operationName = "Operation"
	}

	result := operationName
	if o.err != nil {
		result += ": " + o.err.Error()
	} else {
		result += ": success"
	}

	if len(result) > 100 {
		result = result[:100]
	}

	return result
}

func OperationStatusCmd(name OperationName, err error) tea.Cmd {
	criticalOperations := []OperationName{
		OperationNameEditProfile,
		OperationNameCreateProfile,
		OperationNameEphemeralApply,
		OperationNameHDMConfigReloadRequested,
		OperationNameToggleMonitor,
		OperationNameToggleVRR,
		OperationNameNextBitdepth,
		OperationNameSetColorPreset,
	}
	showSuccessToUser := slices.Contains(criticalOperations, name)
	return func() tea.Msg {
		return OperationStatus{
			name:              name,
			err:               err,
			showSuccessToUser: showSuccessToUser,
		}
	}
}

type Delta int

const (
	DeltaNone Delta = iota
	DeltaMore
	DeltaLess
)

type PreviewScaleMonitorCommand struct {
	monitorID int
	scale     float64
}

type ScaleMonitorCommand struct {
	monitorID int
	scale     float64
}

type AdjustSdrBrightnessCommand struct {
	MonitorID     int
	SdrBrightness float64
}

type AdjustSdrSaturationCommand struct {
	MonitorID  int
	Saturation float64
}

type RotateMonitorCommand struct {
	MonitorID int
}

type MoveMonitorCommand struct {
	MonitorID int
	StepX     Delta
	StepY     Delta
}

type ToggleMonitorVRRCommand struct {
	MonitorID int
}

type ToggleMonitorCommand struct {
	MonitorID int
}

type ChangeColorPresetCommand struct {
	Preset ColorPreset
}

type ChangeColorPresetFinalCommand struct {
	Preset ColorPreset
}

type ChangeModePreviewCommand struct {
	Mode string
}

type ChangeMirrorPreviewCommand struct {
	MirrorOf string
}

type ChangeMirrorCommand struct {
	MirrorOf string
}

type ChangeModeCommand struct {
	Mode string
}

type NextBitdepthCommand struct {
	MonitorID int
}

type CreateNewProfileCommand struct {
	name string
	file string
}

type EditProfileConfirmationCommand struct {
	name string
}

type EditProfileCommand struct {
	name string
}

type ShowGridLineCommand struct {
	x *int
	y *int
}

type ApplyEphemeralCommand struct{}

type ToggleConfirmationPromptCommand struct{}

type CloseMonitorMirrorListCommand struct{}

type CloseColorPickerCommand struct{}

type CloseMonitorModeListCommand struct{}

type editorFinishedMsg struct{ err error }

func ShowGridLineCmd(x, y *int) tea.Cmd {
	return func() tea.Msg {
		return ShowGridLineCommand{
			x: x,
			y: y,
		}
	}
}

func openEditor(file string) tea.Cmd {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vim"
	}
	c := exec.Command(editor, file) //nolint:gosec,noctx
	return tea.ExecProcess(c, func(err error) tea.Msg {
		return editorFinishedMsg{err}
	})
}

func AdjustSdrSaturationCmd(monitorID int, value float64) tea.Cmd {
	return func() tea.Msg {
		return AdjustSdrSaturationCommand{
			MonitorID:  monitorID,
			Saturation: value,
		}
	}
}

func AdjustSdrBrightnessCmd(monitorID int, value float64) tea.Cmd {
	return func() tea.Msg {
		return AdjustSdrBrightnessCommand{
			MonitorID:     monitorID,
			SdrBrightness: value,
		}
	}
}

func ChangeColorPresetFinalCmd(preset ColorPreset) tea.Cmd {
	return func() tea.Msg {
		return ChangeColorPresetFinalCommand{
			preset,
		}
	}
}

func ChangeColorPresetCmd(preset ColorPreset) tea.Cmd {
	return func() tea.Msg {
		return ChangeColorPresetCommand{
			preset,
		}
	}
}

func CloseMonitorModeListCmd() tea.Cmd {
	return func() tea.Msg {
		return CloseMonitorModeListCommand{}
	}
}

func CloseColorPickerCmd() tea.Cmd {
	return func() tea.Msg {
		return CloseColorPickerCommand{}
	}
}

func CloseMonitorMirrorListCmd() tea.Cmd {
	return func() tea.Msg {
		return CloseMonitorMirrorListCommand{}
	}
}

func nextBitdepthCmd(monitorID int) tea.Cmd {
	return func() tea.Msg {
		return NextBitdepthCommand{
			monitorID,
		}
	}
}

func editProfileConfirmationCmd(name string) tea.Cmd {
	return func() tea.Msg {
		return EditProfileConfirmationCommand{
			name: name,
		}
	}
}

func applyEphemeralCmd() tea.Cmd {
	return func() tea.Msg {
		return ApplyEphemeralCommand{}
	}
}

func toggleConfirmationPromptCmd() tea.Cmd {
	return func() tea.Msg {
		return ToggleConfirmationPromptCommand{}
	}
}

func editProfileCmd(profile string) tea.Cmd {
	return func() tea.Msg {
		return EditProfileCommand{
			name: profile,
		}
	}
}

func createNewProfileCmd(profile, file string) tea.Cmd {
	return func() tea.Msg {
		return CreateNewProfileCommand{
			name: profile,
			file: file,
		}
	}
}

func ChangeMirrorPreviewCmd(mirror string) tea.Cmd {
	return func() tea.Msg {
		return ChangeMirrorPreviewCommand{
			MirrorOf: mirror,
		}
	}
}

func ChangeMirrorCmd(mirror string) tea.Cmd {
	return func() tea.Msg {
		return ChangeMirrorCommand{
			MirrorOf: mirror,
		}
	}
}

func ChangeModeCmd(mode string) tea.Cmd {
	return func() tea.Msg {
		return ChangeModeCommand{
			Mode: mode,
		}
	}
}

func ChangeModePreviewCmd(mode string) tea.Cmd {
	return func() tea.Msg {
		return ChangeModePreviewCommand{
			Mode: mode,
		}
	}
}

func scaleMonitorCmd(monitorID int, scale float64) tea.Cmd {
	return func() tea.Msg {
		return ScaleMonitorCommand{
			monitorID: monitorID,
			scale:     scale,
		}
	}
}

func previewScaleMonitorCmd(monitorID int, scale float64) tea.Cmd {
	return func() tea.Msg {
		return PreviewScaleMonitorCommand{
			monitorID: monitorID,
			scale:     scale,
		}
	}
}

func toggleMonitorCmd(monitor *MonitorSpec) tea.Cmd {
	return func() tea.Msg {
		return ToggleMonitorCommand{
			MonitorID: *monitor.ID,
		}
	}
}

func toggleMonitorVRRCmd(monitor *MonitorSpec) tea.Cmd {
	return func() tea.Msg {
		return ToggleMonitorVRRCommand{
			MonitorID: *monitor.ID,
		}
	}
}

// rotateMonitorCmd cycles through the rotations
func rotateMonitorCmd(monitor *MonitorSpec) tea.Cmd {
	return func() tea.Msg {
		return RotateMonitorCommand{
			MonitorID: *monitor.ID,
		}
	}
}

func MoveMonitorCmd(monitorID int, stepX, stepY Delta) tea.Cmd {
	return func() tea.Msg {
		return MoveMonitorCommand{
			MonitorID: monitorID,
			StepX:     stepX,
			StepY:     stepY,
		}
	}
}
