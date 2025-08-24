package main

import (
	"context"
	"flag"

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
	version   = "dev"
	commit    = "none"
	buildDate = "unknown"
)

func main() {
	var (
		configPath = flag.String("config", "$HOME/.config/hyprdynamicmonitors/config.toml", "Path to configuration file")
		dryRun     = flag.Bool("dry-run", false, "Show what would be done without making changes")
		verbose    = flag.Bool("verbose", false, "Enable verbose logging")
		showVer    = flag.Bool("version", false, "Show version information")
	)
	flag.Parse()

	if *verbose {
		logrus.SetLevel(logrus.DebugLevel)
	} else {
		logrus.SetLevel(logrus.InfoLevel)
	}
	logrus.SetFormatter(&logrus.TextFormatter{
		DisableTimestamp: false,
		DisableColors:    false,
		FullTimestamp:    true,
	})

	if *showVer {
		logrus.WithFields(logrus.Fields{
			"version":   version,
			"commit":    commit,
			"buildDate": buildDate,
		}).Info("hyprdynamicmonitors version")
		return
	}

	logrus.WithField("version", version).Debug("Starting Hyprland Dynamic Monitor Manager")

	ctx, cancel := context.WithCancel(context.Background())

	cfg, err := config.Load(*configPath)
	if err != nil {
		logrus.WithError(err).Fatal("Failed to load configuration")
	}

	hyprIPC, err := hypr.NewIPC()
	if err != nil {
		logrus.WithError(err).Fatal("Failed to initialize Hyprland IPC")
	}

	monitorDetector, err := detectors.NewMonitorDetector(hyprIPC)
	if err != nil {
		logrus.WithError(err).Fatal("Failed to initialize MonitorDetector")
	}

	powerDetector, err := detectors.NewPowerDetector(ctx)
	if err != nil {
		logrus.WithError(err).Fatal("Failed to initialize PowerDetector")
	}

	matcher := matchers.NewMatcher(cfg)

	generator := generators.NewConfigGenerator()

	svc := service.NewService(cfg, monitorDetector, powerDetector, &service.Config{
		DryRun: *dryRun,
	}, matcher, generator)

	signalHandler := signal.NewHandler(ctx, cancel)

	if err := run(ctx, svc, hyprIPC, monitorDetector, powerDetector, signalHandler); err != nil {
		logrus.WithError(err).Fatal("Service failed")
	}
}

func run(ctx context.Context, svc *service.Service, hyprIPC *hypr.IPC, monitorDetector *detectors.MonitorDetector, powerDetector *detectors.PowerDetector, signalHandler *signal.Handler) error {
	signalHandler.Start(svc)
	defer signalHandler.Stop()

	eg, ctx := errgroup.WithContext(ctx)

	eg.Go(func() error {
		logrus.Debug("Starting Hypr IPC")
		if err := hyprIPC.Run(ctx); err != nil && err != context.Canceled {
			logrus.WithError(err).Error("Hypr IPC failed")
			return err
		}
		logrus.Debug("Hypr IPC finished")
		return nil
	})

	eg.Go(func() error {
		logrus.Debug("Starting monitor detector")
		if err := monitorDetector.Run(ctx); err != nil && err != context.Canceled {
			logrus.WithError(err).Error("Monitor detector failed")
			return err
		}
		logrus.Debug("Monitor detector finished")
		return nil
	})

	eg.Go(func() error {
		logrus.Debug("Starting power detector")
		if err := powerDetector.Run(ctx); err != nil && err != context.Canceled {
			logrus.WithError(err).Error("Power detector failed")
			return err
		}
		logrus.Debug("Power detector finished")
		return nil
	})

	eg.Go(func() error {
		logrus.Debug("Starting service")
		if err := svc.Run(ctx); err != nil && err != context.Canceled {
			logrus.WithError(err).Error("Service failed")
			return err
		}
		logrus.Debug("Service finished")
		return nil
	})

	eg.Go(func() error {
		<-ctx.Done()
		logrus.Debug("Context cancelled, shutting down")
		return ctx.Err()
	})

	if err := eg.Wait(); err != nil && err != context.Canceled {
		return err
	}

	logrus.Info("Shutdown complete")
	return nil
}
