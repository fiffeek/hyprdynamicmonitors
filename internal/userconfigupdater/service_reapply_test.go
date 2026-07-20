package userconfigupdater_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/fiffeek/hyprdynamicmonitors/internal/config"
	"github.com/fiffeek/hyprdynamicmonitors/internal/generators"
	"github.com/fiffeek/hyprdynamicmonitors/internal/hypr"
	"github.com/fiffeek/hyprdynamicmonitors/internal/matchers"
	"github.com/fiffeek/hyprdynamicmonitors/internal/notifications"
	"github.com/fiffeek/hyprdynamicmonitors/internal/power"
	"github.com/fiffeek/hyprdynamicmonitors/internal/testutils"
	"github.com/fiffeek/hyprdynamicmonitors/internal/userconfigupdater"
	"github.com/fiffeek/hyprdynamicmonitors/internal/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type stubMonitorDetector struct{ monitors hypr.MonitorSpecs }

func (s *stubMonitorDetector) Listen() <-chan hypr.MonitorSpecs { return nil }
func (s *stubMonitorDetector) GetConnectedMonitors() hypr.MonitorSpecs {
	return s.monitors
}

type stubPowerDetector struct{}

func (s *stubPowerDetector) GetCurrentState() power.PowerState  { return power.ACPowerState }
func (s *stubPowerDetector) Listen() <-chan power.PowerEvent    { return nil }

type stubLidDetector struct{}

func (s *stubLidDetector) GetCurrentState() power.LidState { return power.OpenedLidState }
func (s *stubLidDetector) Listen() <-chan power.LidEvent   { return nil }

// TestPostApplyRunsWhenMonitorSetChangesUnderSameProfile reproduces the field failure:
// a dock flap brings the external display back on a different connector. The profile
// matches on description, so it stays selected and the rendered config is identical --
// yet the new output still needs the configuration applied to it.
func TestPostApplyRunsWhenMonitorSetChangesUnderSameProfile(t *testing.T) {
	tempDir := t.TempDir()
	marker := filepath.Join(tempDir, "post_apply_runs")
	destination := filepath.Join(tempDir, "monitors-dynamic.conf")

	cfg := testutils.NewTestConfig(t).
		WithProfiles(map[string]*config.Profile{
			"docked": {
				ConfigType: utils.JustPtr(config.Static),
				Conditions: &config.ProfileCondition{
					RequiredMonitors: []*config.RequiredMonitor{
						{Name: utils.StringPtr("eDP-1")},
						{Description: utils.StringPtr("External Monitor")},
					},
				},
			},
		}).
		WithDestination(destination).
		WithPostExec("echo applied >> " + marker).
		Get()

	generator, err := generators.NewConfigGenerator(cfg)
	require.NoError(t, err, "should be able to build the generator")

	detector := &stubMonitorDetector{}
	svc := userconfigupdater.NewService(cfg, detector, &stubPowerDetector{},
		&userconfigupdater.Config{DryRun: false}, matchers.NewMatcher(), generator,
		notifications.NewService(cfg), &stubLidDetector{})

	applyCount := func() int {
		contents, err := os.ReadFile(marker)
		if os.IsNotExist(err) {
			return 0
		}
		require.NoError(t, err, "should be able to read the marker file")
		return len(strings.Fields(string(contents)))
	}

	laptop := &hypr.MonitorSpec{ID: utils.JustPtr(0), Name: "eDP-1", Description: "Laptop Panel"}
	onDP3 := &hypr.MonitorSpec{ID: utils.JustPtr(1), Name: "DP-3", Description: "External Monitor"}
	onDP2 := &hypr.MonitorSpec{ID: utils.JustPtr(2), Name: "DP-2", Description: "External Monitor"}

	ctx := context.Background()

	// first apply: the config is linked for the first time
	detector.monitors = hypr.MonitorSpecs{laptop, onDP3}
	require.NoError(t, svc.RunOnce(ctx))
	require.Equal(t, 1, applyCount(), "the first apply must run the hook")

	// nothing changed at all: the hook must not run again
	require.NoError(t, svc.RunOnce(ctx))
	assert.Equal(t, 1, applyCount(),
		"an unchanged state must not re-run the hook")

	// the display comes back on another connector; same profile, identical config file
	detector.monitors = hypr.MonitorSpecs{laptop, onDP2}
	require.NoError(t, svc.RunOnce(ctx))
	assert.Equal(t, 2, applyCount(),
		"a monitor set change under an unchanged profile must re-run the hook")

	// settled again on the new topology: still no repeated work
	require.NoError(t, svc.RunOnce(ctx))
	assert.Equal(t, 2, applyCount(),
		"the hook must not run once the new state has been applied")
}
