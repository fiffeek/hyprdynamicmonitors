package tui

import (
	"encoding/json"
	"os"

	"github.com/fiffeek/hyprdynamicmonitors/internal/hypr"
)

// LoadMonitorsFromJSON loads monitor specifications from a JSON file
func LoadMonitorsFromJSON(filename string) ([]*MonitorSpec, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var hyprSpecs []*hypr.MonitorSpec
	err = json.Unmarshal(data, &hyprSpecs)
	if err != nil {
		return nil, err
	}

	var monitors []*MonitorSpec
	for _, spec := range hyprSpecs {
		monitors = append(monitors, NewMonitorSpec(spec))
	}

	return monitors, nil
}
