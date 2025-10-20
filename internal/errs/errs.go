// Package errs provides common errors thrown in the app that are expected to be caught upstream
package errs

import "errors"

var ErrUPowerMisconfigured = errors.New("UPower misconfigured or not running")
