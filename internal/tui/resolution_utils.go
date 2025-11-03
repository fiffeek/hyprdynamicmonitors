package tui

import (
	"fmt"
	"math"
)

const (
	aspectRatioDelta      = 0.01
	minDesirableRefreshHz = 60.0
)

// GetAspectRatio calculates and returns the aspect ratio string
func GetAspectRatio(width, height int) string {
	gcd := func(a, b int) int {
		for b != 0 {
			a, b = b, a%b
		}
		return a
	}

	divisor := gcd(width, height)
	ratioW := width / divisor
	ratioH := height / divisor

	ratio := float64(width) / float64(height)

	switch {
	case math.Abs(ratio-16.0/9.0) < aspectRatioDelta:
		return "16:9"
	case math.Abs(ratio-16.0/10.0) < aspectRatioDelta:
		return "16:10"
	case math.Abs(ratio-21.0/9.0) < aspectRatioDelta:
		return "21:9"
	case math.Abs(ratio-32.0/9.0) < aspectRatioDelta:
		return "32:9"
	case math.Abs(ratio-4.0/3.0) < aspectRatioDelta:
		return "4:3"
	case math.Abs(ratio-5.0/4.0) < aspectRatioDelta:
		return "5:4"
	case math.Abs(ratio-3.0/2.0) < aspectRatioDelta:
		return "3:2"
	default:
		if ratioW <= 32 && ratioH <= 32 {
			return fmt.Sprintf("%d:%d", ratioW, ratioH)
		}
		return ""
	}
}

// GetResolutionClass returns the resolution class (HD, FHD, 2K, 4K, etc.)
func GetResolutionClass(width, height int) string {
	switch {
	case width >= 7680:
		return "8K"
	case width >= 5120:
		return "5K"
	case width >= 3840:
		return "4K"
	case width >= 2560:
		if height >= 1440 {
			return "QHD"
		}
		return "2K"
	case width >= 1920 && height >= 1080:
		return "FHD"
	case width >= 1366 && height >= 768:
		return "HD+"
	case width >= 1280 && height >= 800:
		return "WXGA"
	case width >= 1280 && height >= 720:
		return "HD"
	case width >= 1024 && height >= 768:
		return "XGA"
	case width >= 800 && height >= 600:
		return "SVGA"
	default:
		return ""
	}
}

// IsStandardResolution checks if the resolution matches a standard exactly
func IsStandardResolution(width, height int) bool {
	standardResolutions := map[int]map[int]bool{
		// 8K
		7680: {4320: true},
		// 5K
		5120: {2880: true},
		// 4K / UHD
		3840: {2160: true},
		// QHD / 2K
		2560: {1440: true, 1600: true},
		// FHD
		1920: {1080: true, 1200: true},
		// HD
		1280: {720: true, 800: true, 1024: true},
		1366: {768: true},
		// XGA
		1024: {768: true},
		// SVGA
		800: {600: true},
	}

	if heights, ok := standardResolutions[width]; ok {
		return heights[height]
	}
	return false
}

// IsWholeNumber checks if a float is effectively a whole number
func IsWholeNumber(f float64) bool {
	return math.Abs(f-math.Round(f)) < 0.01
}

// IsFHDOrHigher checks if the resolution is FHD (1920x1080) or higher
func IsFHDOrHigher(width, height int) bool {
	return width >= 1920 && height >= 1080
}
