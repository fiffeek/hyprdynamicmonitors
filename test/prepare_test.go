package test

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/fiffeek/hyprdynamicmonitors/internal/config"
	"github.com/fiffeek/hyprdynamicmonitors/internal/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test__Run_Prepare_Binary(t *testing.T) {
	tests := []struct {
		name                  string
		cfg                   *testutils.TestConfig
		expectError           bool
		expectOutputToContain string
		compareToGolden       string
		prerun                func(*config.RawConfig)
		cleanup               func(*config.RawConfig)
	}{
		{
			name:        "run on nonexiting file",
			cfg:         testutils.NewTestConfig(t),
			expectError: false,
		},
		{
			name: "basic",
			cfg: testutils.NewTestConfig(t).FillDefaults().FillDefaults().WithDestinationContents(
				"testdata/Test__Run_Prepare_Binary/fixtures/basic.conf"),
			expectError:     false,
			compareToGolden: "testdata/Test__Run_Prepare_Binary/golden/basic.golden",
		},
		{
			name: "filtered",
			cfg: testutils.NewTestConfig(t).FillDefaults().FillDefaults().WithDestinationContents(
				"testdata/Test__Run_Prepare_Binary/fixtures/filtered.conf"),
			expectError:     false,
			compareToGolden: "testdata/Test__Run_Prepare_Binary/golden/filtered.golden",
		},
		{
			name: "complex",
			cfg: testutils.NewTestConfig(t).FillDefaults().FillDefaults().WithDestinationContents(
				"testdata/Test__Run_Prepare_Binary/fixtures/complex.conf"),
			expectError:     false,
			compareToGolden: "testdata/Test__Run_Prepare_Binary/golden/complex.golden",
		},
		{
			name: "destination unreadable",
			cfg: testutils.NewTestConfig(t).FillDefaults().FillDefaults().WithDestinationContents(
				"testdata/Test__Run_Prepare_Binary/fixtures/basic.conf"),
			prerun: func(cfg *config.RawConfig) {
				//nolint:gosec
				require.NoError(t, os.Chmod(*cfg.General.Destination, 0o300))
			},
			expectError:           true,
			expectOutputToContain: "cant read",
		},
		{
			name: "destination unwriteable",
			cfg: testutils.NewTestConfig(t).FillDefaults().FillDefaults().WithDestinationContents(
				"testdata/Test__Run_Prepare_Binary/fixtures/basic.conf"),
			prerun: func(cfg *config.RawConfig) {
				//nolint:gosec
				require.NoError(t, os.Chmod(filepath.Dir(*cfg.General.Destination), 0o500))
			},
			expectError:           true,
			expectOutputToContain: "cant write",
			cleanup: func(cfg *config.RawConfig) {
				//nolint:gosec
				require.NoError(t, os.Chmod(filepath.Dir(*cfg.General.Destination), 0o700))
			},
		},
		{
			name: "empty file",
			cfg: testutils.NewTestConfig(t).FillDefaults().FillDefaults().WithDestinationContents(
				"testdata/Test__Run_Prepare_Binary/fixtures/empty.conf"),
			expectError:     false,
			compareToGolden: "testdata/Test__Run_Prepare_Binary/golden/empty.golden",
		},
		{
			name: "file with only disable lines",
			cfg: testutils.NewTestConfig(t).FillDefaults().FillDefaults().WithDestinationContents(
				"testdata/Test__Run_Prepare_Binary/fixtures/only-disable.conf"),
			expectError:     false,
			compareToGolden: "testdata/Test__Run_Prepare_Binary/golden/only-disable.golden",
		},
		{
			name: "file without trailing newline",
			cfg: testutils.NewTestConfig(t).FillDefaults().FillDefaults().WithDestinationContents(
				"testdata/Test__Run_Prepare_Binary/fixtures/no-trailing-newline.conf"),
			expectError:     false,
			compareToGolden: "testdata/Test__Run_Prepare_Binary/golden/no-trailing-newline.golden",
		},
		{
			name: "comments matching regex",
			cfg: testutils.NewTestConfig(t).FillDefaults().FillDefaults().WithDestinationContents(
				"testdata/Test__Run_Prepare_Binary/fixtures/comments-matching.conf"),
			expectError:     false,
			compareToGolden: "testdata/Test__Run_Prepare_Binary/golden/comments-matching.golden",
		},
		{
			name: "file with only whitespace",
			cfg: testutils.NewTestConfig(t).FillDefaults().FillDefaults().WithDestinationContents(
				"testdata/Test__Run_Prepare_Binary/fixtures/whitespace-only.conf"),
			expectError:     false,
			compareToGolden: "testdata/Test__Run_Prepare_Binary/golden/whitespace-only.golden",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 2000*time.Millisecond)
			defer cancel()
			cfg := tt.cfg.Get().Get()

			if tt.cleanup != nil {
				defer tt.cleanup(cfg)
			}

			cmd := prepBinaryRun(t, ctx, []string{"prepare", "--config", cfg.ConfigPath})

			if tt.prerun != nil {
				tt.prerun(cfg)
			}

			out, binaryErr := cmd.CombinedOutput()

			if tt.expectError {
				assert.Error(t, binaryErr, "binary was expected to exit with an error")
			} else {
				assert.NoError(t, binaryErr, "binary was expected to return normally but didnt")
			}

			if tt.expectOutputToContain != "" {
				assert.Contains(t, string(out), tt.expectOutputToContain, "output should contain the message")
			}

			if tt.compareToGolden != "" {
				testutils.AssertFixture(t, *cfg.General.Destination, tt.compareToGolden, *regenerate)
			}
		})
	}
}
