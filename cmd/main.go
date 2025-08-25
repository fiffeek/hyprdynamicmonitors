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

	ctx, cancel := context.WithCancel(context.Background())

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

	if err := run(ctx, svc, hyprIPC, monitorDetector, powerDetector, signalHandler); err != nil && !errors.Is(err, context.Canceled) {
		logrus.WithError(err).Fatal("Service failed")
	}
}

func run(ctx context.Context, svc *service.Service, hyprIPC *hypr.IPC, monitorDetector *detectors.MonitorDetector, powerDetector service.IPowerDetector, signalHandler *signal.Handler) error {
	signalHandler.Start(svc)
	defer signalHandler.Stop()

	eg, ctx := errgroup.WithContext(ctx)

	eg.Go(func() error {
		logrus.Debug("Starting Hypr IPC")
		if err := hyprIPC.Run(ctx); err != nil && !errors.Is(err, context.Canceled) {
			return fmt.Errorf("hypr ipc failed: %w", err)
		}
		logrus.Debug("Hypr IPC finished")
		return nil
	})

	eg.Go(func() error {
		logrus.Debug("Starting monitor detector")
		if err := monitorDetector.Run(ctx); err != nil && !errors.Is(err, context.Canceled) {
			return fmt.Errorf("monitor detector failed: %w", err)
		}
		logrus.Debug("Monitor detector finished")
		return nil
	})

	eg.Go(func() error {
		logrus.Debug("Starting power detector")
		if err := powerDetector.Run(ctx); err != nil && !errors.Is(err, context.Canceled) {
			return fmt.Errorf("power detector failed: %w", err)
		}
		logrus.Debug("Power detector finished")
		return nil
	})

	eg.Go(func() error {
		logrus.Debug("Starting service")
		if err := svc.Run(ctx); err != nil && !errors.Is(err, context.Canceled) {
			return fmt.Errorf("service failed: %w", err)
		}
		logrus.Debug("Service finished")
		return nil
	})

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
