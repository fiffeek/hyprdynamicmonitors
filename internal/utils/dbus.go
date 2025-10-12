package utils

import (
	"fmt"
	"strings"

	"github.com/godbus/dbus/v5"
)

// SignalBodyToString recursively converts any DBus body item into a string
func SignalBodyToString(v interface{}) string {
	switch x := v.(type) {
	case string:
		return x
	case []byte:
		return string(x)
	case dbus.Variant:
		return SignalBodyToString(x.Value())
	case []interface{}:
		parts := make([]string, 0, len(x))
		for _, e := range x {
			parts = append(parts, SignalBodyToString(e))
		}
		return "[" + strings.Join(parts, " ") + "]"
	case map[string]dbus.Variant:
		parts := make([]string, 0, len(x))
		for k, val := range x {
			parts = append(parts, fmt.Sprintf("%s=%v", k, SignalBodyToString(val)))
		}
		return "{" + strings.Join(parts, " ") + "}"
	case map[string]interface{}:
		parts := make([]string, 0, len(x))
		for k, val := range x {
			parts = append(parts, fmt.Sprintf("%s=%v", k, SignalBodyToString(val)))
		}
		return "{" + strings.Join(parts, " ") + "}"
	default:
		return fmt.Sprint(x)
	}
}
