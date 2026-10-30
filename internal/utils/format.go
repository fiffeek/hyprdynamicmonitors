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

var luaStringEscaper = strings.NewReplacer(`\`, `\\`, `"`, `\"`, "\n", `\n`)

// LuaString returns s as a double-quoted Lua string literal.
func LuaString(s string) string {
	return `"` + luaStringEscaper.Replace(s) + `"`
}

// HyprMonitorRule returns a single-line Hyprland `hl.monitor({...})` call built from `key = value` fields.
func HyprMonitorRule(fields []string) string {
	return "hl.monitor({ " + strings.Join(fields, ", ") + " })"
}
