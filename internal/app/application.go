// Package app provides an application runner.
package app

import (
	"context"
	"fmt"

	"github.com/fiffeek/hyprdynamicmonitors/internal/config"
	"github.com/fiffeek/hyprdynamicmonitors/internal/filewatcher"
	"github.com/fiffeek/hyprdynamicmonitors/internal/generators"
	"github.com/fiffeek/hyprdynamicmonitors/internal/hypr"
	"github.com/fiffeek/hyprdynamicmonitors/internal/matchers"
	"github.com/fiffeek/hyprdynamicmonitors/internal/notifications"
	"github.com/fiffeek/hyprdynamicmonitors/internal/power"
	"github.com/fiffeek/hyprdynamicmonitors/internal/reloader"
	"github.com/fiffeek/hyprdynamicmonitors/internal/signal"
	"github.com/fiffeek/hyprdynamicmonitors/internal/userconfigupdater"
	"github.com/fiffeek/hyprdynamicmonitors/internal/utils"
	"github.com/godbus/dbus/v5"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

type Application struct {
	cfg           *config.Config
	hyprIPC       *hypr.IPC
	fswatcher     *filewatcher.Service
	powerDetector *power.PowerDetector
	matcher       *matchers.Matcher
	generator     *generators.ConfigGenerator
	notifications *notifications.Service
	svc           *userconfigupdater.Service
	reloader      *reloader.Service
	signal        *signal.Handler
}

func NewApplication(
	configPath *string, dryRun *bool, ctx context.Context,
	cancel context.CancelCauseFunc, disablePowerEvents, disableAutoHotReload *bool,
	connectToSessionBus *bool,
) (*Application, error) {
	cfg, err := config.NewConfig(*configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load configuration: %w", err)
	}

	hyprIPC, err := hypr.NewIPC(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Hyprland IPC: %w", err)
	}

	fswatcher := filewatcher.NewService(cfg, disableAutoHotReload)

	var conn *dbus.Conn
	if !*disablePowerEvents {
		if *connectToSessionBus {
			logrus.Debug("Trying to connect to session bus")
			conn, err = dbus.ConnectSessionBus()
		} else {
			logrus.Debug("Trying to connect to system bus")
			conn, err = dbus.ConnectSystemBus()
		}
		if err != nil {
			return nil, fmt.Errorf("cant connect to dbus: %w", err)
		}
	}
	powerDetector, err := power.NewPowerDetector(ctx, cfg, conn, *disablePowerEvents)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize PowerDetector: %w", err)
	}

	matcher := matchers.NewMatcher()

	generator := generators.NewConfigGenerator(cfg)
	notifications := notifications.NewService(cfg)

	svc := userconfigupdater.NewService(cfg, hyprIPC, powerDetector, &userconfigupdater.Config{
		DryRun: *dryRun,
	}, matcher, generator, notifications)

	reloader := reloader.NewService(cfg, fswatcher, powerDetector, svc, *disableAutoHotReload)

	signalHandler := signal.NewHandler(cancel, reloader, svc)

	return &Application{
		cfg:           cfg,
		hyprIPC:       hyprIPC,
		fswatcher:     fswatcher,
		powerDetector: powerDetector,
		matcher:       matcher,
		generator:     generator,
		notifications: notifications,
		svc:           svc,
		reloader:      reloader,
		signal:        signalHandler,
	}, nil
}

func (a *Application) RunOnce(ctx context.Context) error {
	logrus.Info("Will run one user config update")
	if err := a.svc.RunOnce(ctx); err != nil {
		return fmt.Errorf("run failed: %w", err)
	}
	logrus.Info("Run succeeded, exiting")
	return nil
}

func (a *Application) Run(ctx context.Context) error {
	eg, ctx := errgroup.WithContext(ctx)

	backgroundGoroutines := []struct {
		Fun  func(context.Context) error
		Name string
	}{
		{Fun: a.signal.Run, Name: "signal handler"},
		{Fun: a.fswatcher.Run, Name: "filewatcher"},
		{Fun: a.hyprIPC.RunEventLoop, Name: "hypr ipc"},
		{Fun: a.powerDetector.Run, Name: "power detector dbus"},
		{Fun: a.reloader.Run, Name: "reloader"},
		{Fun: a.svc.Run, Name: "main service"},
	}
	for _, bg := range backgroundGoroutines {
		eg.Go(func() error {
			fields := logrus.Fields{"name": bg.Name, "fun": utils.GetFunctionName(bg.Fun)}
			logrus.WithFields(fields).Debug("Starting")
			if err := bg.Fun(ctx); err != nil {
				logrus.WithFields(fields).WithError(err).Errorf("Service failed %s", bg.Name)
				return fmt.Errorf("%s failed: %w", bg.Name, err)
			}
			logrus.WithFields(fields).Debug("Finished")
			return nil
		})
	}

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
