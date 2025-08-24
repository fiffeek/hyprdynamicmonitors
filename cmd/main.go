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

	logrus.WithField("version", version).Info("Starting Hyprland Dynamic Monitor Manager")

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

	powerDetector, err := detectors.NewPowerDetector()
	if err != nil {
		logrus.WithError(err).Fatal("Failed to initialize PowerDetector")
	}

	matcher := matchers.NewMatcher(cfg)

	generator := generators.NewConfigGenerator()

	svc := service.NewService(cfg, monitorDetector, powerDetector, &service.Config{
		DryRun: *dryRun,
	}, matcher, generator)

	run(svc, hyprIPC, monitorDetector, powerDetector)
}

func run(svc *service.Service, hyprIPC *hypr.IPC, monitorDetector *detectors.MonitorDetector, powerDetector *detectors.PowerDetector) {
	ctx, cancel := context.WithCancel(context.Background())
	signalHandler := signal.NewHandler(ctx, cancel)
	logrus.Info("Created signal handler")
	signalHandler.Start(svc)
	logrus.Info("Started signal handler")
	defer signalHandler.Stop()
	logrus.Info("Signal handlers registered")

	eg, ctx := errgroup.WithContext(ctx)

	eg.Go(func() error {
		logrus.Info("Starting Hypr IPC")
		if err := hyprIPC.Run(ctx); err != nil && err != context.Canceled {
			logrus.WithError(err).Error("Hypr IPC failed")
			return err
		}
		logrus.Info("Hypr IPC finished")
		return nil
	})

	eg.Go(func() error {
		logrus.Info("Starting monitor detector")
		if err := monitorDetector.Run(ctx); err != nil && err != context.Canceled {
			logrus.WithError(err).Error("Monitor detector failed")
			return err
		}
		logrus.Info("Monitor detector finished")
		return nil
	})

	eg.Go(func() error {
		logrus.Info("Starting power detector")
		if err := powerDetector.Run(ctx); err != nil && err != context.Canceled {
			logrus.WithError(err).Error("Power detector failed")
			return err
		}
		logrus.Info("Power detector finished")
		return nil
	})

	eg.Go(func() error {
		logrus.Info("Starting service")
		if err := svc.Run(ctx); err != nil && err != context.Canceled {
			logrus.WithError(err).Error("Service failed")
			return err
		}
		logrus.Info("Service finished")
		return nil
	})

	eg.Go(func() error {
		<-ctx.Done()
		logrus.Info("Context cancelled, shutting down")
		return ctx.Err()
	})

	if err := eg.Wait(); err != nil && err != context.Canceled {
		logrus.WithError(err).Error("Run failed")
	}

	logrus.Info("Shutdown complete")
}
