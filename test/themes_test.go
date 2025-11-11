package test

import (
	"bytes"
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/fiffeek/hyprdynamicmonitors/internal/config"
	"github.com/fiffeek/hyprdynamicmonitors/internal/testutils"
	"github.com/fiffeek/hyprdynamicmonitors/internal/utils"
	"github.com/stretchr/testify/require"
)

var staticThemes = filepath.Join(basepath, "themes/static/")

func Test__Run_With_Themes(t *testing.T) {
	files, err := find(staticThemes, ".toml")
	require.NoError(t, err, "didnt find all example configs")
	for _, file := range files {
		theme := filepath.Base(filepath.Dir(file))
		t.Run(theme, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
			defer cancel()

			cfg := testutils.NewTestConfig(t).WithThemeFile(file)
			cfgPath := cfg.Get().Get().ConfigPath

			done := make(chan any, 1)
			defer close(done)

			go func() {
				out, err := runBinary(t, ctx, []string{"--config", cfgPath, "validate"})
				require.NoError(t, err, "binary failed %s", string(out))
				done <- true
			}()

			select {
			case <-ctx.Done():
				require.NoError(t, ctx.Err(), "timeout")
			case <-done:
				cfg, err := config.Load(cfg.Get().Get().ConfigPath)
				require.NoError(t, err, "config should pass validation")

				buf := new(bytes.Buffer)
				encoder := toml.NewEncoder(buf)
				encoder.Indent = ""
				require.NoError(t, encoder.Encode(cfg.TUISection.Colors))
				require.NoError(t, utils.WriteAtomic(cfgPath, buf.Bytes()))
				testutils.AssertFixture(t, cfgPath, "testdata/Test__Run_With_Themes/"+theme, *regenerate)
			}
		})
	}
}
