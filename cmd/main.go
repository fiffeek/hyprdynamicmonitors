// Package main provides the entry point for the hyprdynamicmonitors application.
// It dynamically manages Hyprland monitor configurations based on connected displays.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"

	"github.com/fiffeek/hyprdynamicmonitors/internal/app"
	"github.com/fiffeek/hyprdynamicmonitors/internal/config"
	"github.com/fiffeek/hyprdynamicmonitors/internal/signal"
	"github.com/sirupsen/logrus"
)

var (
	Version   = "dev"
	Commit    = "none"
	BuildDate = "unknown"
)

const (
	commandValidate = "validate"
	commandRun      = "run"
)

func main() {
	var (
		configPath = flag.String("config", "$HOME/.config/hyprdynamicmonitors/config.toml", "Path to configuration file")
		dryRun     = flag.Bool("dry-run", false,
			"Show what would be done without making changes")
		runOnce              = flag.Bool("run-once", false, "Run once and exit immediately")
		debug                = flag.Bool("debug", false, "Enable debug logging")
		enableJSONLogsFormat = flag.Bool("enable-json-logs-format", false, "Enable structured logging")
		verbose              = flag.Bool("verbose", false, "Enable verbose logging")
		showVer              = flag.Bool("version", false, "Show version information")
		disablePowerEvents   = flag.Bool("disable-power-events", false, "Disable power events (dbus)")
		connectToSessionBus  = flag.Bool("connect-to-session-bus", false,
			"Connect to session bus instead of system bus for power events: https://wiki.archlinux.org/title/D-Bus. You can switch as long as you expose power line events in your user session bus.")
		disableAutoHotReload = flag.Bool("disable-auto-hot-reload", false,
			"Disable automatic hot reload (no file watchers)")
	)
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage: %s [options] [command]\n\n", filepath.Base(flag.CommandLine.Name()))
		fmt.Fprintf(flag.CommandLine.Output(), "Commands:\n")
		fmt.Fprintf(flag.CommandLine.Output(), "  %s      Run the service (default)\n", commandRun)
		fmt.Fprintf(flag.CommandLine.Output(), "  %s Validate configuration file and exit\n\n", commandValidate)
		fmt.Fprintf(flag.CommandLine.Output(), "Options:\n")
		flag.PrintDefaults()
	}
	flag.Parse()

	command := commandRun
	if flag.NArg() > 0 {
		command = flag.Arg(0)
	}

	if !slices.Contains([]string{commandRun, commandValidate}, command) {
		fmt.Fprintf(flag.CommandLine.Output(), "Error: unknown command '%s'\n", command)
		flag.Usage()
		return
	}

	if *debug {
		logrus.SetLevel(logrus.DebugLevel)
	} else {
		logrus.SetLevel(logrus.InfoLevel)
	}
	if *verbose {
		logrus.SetReportCaller(true)
	}
	if *enableJSONLogsFormat {
		logrus.SetFormatter(&logrus.JSONFormatter{
			DisableTimestamp: false,
		})
	} else {
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
	}

	if *showVer {
		logrus.WithFields(logrus.Fields{
			"version":   Version,
			"commit":    Commit,
			"buildDate": BuildDate,
		}).Info("hyprdynamicmonitors version")
		return
	}

	if command == commandValidate {
		validateConfig(*configPath)
		return
	}

	logrus.WithField("version", Version).Debug("Starting Hyprland Dynamic Monitor Manager")

	ctx, cancel := context.WithCancelCause(context.Background())
	app, err := app.NewApplication(configPath, dryRun, ctx, cancel, disablePowerEvents,
		disableAutoHotReload, connectToSessionBus)
	if err != nil {
		logrus.WithError(err).Fatal("Failed on app creation")
	}

	if *runOnce {
		if err := app.RunOnce(ctx); err != nil {
			logrus.WithError(err).Fatal("Run failed")
		}
		return
	}

	err = app.Run(ctx)
	if err == nil {
		logrus.Info("Exiting...")
		return
	}
	var target *signal.Interrupted
	if errors.As(err, &target) {
		logrus.WithError(err).Info("Service interrupted")
		os.Exit(target.ExitCode())
	}
	if errors.Is(err, context.Canceled) {
		logrus.WithError(err).Info("Context cancelled, exiting")
		return
	}

	// otherwise there is a real error
	logrus.WithError(err).Fatal("Service failed")
}

func validateConfig(configPath string) {
	logrus.WithField("config_path", configPath).Debug("Validating configuration")

	_, err := config.NewConfig(configPath)
	if err != nil {
		parts := strings.Split(err.Error(), ": ")
		indent := 0
		for _, part := range parts {
			fmt.Fprintf(flag.CommandLine.Output(), "%s%s\n", strings.Repeat(" ", indent), part)
			indent += 2
		}
		logrus.Fatal("Configuration validation failed")
		return
	}

	logrus.Info("Configuration is valid")
}
