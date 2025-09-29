package tui

import (
	"slices"

	tea "github.com/charmbracelet/bubbletea"
)

type MonitorBeingEdited struct {
	ListIndex     int
	Scaling       bool
	MonitorID     int
	ModesEditor   bool
	MirroringMode bool
}

type MonitorUnselected struct{}

type StateChanged struct {
	state AppState
}

type ViewChanged struct {
	view ViewMode
}

func viewChangedCmd(view ViewMode) tea.Cmd {
	return func() tea.Msg {
		return ViewChanged{
			view: view,
		}
	}
}

type ConfigReloaded struct{}

type ProfileNameToggled struct{}

func profileNameToogled() tea.Cmd {
	return func() tea.Msg {
		return ProfileNameToggled{}
	}
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

func operationStatusCmd(name OperationName, err error) tea.Cmd {
	criticalOperations := []OperationName{
		OperationNameEditProfile,
		OperationNameCreateProfile,
		OperationNameEphemeralApply,
		OperationNameHDMConfigReloadRequested,
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

type RotateMonitorCommand struct {
	monitorID int
}

type MoveMonitorCommand struct {
	monitorID int
	stepX     Delta
	stepY     Delta
}

type ToggleMonitorVRRCommand struct {
	monitorID int
}

type ToggleMonitorCommand struct {
	monitorID int
}

type ChangeModePreviewCommand struct {
	mode string
}

type ChangeMirrorPreviewCommand struct {
	mirrorOf string
}

type ChangeMirrorCommand struct {
	mirrorOf string
}

type ChangeModeCommand struct {
	mode string
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

type ApplyEphemeralCommand struct{}

type ToggleConfirmationPromptCommand struct{}

type CloseMonitorModeListCommand struct{}

func closeMonitorModeListCmd() tea.Cmd {
	return func() tea.Msg {
		return CloseMonitorModeListCommand{}
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

func changeMirrorPreviewCmd(mirror string) tea.Cmd {
	return func() tea.Msg {
		return ChangeMirrorPreviewCommand{
			mirrorOf: mirror,
		}
	}
}

func changeMirrorCmd(mirror string) tea.Cmd {
	return func() tea.Msg {
		return ChangeMirrorCommand{
			mirrorOf: mirror,
		}
	}
}

func changeModeCmd(mode string) tea.Cmd {
	return func() tea.Msg {
		return ChangeModeCommand{
			mode: mode,
		}
	}
}

func changeModePreviewCmd(mode string) tea.Cmd {
	return func() tea.Msg {
		return ChangeModePreviewCommand{
			mode: mode,
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
			monitorID: *monitor.ID,
		}
	}
}

func toggleMonitorVRRCmd(monitor *MonitorSpec) tea.Cmd {
	return func() tea.Msg {
		return ToggleMonitorVRRCommand{
			monitorID: *monitor.ID,
		}
	}
}

// rotateMonitorCmd cycles through the rotations
func rotateMonitorCmd(monitor *MonitorSpec) tea.Cmd {
	return func() tea.Msg {
		return RotateMonitorCommand{
			monitorID: *monitor.ID,
		}
	}
}

func moveMonitorCmd(monitorID int, stepX, stepY Delta) tea.Cmd {
	return func() tea.Msg {
		return MoveMonitorCommand{
			monitorID: monitorID,
			stepX:     stepX,
			stepY:     stepY,
		}
	}
}
