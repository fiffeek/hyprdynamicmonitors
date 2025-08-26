// Package main provides the entry point for the hyprdynamicmonitors application.
// It dynamically manages Hyprland monitor configurations based on connected displays.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"path/filepath"
	"runtime"

	"github.com/fiffeek/hyprdynamicmonitors/internal/config"
	"github.com/fiffeek/hyprdynamicmonitors/internal/detectors"
	"github.com/fiffeek/hyprdynamicmonitors/internal/generators"
	"github.com/fiffeek/hyprdynamicmonitors/internal/hypr"
	"github.com/fiffeek/hyprdynamicmonitors/internal/matchers"
	"github.com/fiffeek/hyprdynamicmonitors/internal/service"
	"github.com/fiffeek/hyprdynamicmonitors/internal/signal"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

var (
	Version   = "dev"
	Commit    = "none"
	BuildDate = "unknown"
)

func main() {
	var (
		configPath = flag.String("config", "$HOME/.config/hyprdynamicmonitors/config.toml", "Path to configuration file")
		dryRun     = flag.Bool("dry-run", false, "Show what would be done without making changes")
		debug      = flag.Bool("debug", false, "Enable debug logging")
		verbose    = flag.Bool("verbose", false, "Enable verbose logging")
		showVer    = flag.Bool("version", false, "Show version information")
	)
	flag.Parse()

	if *debug {
		logrus.SetLevel(logrus.DebugLevel)
	} else {
		logrus.SetLevel(logrus.InfoLevel)
	}
	if *verbose {
		logrus.SetReportCaller(true)
	}
	logrus.SetFormatter(&logrus.TextFormatter{
		DisableTimestamp: false,
		DisableColors:    false,
		FullTimestamp:    true,
		ForceQuote:       true,
		CallerPrettyfier: func(f *runtime.Frame) (string, string) {
			fn := filepath.Base(f.Function)
			file := fmt.Sprintf("%s:%d", filepath.Base(f.File), f.Line)
			return fn, file
		},
	})

	if *showVer {
		logrus.WithFields(logrus.Fields{
			"version":   Version,
			"commit":    Commit,
			"buildDate": BuildDate,
		}).Info("hyprdynamicmonitors version")
		return
	}

	logrus.WithField("version", Version).Debug("Starting Hyprland Dynamic Monitor Manager")

	ctx, cancel := context.WithCancelCause(context.Background())
	err := createApplication(configPath, dryRun, ctx, cancel)
	if err == nil {
		logrus.Info("Exiting...")
		return
	}
	if errors.Is(err, context.Canceled) {
		logrus.WithError(err).Info("Context cancelled, exiting")
		return
	}

	// otherwise there is a real error
	logrus.WithError(err).Fatal("Service failed")
}

func createApplication(configPath *string, dryRun *bool, ctx context.Context, cancel context.CancelCauseFunc) error {
	cfg, err := config.Load(*configPath)
	if err != nil {
		logrus.WithError(err).Fatal("Failed to load configuration")
	}

	hyprIPC, err := hypr.NewIPC()
	if err != nil {
		logrus.WithError(err).Fatal("Failed to initialize Hyprland IPC")
	}

	monitorDetector, err := detectors.NewMonitorDetector(ctx, hyprIPC)
	if err != nil {
		logrus.WithError(err).Fatal("Failed to initialize MonitorDetector")
	}

	var powerDetector service.IPowerDetector
	if *cfg.PowerEvents.Disabled {
		powerDetector = detectors.NewStaticPowerDetector(cfg.PowerEvents)
	} else {
		powerDetector, err = detectors.NewPowerDetector(ctx, cfg.PowerEvents)
		if err != nil {
			logrus.WithError(err).Fatal("Failed to initialize PowerDetector")
		}
	}

	matcher := matchers.NewMatcher(cfg)

	generator := generators.NewConfigGenerator()

	svc := service.NewService(cfg, monitorDetector, powerDetector, &service.Config{
		DryRun: *dryRun,
	}, matcher, generator)

	signalHandler := signal.NewHandler(ctx, cancel)

	if err := run(ctx, svc, hyprIPC, monitorDetector, powerDetector, signalHandler); err != nil {
		return err
	}

	return nil
}

func run(ctx context.Context, svc *service.Service, hyprIPC *hypr.IPC,
	monitorDetector *detectors.MonitorDetector, powerDetector service.IPowerDetector, signalHandler *signal.Handler,
) error {
	signalHandler.Start(svc)
	defer signalHandler.Stop()

	eg, ctx := errgroup.WithContext(ctx)

	backgroundGoroutines := []struct {
		Fun  func(context.Context) error
		Name string
	}{
		{Fun: hyprIPC.RunEventLoop, Name: "hypr ipc"},
		{Fun: monitorDetector.Run, Name: "monitor detector proxy"},
		{Fun: powerDetector.Run, Name: "power detector dbus"},
		{Fun: svc.Run, Name: "main service"},
	}
	for _, bg := range backgroundGoroutines {
		eg.Go(func() error {
			fields := logrus.Fields{"name": bg.Name, "fun": bg.Fun}
			logrus.WithFields(fields).Debug("Starting")
			if err := bg.Fun(ctx); err != nil {
				return fmt.Errorf("%s failed: %w", bg.Name, err)
			}
			logrus.WithFields(fields).Debug("Finished")
			return nil
		})
	}

	eg.Go(func() error {
		<-ctx.Done()
		logrus.Debug("Context cancelled, shutting down")
		return ctx.Err()
	})

	if err := eg.Wait(); err != nil {
		return fmt.Errorf("main eg failed: %w", err)
	}

	logrus.Info("Shutdown complete")
	return nil
}
