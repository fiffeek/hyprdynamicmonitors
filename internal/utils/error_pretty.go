package utils

import (
	"flag"
	"fmt"
	"strings"
)

func PrettyPrintError(err error) {
	parts := strings.Split(err.Error(), ": ")
	indent := 0
	for _, part := range parts {
		fmt.Fprintf(flag.CommandLine.Output(), "%s%s\n", strings.Repeat(" ", indent), part)
		indent += 2
	}
}
