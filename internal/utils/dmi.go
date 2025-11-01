package utils

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"
)

var (
	dmiPathEnvVarOverride = "HYPRDYNAMICMONITORS_DMI_PATH_OVERRIDE"
	dmiPath               = "/sys/class/dmi/id/chassis_type"
)

// laptopLikeCodes are the SMBIOS/ DMI chassis_type numeric codes that indicate "laptop-like".
var laptopLikeCodes = map[int]struct{}{
	8:  {}, // Portable
	9:  {}, // Laptop
	10: {}, // Notebook
	14: {}, // Sub Notebook
	31: {}, // Convertible
	32: {}, // Detachable
}

func readChassisType() (int, error) {
	path, ok := os.LookupEnv(dmiPathEnvVarOverride)
	if !ok {
		path = dmiPath
	}

	//nolint:gosec
	dmiContents, err := os.ReadFile(path)
	if err != nil {
		return 0, fmt.Errorf("can't read chassis type: %w", err)
	}
	s := strings.TrimSpace(string(dmiContents))

	// Some kernels expose numeric; some distros might echo names (rare). Try parsing int first.
	if num, err := strconv.Atoi(s); err == nil {
		return num, nil
	}

	// fallback: map some textual values to numbers (if present)
	// common textual outputs: "Notebook", "Laptop", "Desktop"
	switch strings.ToLower(s) {
	case "portable":
		return 8, nil
	case "laptop", "notebook":
		return 10, nil
	case "sub-notebook", "sub notebook":
		return 14, nil
	case "desktop":
		return 3, nil
	case "server":
		return 17, nil
	default:
		return 0, nil
	}
}

func IsLaptop() bool {
	dmiCode, err := readChassisType()
	if err != nil {
		logrus.WithError(err).Warn("can't read DMI code")
		// assume laptop for backwards compatibility
		return true
	}

	_, ok := laptopLikeCodes[dmiCode]
	return ok
}
