package generators

import "github.com/fiffeek/hyprdynamicmonitors/internal/hypr"

type MonitorSpec struct {
	Name            string   `json:"name"`
	ID              *int     `json:"id"`
	Description     string   `json:"description"`
	Disabled        bool     `json:"disabled"`
	AvailableModes  []string `json:"availableModes"`
	Mirror          string   `json:"mirrorOf"`
	CurrentFormat   string   `json:"currentFormat"`
	DpmsStatus      bool     `json:"dpmsStatus"`
	ActivelyTearing bool     `json:"activelyTearing"`
	DirectScanoutTo string   `json:"directScanoutTo"`
	Solitary        string   `json:"solitary"`
}

func NewMonitorSpec(monitor *hypr.MonitorSpec) *MonitorSpec {
	return &MonitorSpec{
		Name:            monitor.Name,
		Description:     monitor.Description,
		ID:              monitor.ID,
		Disabled:        monitor.Disabled,
		AvailableModes:  monitor.AvailableModes,
		Mirror:          monitor.Mirror,
		CurrentFormat:   monitor.CurrentFormat,
		DpmsStatus:      monitor.DpmsStatus,
		ActivelyTearing: monitor.ActivelyTearing,
		DirectScanoutTo: monitor.DirectScanoutTo,
		Solitary:        monitor.Solitary,
	}
}
