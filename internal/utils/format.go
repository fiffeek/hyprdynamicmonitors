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
