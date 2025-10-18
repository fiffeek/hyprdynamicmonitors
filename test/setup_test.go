package test

import (
	"context"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/fiffeek/hyprdynamicmonitors/internal/testutils"
)

const (
	binaryPathEnvVar = "HDM_BINARY_PATH"
)

var (
	_, b, _, _ = runtime.Caller(0)
	basepath   = filepath.Dir(filepath.Dir(b))
	binaryPath = ""

	regenerate = flag.Bool("regenerate", false, "regenerate fixtures instead of comparing")
	debug      = flag.Bool("debug", false, "use to run the binary in debug mode")
)

func waitTillHolds(ctx context.Context, t *testing.T, funcs []func() error, timeout time.Duration) {
	deadline := time.Now().Add(timeout)
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()

	testutils.Logf(t, "waitTillHolds starting, deadline in %v", timeout)

	for {
		select {
		case <-ticker.C:
			allPass := true
			for _, f := range funcs {
				if err := f(); err != nil {
					allPass = false
					break
				}
			}
			if !allPass {
				testutils.Logf(t, "Conditions do not hold yet")
			}
			if allPass {
				testutils.Logf(t, "All conditions hold, returning")
				return
			}
			if time.Now().After(deadline) {
				testutils.Logf(t, "After deadline, exiting")
				return
			}
		case <-ctx.Done():
			testutils.Logf(t, "waitTillHolds: Context cancelled, cause: %v", context.Cause(ctx))
			return
		}
	}
}

func compareWithFixture(t *testing.T, target, fixture string) {
	if *regenerate {
		testutils.UpdateFixture(t, target, fixture)
		return
	}
	testutils.AssertContentsSameAsFixture(t, target, fixture)
}

func prepBinaryRun(t *testing.T, ctx context.Context, args []string) *exec.Cmd {
	// nolint:gosec
	cmd := exec.CommandContext(
		ctx,
		filepath.Join(basepath, binaryPath),
		args...)
	cmd.Dir = basepath
	cmd.Env = append(os.Environ(), "GOCOVERDIR=.coverdata")

	cmd.Cancel = func() error {
		return cmd.Process.Signal(os.Interrupt)
	}

	cmd.WaitDelay = 200 * time.Millisecond
	testutils.Logf(t, "Will execute: %v", cmd.Args)

	return cmd
}

func runBinary(t *testing.T, ctx context.Context, args []string) ([]byte, error) {
	// nolint:wrapcheck
	return prepBinaryRun(t, ctx, args).CombinedOutput()
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
