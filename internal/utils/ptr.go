// Package utils provides utility functions.
package utils

func StringPtr(s string) *string {
	return &s
}

func IntPtr(i int) *int {
	return &i
}

func BoolPtr(b bool) *bool {
	return &b
}

func JustPtr[T any](v T) *T {
	return &v
}
