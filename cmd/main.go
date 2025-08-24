package main

import (
	"flag"

	"github.com/fiffeek/hyprdynamicmonitors/internal/config"
	"github.com/fiffeek/hyprdynamicmonitors/internal/detectors"
	"github.com/fiffeek/hyprdynamicmonitors/internal/generators"
	"github.com/fiffeek/hyprdynamicmonitors/internal/hypr"
	"github.com/fiffeek/hyprdynamicmonitors/internal/matchers"
	"github.com/fiffeek/hyprdynamicmonitors/internal/service"
	"github.com/sirupsen/logrus"
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

	if err := svc.Run(); err != nil {
		logrus.WithError(err).Fatal("Service failed")
	}
}
