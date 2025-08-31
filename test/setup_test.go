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
			t.Log("Conditions do not hold yet")
			if allPass {
				t.Log("All condition hold, returning")
				return
			}
			if time.Now().After(deadline) {
				t.Log("After deadline, exiting")
				return
			}
		case <-ctx.Done():
			t.Log("Ctx is done, exiting")
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

func prepBinaryRun(ctx context.Context, args []string) *exec.Cmd {
	// nolint:gosec
	cmd := exec.CommandContext(
		ctx,
		filepath.Join(basepath, binaryPath),
		args...)
	cmd.Env = append(os.Environ(), "GOCOVERDIR=.coverdata")
	return cmd
}

func runBinary(ctx context.Context, args []string) ([]byte, error) {
	// nolint:wrapcheck
	return prepBinaryRun(ctx, args).CombinedOutput()
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
