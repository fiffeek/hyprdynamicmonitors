package userconfigupdater

import (
	"testing"

	"github.com/fiffeek/hyprdynamicmonitors/internal/hypr"
	"github.com/fiffeek/hyprdynamicmonitors/internal/power"
	"github.com/stretchr/testify/assert"
)

func monitor(name, description string) *hypr.MonitorSpec {
	return &hypr.MonitorSpec{Name: name, Description: description}
}

func TestAppliedStateKey_SameSetDifferentOrderYieldsSameKey(t *testing.T) {
	laptop := monitor("eDP-1", "Some Laptop Panel")
	external := monitor("DP-2", "Some External Display")

	first := appliedStateKey("docked", hypr.MonitorSpecs{laptop, external},
		power.ACPowerState, power.OpenedLidState)
	second := appliedStateKey("docked", hypr.MonitorSpecs{external, laptop},
		power.ACPowerState, power.OpenedLidState)

	assert.Equal(t, first, second, "monitor event ordering must not affect the state key")
}

func TestAppliedStateKey_IgnoresVolatileFields(t *testing.T) {
	before := &hypr.MonitorSpec{Name: "eDP-1", Description: "Some Laptop Panel", DpmsStatus: true}
	// same output, but blanked by idle and now the solitary/tearing candidate
	after := &hypr.MonitorSpec{
		Name: "eDP-1", Description: "Some Laptop Panel",
		DpmsStatus: false, ActivelyTearing: true, Solitary: "1", DirectScanoutTo: "1",
	}

	assert.Equal(t,
		appliedStateKey("laptop", hypr.MonitorSpecs{before}, power.BatteryPowerState, power.OpenedLidState),
		appliedStateKey("laptop", hypr.MonitorSpecs{after}, power.BatteryPowerState, power.OpenedLidState),
		"screen blanking and content-driven fields must not trigger a re-apply")
}

func TestAppliedStateKey_DetectsOutputRename(t *testing.T) {
	// The same physical display can come back on a different connector, e.g. after
	// re-attaching a dock. A profile matching on description keeps selecting the same
	// profile, so the rendered config is byte-identical -- yet the new output still
	// needs the configuration applied to it.
	onDP3 := monitor("DP-3", "Some External Display")
	onDP2 := monitor("DP-2", "Some External Display")
	laptop := monitor("eDP-1", "Some Laptop Panel")

	assert.NotEqual(t,
		appliedStateKey("docked", hypr.MonitorSpecs{laptop, onDP3}, power.ACPowerState, power.OpenedLidState),
		appliedStateKey("docked", hypr.MonitorSpecs{laptop, onDP2}, power.ACPowerState, power.OpenedLidState),
		"an output rename must be detected even though the profile is unchanged")
}

func TestAppliedStateKey_DetectsMonitorAppearingAndDisappearing(t *testing.T) {
	laptop := monitor("eDP-1", "Some Laptop Panel")
	external := monitor("DP-2", "Some External Display")

	alone := appliedStateKey("laptop", hypr.MonitorSpecs{laptop},
		power.ACPowerState, power.OpenedLidState)
	withExternal := appliedStateKey("laptop", hypr.MonitorSpecs{laptop, external},
		power.ACPowerState, power.OpenedLidState)

	assert.NotEqual(t, alone, withExternal, "a monitor appearing must change the state key")
}

func TestAppliedStateKey_DetectsDisabledFlip(t *testing.T) {
	enabled := &hypr.MonitorSpec{Name: "DP-2", Description: "Some External Display"}
	disabled := &hypr.MonitorSpec{Name: "DP-2", Description: "Some External Display", Disabled: true}

	assert.NotEqual(t,
		appliedStateKey("docked", hypr.MonitorSpecs{enabled}, power.ACPowerState, power.OpenedLidState),
		appliedStateKey("docked", hypr.MonitorSpecs{disabled}, power.ACPowerState, power.OpenedLidState),
		"an output being disabled must change the state key")
}

func TestAppliedStateKey_DetectsPowerAndLidAndProfile(t *testing.T) {
	monitors := hypr.MonitorSpecs{monitor("eDP-1", "Some Laptop Panel")}
	base := appliedStateKey("laptop", monitors, power.ACPowerState, power.OpenedLidState)

	assert.NotEqual(t, base,
		appliedStateKey("laptop", monitors, power.BatteryPowerState, power.OpenedLidState),
		"power state must be part of the key")
	assert.NotEqual(t, base,
		appliedStateKey("laptop", monitors, power.ACPowerState, power.ClosedLidState),
		"lid state must be part of the key")
	assert.NotEqual(t, base,
		appliedStateKey("docked", monitors, power.ACPowerState, power.OpenedLidState),
		"profile name must be part of the key")
}

func TestAppliedStateKey_ToleratesNilEntries(t *testing.T) {
	withNil := hypr.MonitorSpecs{monitor("eDP-1", "Some Laptop Panel"), nil}
	withoutNil := hypr.MonitorSpecs{monitor("eDP-1", "Some Laptop Panel")}

	assert.Equal(t,
		appliedStateKey("laptop", withoutNil, power.ACPowerState, power.OpenedLidState),
		appliedStateKey("laptop", withNil, power.ACPowerState, power.OpenedLidState),
		"nil entries must not corrupt the key")
}

func TestAppliedStateDiffers_Lifecycle(t *testing.T) {
	s := &Service{}

	assert.True(t, s.appliedStateDiffers("first"),
		"nothing applied yet, so any state must count as different")

	s.recordApplied("first")
	assert.False(t, s.appliedStateDiffers("first"),
		"the state just applied must not be re-applied")
	assert.True(t, s.appliedStateDiffers("second"),
		"a different state must be detected")

	s.recordApplied("second")
	assert.False(t, s.appliedStateDiffers("second"))
	assert.True(t, s.appliedStateDiffers("first"),
		"returning to an earlier state must still require applying it")
}
