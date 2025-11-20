package utils

import "strings"

func IsComment(line string) bool {
	line = strings.TrimLeft(line, " \t")
	return len(line) > 0 && line[0] == '#'
}
