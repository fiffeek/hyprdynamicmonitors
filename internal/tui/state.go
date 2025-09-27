package tui

type ViewMode int

const (
	MonitorsListView ViewMode = iota
	ProfileView
)

type RootState struct {
	CurrentView ViewMode
}

func NewState() *RootState {
	return &RootState{
		CurrentView: MonitorsListView,
	}
}
