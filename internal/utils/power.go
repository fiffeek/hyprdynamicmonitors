package utils

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

var preferredSuffixes = []string{
	"line_power_AC",
	"line_power_ACAD",
	"line_power_ADP1",
	"line_power_Mains",
	"line_power_DCIN",
	"line_power_USB",
	"line_power_usb",
	"line_power_ACAD0",
	"line_power_ACAD1",
	"line_power_ADP0",
	"line_power_ADP2",
	"line_power_C0",
	"line_power_acadapter",
	"line_power_ac_adapter",
	"line_power_WALL",
	"line_power_mains",
	"line_power_charger",
	"line_power_USB_C",
	"line_power_USBC",
	"line_power_PD",
	"line_power_wireless",
}

func matchesPreferred(path string) bool {
	for _, suf := range preferredSuffixes {
		if strings.HasSuffix(path, suf) {
			return true
		}
	}
	return false
}

func GetPowerLine() (*string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	out, err := runCmd(ctx, "upower", "-e")
	if err != nil {
		return nil, fmt.Errorf("cant get the current power line: %w", err)
	}

	var paths []string
	scanner := bufio.NewScanner(strings.NewReader(out))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			paths = append(paths, line)
		}
	}

	for _, p := range paths {
		if matchesPreferred(p) {
			logrus.Debug("Found matching power line " + p)
			return &p, nil
		}
	}

	return nil, errors.New("cant find the current power line")
}
