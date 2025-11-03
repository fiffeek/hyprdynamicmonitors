package tui_test

import (
	"testing"

	"github.com/fiffeek/hyprdynamicmonitors/internal/tui"
	"github.com/stretchr/testify/assert"
)

func TestGetAspectRatio(t *testing.T) {
	testCases := []struct {
		name     string
		width    int
		height   int
		expected string
	}{
		// Standard aspect ratios
		{"16:9 - 1920x1080", 1920, 1080, "16:9"},
		{"16:9 - 3840x2160", 3840, 2160, "16:9"},
		{"16:9 - 2560x1440", 2560, 1440, "16:9"},
		{"16:10 - 1920x1200", 1920, 1200, "16:10"},
		{"16:10 - 2560x1600", 2560, 1600, "16:10"},
		{"21:9 - 2560x1080", 2560, 1080, ""},
		{"21:9 - 3440x1440", 3440, 1440, ""},
		{"32:9 - 5120x1440", 5120, 1440, "32:9"},
		{"4:3 - 1024x768", 1024, 768, "4:3"},
		{"4:3 - 1600x1200", 1600, 1200, "4:3"},
		{"5:4 - 1280x1024", 1280, 1024, "5:4"},
		{"3:2 - 2880x1920", 2880, 1920, "3:2"},
		{"3:2 - 3000x2000", 3000, 2000, "3:2"},

		// Non-standard but simplified ratios
		{"8:5 (16:10) - 1280x800", 1280, 800, "16:10"},
		{"17:9 - 2048x1080", 2048, 1080, ""},

		// Very non-standard (matches 16:9 within delta)
		{"Complex ratio", 1366, 768, "16:9"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := tui.GetAspectRatio(tc.width, tc.height)
			assert.Equal(t, tc.expected, result, "Aspect ratio mismatch for %dx%d", tc.width, tc.height)
		})
	}
}

func TestGetResolutionClass(t *testing.T) {
	testCases := []struct {
		name     string
		width    int
		height   int
		expected string
	}{
		// 8K
		{"8K - 7680x4320", 7680, 4320, "8K"},
		{"8K - 8000x4000", 8000, 4000, "8K"},

		// 5K
		{"5K - 5120x2880", 5120, 2880, "5K"},
		{"5K - 5200x2900", 5200, 2900, "5K"},

		// 4K
		{"4K - 3840x2160", 3840, 2160, "4K"},
		{"4K - 4096x2160", 4096, 2160, "4K"},

		// QHD
		{"QHD - 2560x1440", 2560, 1440, "QHD"},
		{"QHD - 2560x1600", 2560, 1600, "QHD"},

		// 2K (width >= 2560 but height < 1440)
		{"2K - 2560x1080", 2560, 1080, "2K"},

		// FHD
		{"FHD - 1920x1080", 1920, 1080, "FHD"},
		{"FHD - 1920x1200", 1920, 1200, "FHD"},
		{"FHD - 2048x1080", 2048, 1080, "FHD"},

		// HD
		{"HD - 1280x720", 1280, 720, "HD"},
		{"HD - 1366x768", 1366, 768, "HD+"},

		// WXGA
		{"WXGA - 1280x800", 1280, 800, "WXGA"},

		// XGA
		{"XGA - 1024x768", 1024, 768, "XGA"},

		// SVGA
		{"SVGA - 800x600", 800, 600, "SVGA"},

		// Below SVGA
		{"Below SVGA - 640x480", 640, 480, ""},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := tui.GetResolutionClass(tc.width, tc.height)
			assert.Equal(t, tc.expected, result, "Resolution class mismatch for %dx%d", tc.width, tc.height)
		})
	}
}

func TestIsStandardResolution(t *testing.T) {
	testCases := []struct {
		name     string
		width    int
		height   int
		expected bool
	}{
		// Standard resolutions (should return true)
		{"8K - 7680x4320", 7680, 4320, true},
		{"5K - 5120x2880", 5120, 2880, true},
		{"4K - 3840x2160", 3840, 2160, true},
		{"QHD - 2560x1440", 2560, 1440, true},
		{"2K - 2560x1600", 2560, 1600, true},
		{"FHD - 1920x1080", 1920, 1080, true},
		{"FHD - 1920x1200", 1920, 1200, true},
		{"HD - 1280x720", 1280, 720, true},
		{"HD - 1280x800", 1280, 800, true},
		{"HD - 1280x1024", 1280, 1024, true},
		{"HD+ - 1366x768", 1366, 768, true},
		{"XGA - 1024x768", 1024, 768, true},
		{"SVGA - 800x600", 800, 600, true},

		// Non-standard resolutions (should return false)
		{"Non-standard - 1680x1050", 1680, 1050, false},
		{"Non-standard - 2560x1080", 2560, 1080, false},
		{"Non-standard - 3440x1440", 3440, 1440, false},
		{"Non-standard - 2880x1920", 2880, 1920, false},
		{"Non-standard - 1600x900", 1600, 900, false},
		{"Non-standard - 640x480", 640, 480, false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := tui.IsStandardResolution(tc.width, tc.height)
			assert.Equal(t, tc.expected, result, "Standard resolution check mismatch for %dx%d", tc.width, tc.height)
		})
	}
}

func TestIsWholeNumber(t *testing.T) {
	testCases := []struct {
		name     string
		value    float64
		expected bool
	}{
		{"Whole number - 1.0", 1.0, true},
		{"Whole number - 60.0", 60.0, true},
		{"Whole number - 120.0", 120.0, true},
		{"Whole number - 144.0", 144.0, true},
		{"Almost whole - 60.001", 60.001, true},
		{"Almost whole - 59.999", 59.999, true},
		{"Almost whole - 120.005", 120.005, true},

		{"Not whole - 59.5", 59.5, false},
		{"Not whole - 60.94", 60.94, false},
		{"Not whole - 119.02", 119.02, false},
		{"Not whole - 30.5", 30.5, false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := tui.IsWholeNumber(tc.value)
			assert.Equal(t, tc.expected, result, "Whole number check mismatch for %.3f", tc.value)
		})
	}
}

func TestIsFHDOrHigher(t *testing.T) {
	testCases := []struct {
		name     string
		width    int
		height   int
		expected bool
	}{
		// FHD and higher (should return true)
		{"FHD - 1920x1080", 1920, 1080, true},
		{"FHD+ - 1920x1200", 1920, 1200, true},
		{"QHD - 2560x1440", 2560, 1440, true},
		{"4K - 3840x2160", 3840, 2160, true},
		{"5K - 5120x2880", 5120, 2880, true},
		{"8K - 7680x4320", 7680, 4320, true},
		{"Ultrawide FHD - 2560x1080", 2560, 1080, true},
		{"21:9 FHD - 3440x1440", 3440, 1440, true},

		// Below FHD (should return false)
		{"HD - 1280x720", 1280, 720, false},
		{"HD+ - 1366x768", 1366, 768, false},
		{"WXGA - 1280x800", 1280, 800, false},
		{"XGA - 1024x768", 1024, 768, false},
		{"SVGA - 800x600", 800, 600, false},
		{"Just below FHD width - 1919x1080", 1919, 1080, false},
		{"Just below FHD height - 1920x1079", 1920, 1079, false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := tui.IsFHDOrHigher(tc.width, tc.height)
			assert.Equal(t, tc.expected, result, "FHD or higher check mismatch for %dx%d", tc.width, tc.height)
		})
	}
}
