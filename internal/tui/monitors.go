package tui

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/fiffeek/hyprdynamicmonitors/internal/hypr"
)

// LoadMonitorsFromJSON loads monitor specifications from a JSON file
func LoadMonitorsFromJSON(filename string) ([]*MonitorSpec, error) {
	// nolint:gosec
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("cant read hypr monitors: %w", err)
	}

	var hyprSpecs []*hypr.MonitorSpec
	err = json.Unmarshal(data, &hyprSpecs)
	if err != nil {
		return nil, fmt.Errorf("cant marshal: %w", err)
	}

	var monitors []*MonitorSpec
	for _, spec := range hyprSpecs {
		monitors = append(monitors, NewMonitorSpec(spec))
	}

	return monitors, nil
}
