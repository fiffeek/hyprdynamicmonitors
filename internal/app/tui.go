package app

import (
	"context"
	"errors"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/fiffeek/hyprdynamicmonitors/internal/config"
	"github.com/fiffeek/hyprdynamicmonitors/internal/filewatcher"
	"github.com/fiffeek/hyprdynamicmonitors/internal/hypr"
	"github.com/fiffeek/hyprdynamicmonitors/internal/power"
	"github.com/fiffeek/hyprdynamicmonitors/internal/profilemaker"
	"github.com/fiffeek/hyprdynamicmonitors/internal/tui"
	"github.com/fiffeek/hyprdynamicmonitors/internal/utils"
	"github.com/godbus/dbus/v5"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

type TUI struct {
	program   *tea.Program
	fswatcher *filewatcher.Service
	cfg       *config.Config
	pw        *power.PowerDetector
}

func NewTUI(ctx context.Context, configPath, mockedHyprMonitors string,
	version string, disablePowerEvents, connectToSessionBus bool,
) (*TUI, error) {
	cfg, err := config.NewConfig(configPath)
	if err != nil {
		logrus.WithError(err).Error("cant read config, ignoring")
	}

	var monitors hypr.MonitorSpecs
	var profileMaker *profilemaker.Service
	if mockedHyprMonitors != "" {
		//nolint:gosec
		contents, err := os.ReadFile(mockedHyprMonitors)
		if err != nil {
			return nil, fmt.Errorf("cant read the mocked hypr monitors file: %w", err)
		}

		if err := utils.UnmarshalResponse(contents, &monitors); err != nil {
			return nil, fmt.Errorf("failed to parse contents: %w", err)
		}

		profileMaker = profilemaker.NewService(cfg, nil)
	} else {
		hyprIPC, err := hypr.NewIPC(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to connect to Hyprland IPC: %w", err)
		}

		monitors = hyprIPC.GetConnectedMonitors()
	}

	if err := monitors.Validate(); err != nil {
		return nil, fmt.Errorf("failed to get valid monitor information: %w", err)
	}

	var fw *filewatcher.Service
	var pw *power.PowerDetector
	var currentState power.PowerState
	if cfg != nil {
		fw = filewatcher.NewService(cfg, utils.BoolPtr(false))
	}
	if cfg != nil && !disablePowerEvents {
		var conn *dbus.Conn
		if connectToSessionBus {
			logrus.Debug("Trying to connect to session bus")
			conn, err = dbus.ConnectSessionBus()
		} else {
			logrus.Debug("Trying to connect to system bus")
			conn, err = dbus.ConnectSystemBus()
		}
		if err != nil {
			return nil, fmt.Errorf("cant connect to dbus: %w", err)
		}
		pw, err = power.NewPowerDetector(ctx, cfg, conn, disablePowerEvents)
		if err != nil {
			return nil, fmt.Errorf("cant init power detector: %w", err)
		}
		currentState = pw.GetCurrentState()
	}

	model := tui.NewModel(cfg, monitors, profileMaker, version, currentState, nil)
	program := tea.NewProgram(
		model,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	return &TUI{
		program:   program,
		fswatcher: fw,
		cfg:       cfg,
		pw:        pw,
	}, nil
}

func (t *TUI) Run(ctx context.Context, cancel context.CancelCauseFunc) error {
	eg, ctx := errgroup.WithContext(ctx)

	// listen to fs events if the config is provided and valid
	if t.fswatcher != nil && t.cfg != nil {
		eg.Go(func() error {
			return t.fswatcher.Run(ctx)
		})

		eg.Go(func() error {
			c := t.fswatcher.Listen()
			for {
				select {
				case _, ok := <-c:
					if !ok {
						return errors.New("watcher event channel closed")
					}
					logrus.Debug("Watcher event received")
					if err := t.cfg.Reload(); err != nil {
						return fmt.Errorf("cant reload user configuration: %w", err)
					}
					t.program.Send(tui.ConfigReloaded{})

				case <-ctx.Done():
					logrus.Debug("Reloader event processor context cancelled, shutting down")
					return context.Cause(ctx)
				}
			}
		})
	}

	if t.pw != nil && t.cfg != nil {
		eg.Go(func() error {
			return t.pw.Run(ctx)
		})

		eg.Go(func() error {
			c := t.pw.Listen()
			for {
				select {
				case event, ok := <-c:
					if !ok {
						return errors.New("watcher event channel closed")
					}
					logrus.Debug("Watcher event received")
					t.program.Send(tui.PowerStateChangedCmd(event.State))

				case <-ctx.Done():
					logrus.Debug("Reloader event processor context cancelled, shutting down")
					return context.Cause(ctx)
				}
			}
		})

	}

	eg.Go(func() error {
		if _, err := t.program.Run(); err != nil {
			return fmt.Errorf("failed to run TUI: %w", err)
		}
		cancel(context.Canceled)
		logrus.Debug("Exiting tea")
		return nil
	})

	eg.Go(func() error {
		<-ctx.Done()
		logrus.Debug("Context cancelled, shutting down")
		return context.Cause(ctx)
	})

	if err := eg.Wait(); err != nil {
		return fmt.Errorf("main eg failed: %w", err)
	}

	logrus.Info("Shutdown complete")
	return nil
}
