package utils

import "strings"

type HasValue interface {
	Value() string
}

func FormatEnumTypes[T HasValue](enums []T) string {
	mapped := []string{}
	for _, configFileType := range enums {
		mapped = append(mapped, configFileType.Value())
	}
	return "[" + strings.Join(mapped, ", ") + "]"
}

// EscapeHyprDescription escapes # characters in monitor descriptions by doubling them.
// In Hyprland's configuration syntax, # needs to be escaped as ## when used in desc: fields.
func EscapeHyprDescription(desc string) string {
	return strings.ReplaceAll(desc, "#", "##")
}
