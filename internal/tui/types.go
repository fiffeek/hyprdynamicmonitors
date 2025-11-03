package tui

import (
	"errors"
	"fmt"
	"math"

	"github.com/fiffeek/hyprdynamicmonitors/internal/hypr"
	"github.com/fiffeek/hyprdynamicmonitors/internal/utils"
	"github.com/sirupsen/logrus"
)

type ColorPreset int

const (
	AutoColorPreset ColorPreset = iota
	SRGBColorPreset
	WideColorPreset
	EDIDColorPreset
	HDRColorPreset
	HDREDIDColorPreset
	DCIP3ColorPreset
	DP3ColorPreset
	AdobeColorPreset
)

var colorPresetMapping = map[ColorPreset]string{
	AutoColorPreset:    "auto",
	SRGBColorPreset:    "srgb",
	WideColorPreset:    "wide",
	EDIDColorPreset:    "edid",
	HDRColorPreset:     "hdr",
	HDREDIDColorPreset: "hdredid",
	DCIP3ColorPreset:   "dcip3",
	DP3ColorPreset:     "dp3",
	AdobeColorPreset:   "adobe",
}

var allColorPresets = []ColorPreset{
	AutoColorPreset,
	SRGBColorPreset,
	WideColorPreset,
	EDIDColorPreset,
	HDRColorPreset,
	HDREDIDColorPreset,
	DCIP3ColorPreset,
	DP3ColorPreset,
	AdobeColorPreset,
}

func ColorPresetFromString(color string) (ColorPreset, error) {
	for key, val := range colorPresetMapping {
		if val == color {
			return key, nil
		}
	}

	return AutoColorPreset, errors.New("cant find color")
}

func (e ColorPreset) Value() string {
	val, ok := colorPresetMapping[e]
	if ok {
		return val
	}
	return colorPresetMapping[SRGBColorPreset]
}

func (e ColorPreset) CanAdjustSdr() bool {
	return e == HDRColorPreset
}

func (e ColorPreset) IsDefault() bool {
	return e == AutoColorPreset || e == SRGBColorPreset
}

type Bitdepth int

const (
	DefaultBitdepth Bitdepth = iota
	TenBitdepth
)

func (e Bitdepth) Value() string {
	switch e {
	case DefaultBitdepth:
		return "default"
	case TenBitdepth:
		return "10"
	default:
		return "default"
	}
}

func (e Bitdepth) Bool() bool {
	return e == TenBitdepth
}

func GetBitdepth(spec *hypr.MonitorSpec) Bitdepth {
	if spec.IsTenBitdepth() {
		return TenBitdepth
	}
	return DefaultBitdepth
}

var allBitdepths = []Bitdepth{DefaultBitdepth, TenBitdepth}

type MonitorSpec struct {
	Name             string               `json:"name"`
	ID               *int                 `json:"id"`
	Description      string               `json:"description"`
	Disabled         bool                 `json:"disabled"`
	Width            int                  `json:"width"`
	Height           int                  `json:"height"`
	RefreshRate      float64              `json:"refreshRate"`
	Transform        int                  `json:"transform"`
	Vrr              bool                 `json:"vrr"`
	Scale            float64              `json:"scale"`
	X                int                  `json:"x"`
	Y                int                  `json:"y"`
	AvailableModes   []string             `json:"availableModes"`
	Mirror           string               `json:"mirrorOf"`
	CurrentFormat    string               `json:"currentFormat"`
	DpmsStatus       bool                 `json:"dpmsStatus"`
	ActivelyTearing  bool                 `json:"activelyTearing"`
	DirectScanoutTo  string               `json:"directScanoutTo"`
	Solitary         string               `json:"solitary"`
	Bitdepth         Bitdepth             `json:"-"`
	ColorPreset      ColorPreset          `json:"-"`
	SdrBrightness    float64              `json:"-"`
	SdrSaturation    float64              `json:"-"`
	Flipped          bool                 `json:"-"`
	ValidScalesCache map[float64]struct{} `json:"-"`
}

func NewMonitorSpec(spec *hypr.MonitorSpec) *MonitorSpec {
	m := &MonitorSpec{
		Name:            spec.Name,
		ID:              spec.ID,
		Description:     spec.Description,
		Disabled:        spec.Disabled,
		Width:           spec.Width,
		Height:          spec.Height,
		RefreshRate:     spec.RefreshRate,
		Transform:       spec.Transform % 4,
		Vrr:             spec.Vrr,
		Scale:           spec.Scale,
		X:               spec.X,
		Y:               spec.Y,
		AvailableModes:  spec.AvailableModes,
		Mirror:          spec.Mirror,
		CurrentFormat:   spec.CurrentFormat,
		DpmsStatus:      spec.DpmsStatus,
		ActivelyTearing: spec.ActivelyTearing,
		DirectScanoutTo: spec.DirectScanoutTo,
		Solitary:        spec.Solitary,
		Bitdepth:        GetBitdepth(spec),
		// TODO(fmikina, 12.10.25): fix after patching hyprctl to expose this information
		ColorPreset:      AutoColorPreset,
		SdrBrightness:    1.0,
		SdrSaturation:    1.0,
		Flipped:          spec.Transform >= 4,
		ValidScalesCache: make(map[float64]struct{}),
	}

	scale, err := m.ClosestValidScale(m.Scale, true, true)
	if err != nil {
		logrus.Warn("cant find any closest valid scale, leaving the current value")
	} else {
		m.Scale = scale
	}

	return m
}

func (m *MonitorSpec) Center() (int, int) {
	scaledWidth := int(float64(m.Width) / m.Scale)
	scaledHeight := int(float64(m.Height) / m.Scale)
	visualWidth := scaledWidth
	visualHeight := scaledHeight

	if m.NeedsDimensionsSwap() {
		visualWidth = scaledHeight
		visualHeight = scaledWidth
	}

	x := m.X + (visualWidth / 2)
	y := m.Y + (visualHeight / 2)

	return x, y
}

func (m *MonitorSpec) NextBitdepth() {
	current := int(m.Bitdepth)
	m.Bitdepth = Bitdepth((current + 1) % len(allBitdepths))
}

func (m *MonitorSpec) RotationPretty(showFlip bool) string {
	flipped := ", Flip: Off"
	if m.Flipped {
		flipped = ", Flip: On"
	}
	if !showFlip {
		flipped = ""
	}

	switch m.Transform {
	case 0:
		return "Rotation: 0°" + flipped
	case 1:
		return "Rotation: 90°" + flipped
	case 2:
		return "Rotation: 180°" + flipped
	case 3:
		return "Rotation: 270°" + flipped
	default:
		return fmt.Sprintf("Transform: %d", m.Transform)
	}
}

func (m *MonitorSpec) SetPreset(preset ColorPreset) {
	m.ColorPreset = preset
}

func (m *MonitorSpec) MirrorPretty() string {
	return "Mirror: " + m.Mirror
}

func (m *MonitorSpec) PositionPretty() string {
	return fmt.Sprintf("Position: %d,%d", m.X, m.Y)
}

func (m *MonitorSpec) VRRPretty() string {
	if m.Vrr {
		return "VRR: On"
	}
	return "VRR: Off"
}

func (m *MonitorSpec) LogicalSize(scale float64) (float64, float64) {
	return float64(m.Width) / scale, float64(m.Height) / scale
}

func (m *MonitorSpec) LogicalSizeFractional(scale float64) bool {
	width, height := m.LogicalSize(scale)
	return width != math.Round(width) || height != math.Round(height)
}

func (m *MonitorSpec) ClosestValidScale(scale float64, checkUp, checkDown bool) (float64, error) {
	// lots of this is based on
	// https://github.com/hyprwm/Hyprland/blob/8e9add2afda58d233a75e4c5ce8503b24fa59ceb/src/helpers/Monitor.cpp#L895
	if !m.LogicalSizeFractional(scale) {
		return scale, nil
	}

	// Invalid scale, would produce fractional pixels
	// Find the nearest valid scale by searching in increments of 1/120th
	logrus.Debugf("Scale is invalid, would produce fractional pixels, %f", scale)

	// hardcode a list of "well-known" scales to try first
	commonScales := []float64{
		0.50, 0.75, 0.90, 1.00, 1.10, 1.125, 1.25, 1.33333333, 1.3334,
		1.50, 1.6667, 1.75, 2.00, 2.125, 2.25, 2.3334, 2.50, 2.6667, 2.75, 3.00,
	}
	for _, commonScale := range commonScales {
		if m.LogicalSizeFractional(commonScale) {
			continue
		}
		if commonScale < scale && checkDown {
			m.ValidScalesCache[commonScale] = struct{}{}
		}
		if commonScale > scale && checkUp {
			m.ValidScalesCache[commonScale] = struct{}{}
		}
	}

	// try 1/initialDivider, so by increments: 0.008333333333, 0.0001, 0.001666666667
	m.searchScale(scale, checkUp, checkDown, 120.0)
	m.searchScale(scale, checkUp, checkDown, 1000.0)
	m.searchScale(scale, checkUp, checkDown, 600.0)

	// Find the closest valid scale to the requested scale
	if len(m.ValidScalesCache) == 0 {
		return 0.0, errors.New("cant find suitable scale")
	}

	var closestScale float64
	minDiff := math.MaxFloat64
	first := true

	for validScale := range m.ValidScalesCache {
		isDown := validScale < scale && checkDown
		isUp := validScale > scale && checkUp
		if !isDown && !isUp {
			continue
		}

		diff := math.Abs(scale - validScale)
		if first || diff < minDiff {
			minDiff = diff
			closestScale = validScale
			first = false
		}
	}

	if closestScale == 0 {
		return 0.0, errors.New("cant find suitable scale")
	}

	logrus.Debugf("Found closest valid scale %f for requested scale %f", closestScale, scale)
	return closestScale, nil
}

func (m *MonitorSpec) searchScale(scale float64, checkUp, checkDown bool, initialDivider float64) {
	initialValidScales := len(m.ValidScalesCache)
	searchScale := math.Round(scale * initialDivider)

	// First try the rounded scale
	scaleZero := searchScale / initialDivider
	if !m.LogicalSizeFractional(scaleZero) && scale < scaleZero && checkUp {
		m.ValidScalesCache[scaleZero] = struct{}{}
	}
	if !m.LogicalSizeFractional(scaleZero) && scale > scaleZero && checkDown {
		m.ValidScalesCache[scaleZero] = struct{}{}
	}

	// Search up and down in increments of 1/initialDivider
	for i := 1; i < 90; i++ {
		scaleUp := (searchScale + float64(i)) / initialDivider
		scaleDown := (searchScale - float64(i)) / initialDivider
		logrus.Debugf("Checking %f %f for %f", scaleUp, scaleDown, initialDivider)

		if !m.LogicalSizeFractional(scaleUp) && checkUp {
			logrus.Debugf("Scale up is valid %f for %f", scaleUp, initialDivider)
			m.ValidScalesCache[scaleUp] = struct{}{}
		}

		if !m.LogicalSizeFractional(scaleDown) && checkDown {
			logrus.Debugf("Scale down is valid %f for %f", scaleDown, initialDivider)
			m.ValidScalesCache[scaleDown] = struct{}{}
		}

		// exit on the closest found
		if initialValidScales != len(m.ValidScalesCache) {
			break
		}
	}
}

func (m *MonitorSpec) ScalePretty(withLogicalSize bool) string {
	width, height := m.LogicalSize(m.Scale)
	suffix := ""
	if m.NeedsDimensionsSwap() {
		width, height = height, width
	}
	if withLogicalSize {
		suffix = fmt.Sprintf(" (%.0fx%0.f)", width, height)
	}
	return fmt.Sprintf("Scale: %.4f", m.Scale) + suffix
}

func (m *MonitorSpec) StatusPretty() string {
	if m.Disabled {
		return "Disabled"
	}
	return "Active"
}

func (m *MonitorSpec) ModePretty() string {
	return "Mode: " +
		m.ModeForComparison()
}

func (m *MonitorSpec) Mode() string {
	return fmt.Sprintf("%dx%d@%.5f",
		m.Width,
		m.Height,
		m.RefreshRate)
}

func (m *MonitorSpec) ModeForComparison() string {
	return fmt.Sprintf("%dx%d@%.2fHz",
		m.Width,
		m.Height,
		m.RefreshRate)
}

func (m *MonitorSpec) SetMirror(mirrorOf string) {
	m.Mirror = mirrorOf
}

func (m *MonitorSpec) SetMode(mode string) error {
	var width, height int
	var refreshRate float64
	n, err := fmt.Sscanf(mode, "%dx%d@%fHz", &width, &height, &refreshRate)
	if err != nil || n != 3 {
		return fmt.Errorf("failed to parse mode: %s", mode)
	}

	m.Width = width
	m.Height = height
	m.RefreshRate = refreshRate

	// set mode changes the current scale as well, since the last set might have been invalid
	m.ValidScalesCache = make(map[float64]struct{})
	scale, err := m.ClosestValidScale(m.Scale, true, true)
	if err != nil {
		return fmt.Errorf("cant find closest valid scale: %w", err)
	}
	m.Scale = scale

	return nil
}

func (m *MonitorSpec) PositionArrowView() string {
	switch m.Transform {
	case 0: // Normal (0°) - top is up
		return "↑"
	case 1: // 90° clockwise - top is right
		return "→"
	case 2: // 180° - top is down
		return "↓"
	case 3: // 270° clockwise - top is left
		return "←"
	default:
		return "↑" // Default to up
	}
}

func (m *MonitorSpec) Rotate() {
	m.Transform = (m.Transform + 1) % 4
}

func (m *MonitorSpec) ToggleFlip() {
	m.Flipped = !m.Flipped
}

func (m *MonitorSpec) HyprTransform() int {
	if m.Flipped {
		return 4 + m.Transform
	}
	return m.Transform
}

func (m *MonitorSpec) ToggleVRR() {
	m.Vrr = !m.Vrr
}

func (m *MonitorSpec) ToggleMonitor() {
	m.Disabled = !m.Disabled
}

func (m *MonitorSpec) NeedsDimensionsSwap() bool {
	return m.Transform == 1 || m.Transform == 3
}

func (m *MonitorSpec) ToHyprMonitors() (*hypr.MonitorSpec, error) {
	monitor := &hypr.MonitorSpec{
		Name:            m.Name,
		ID:              m.ID,
		Description:     m.Description,
		Disabled:        m.Disabled,
		Width:           m.Width,
		Height:          m.Height,
		RefreshRate:     m.RefreshRate,
		Transform:       m.HyprTransform(),
		Vrr:             m.Vrr,
		Scale:           m.Scale,
		X:               m.X,
		Y:               m.Y,
		AvailableModes:  m.AvailableModes,
		Mirror:          m.Mirror,
		CurrentFormat:   m.CurrentFormat,
		DpmsStatus:      m.DpmsStatus,
		ActivelyTearing: m.ActivelyTearing,
		DirectScanoutTo: m.DirectScanoutTo,
		Solitary:        m.Solitary,
		TenBitdepth:     m.Bitdepth.Bool(),
		SdrBrightness:   m.SdrBrightness,
		SdrSaturation:   m.SdrSaturation,
		ColorPreset:     m.ColorPreset.Value(),
	}
	if err := monitor.Validate(); err != nil {
		return nil, fmt.Errorf("cant validate monitor: %w", err)
	}

	return monitor, nil
}

func (m *MonitorSpec) ToHypr() string {
	// desc can be empty, use the name as a fallback
	identifier := m.Name
	if m.Description != "" {
		identifier = "desc:" + utils.EscapeHyprDescription(m.Description)
	}
	if m.Disabled {
		// nolint:perfsprint
		return fmt.Sprintf("%s,disable", identifier)
	}
	line := fmt.Sprintf("%s,%dx%d@%.2f,%dx%d,%.8f,transform,%d", identifier, m.Width,
		m.Height, m.RefreshRate, m.X, m.Y, m.Scale, m.HyprTransform())
	if m.Vrr {
		line += ",vrr,1"
	}
	if m.Bitdepth != DefaultBitdepth {
		line = fmt.Sprintf("%s,bitdepth,%s", line, m.Bitdepth.Value())
	}
	if !m.ColorPreset.IsDefault() {
		line = fmt.Sprintf("%s,cm,%s", line, m.ColorPreset.Value())
	}
	if m.SdrBrightness != 1.0 && m.ColorPreset.CanAdjustSdr() {
		line = fmt.Sprintf("%s,sdrbrightness,%.2f", line, m.SdrBrightness)
	}
	if m.SdrSaturation != 1.0 && m.ColorPreset.CanAdjustSdr() {
		line = fmt.Sprintf("%s,sdrsaturation,%.2f", line, m.SdrSaturation)
	}
	if m.Mirror != "none" {
		line = fmt.Sprintf("%s,mirror,%s", line, m.Mirror)
	}
	return line
}

type MonitorRectangle struct {
	startX  int
	startY  int
	endX    int
	endY    int
	monitor *MonitorSpec
}

func NewMonitorRectangle(startX, startY, endX, endY int, monitor *MonitorSpec) *MonitorRectangle {
	rec := &MonitorRectangle{
		startX:  startX,
		startY:  startY,
		endX:    endX,
		endY:    endY,
		monitor: monitor,
	}

	if rec.endX <= rec.startX {
		rec.endX = rec.startX + 4
	}
	if rec.endY <= rec.startY {
		rec.endY = rec.startY + 2
	}

	return rec
}

// Clamp to grid bounds
func (m *MonitorRectangle) Clamp(gridWidth, gridHeight int) {
	if m.startX < 0 {
		m.startX = 0
	}
	if m.startY < 0 {
		m.startY = 0
	}
	if m.endX >= gridWidth {
		m.endX = gridWidth - 1
	}
	if m.endY >= gridHeight {
		m.endY = gridHeight - 1
	}
}

func (m *MonitorRectangle) isBottomEdge(x, y int) bool {
	switch m.monitor.Transform {
	case 0: // Normal (0°) - bottom is bottom
		return (y == m.endY)
	case 1: // 90° clockwise - bottom is now left
		return (x == m.startX)
	case 2: // 180° - bottom is now top
		return (y == m.startY)
	case 3: // 270° clockwise - bottom is now right
		return (x == m.endX)
	default:
		return false
	}
}

func ConvertToHyprMonitors(monitors []*MonitorSpec) ([]*hypr.MonitorSpec, error) {
	var hyprMonitors []*hypr.MonitorSpec
	for _, monitor := range monitors {
		mon, err := monitor.ToHyprMonitors()
		if err != nil {
			return nil, fmt.Errorf("cant validate monitor: %w", err)
		}
		hyprMonitors = append(hyprMonitors, mon)
	}
	return hyprMonitors, nil
}
