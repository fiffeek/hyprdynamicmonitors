package utils

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
)

func runCmd(ctx context.Context, name string, args ...string) (string, error) {
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
