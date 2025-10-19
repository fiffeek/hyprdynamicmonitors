package daemon_test

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

var (
	_, b, _, _ = runtime.Caller(0)
	basepath   = filepath.Dir(filepath.Dir(filepath.Dir(b)))
)

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}
