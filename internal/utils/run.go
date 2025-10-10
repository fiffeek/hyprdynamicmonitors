package utils

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
)

// cmdRunner is the function type for running commands
type cmdRunner func(ctx context.Context, name string, args ...string) (string, error)

// runCmd is a variable that can be swapped in tests
var runCmd cmdRunner = execCommand

func execCommand(ctx context.Context, name string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("command `%s %s` failed: %w (%s)",
			name, strings.Join(args, " "), err, stderr.String())
	}
	return strings.TrimSpace(out.String()), nil
}

// GetRunCmd returns the current command runner (for testing)
func GetRunCmd() cmdRunner {
	return runCmd
}

// SetRunCmd sets the command runner (for testing)
func SetRunCmd(r cmdRunner) {
	runCmd = r
}
