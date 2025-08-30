package test

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
)

const (
	binaryPathEnvVar = "HDM_BINARY_PATH"
)

var (
	_, b, _, _ = runtime.Caller(0)
	basepath   = filepath.Dir(filepath.Dir(b))
	binaryPath = ""
)

func runBinary(ctx context.Context, args []string) ([]byte, error) {
	// nolint:gosec
	cmd := exec.CommandContext(
		ctx,
		filepath.Join(basepath, binaryPath),
		args...)
	cmd.Env = append(os.Environ(), "GOCOVERDIR=.coverdata")
	// nolint:wrapcheck
	return cmd.CombinedOutput()
}

func find(root, ext string) ([]string, error) {
	var files []string
	err := filepath.WalkDir(root, func(s string, d fs.DirEntry, e error) error {
		if e != nil {
			return e
		}
		if filepath.Ext(d.Name()) == ext {
			files = append(files, s)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("cant walk dir: %w", err)
	}
	return files, nil
}

func TestMain(m *testing.M) {
	binaryPath = os.Getenv(binaryPathEnvVar)
	if binaryPath == "" {
		fmt.Printf("no binary provided")
		os.Exit(1)
	}

	os.Exit(m.Run())
}
