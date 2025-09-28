package testutils

import (
	"encoding/json"
	"os"

	"github.com/fiffeek/hyprdynamicmonitors/internal/hypr"
	"github.com/fiffeek/hyprdynamicmonitors/internal/tui"
)

// LoadMonitorsFromJSON loads monitor specifications from a JSON file
func LoadMonitorsFromJSON(filename string) ([]*tui.MonitorSpec, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var hyprSpecs []*hypr.MonitorSpec
	err = json.Unmarshal(data, &hyprSpecs)
	if err != nil {
		return nil, err
	}

	var monitors []*tui.MonitorSpec
	for _, spec := range hyprSpecs {
		monitors = append(monitors, tui.NewMonitorSpec(spec))
	}

	return monitors, nil
}

// IntPtr returns a pointer to the given int value
func IntPtr(i int) *int {
	return &i
}
