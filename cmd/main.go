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
	"slices"
	"strings"

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
		debug                = flag.Bool("debug", false, "Enable debug logging")
		verbose              = flag.Bool("verbose", false, "Enable verbose logging")
		showVer              = flag.Bool("version", false, "Show version information")
		disablePowerEvents   = flag.Bool("disable-power-events", false, "Disable power events (dbus)")
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

	if command == commandValidate {
		validateConfig(*configPath)
		return
	}

	logrus.WithField("version", Version).Debug("Starting Hyprland Dynamic Monitor Manager")

	ctx, cancel := context.WithCancelCause(context.Background())
	err := createApplication(configPath, dryRun, ctx, cancel, disablePowerEvents, disableAutoHotReload)
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

func createApplication(configPath *string, dryRun *bool, ctx context.Context,
	cancel context.CancelCauseFunc, disablePowerEvents, disableAutoHotReload *bool,
) error {
	cfg, err := config.NewConfig(*configPath)
	if err != nil {
		logrus.WithError(err).Fatal("Failed to load configuration")
	}

	hyprIPC, err := hypr.NewIPC()
	if err != nil {
		logrus.WithError(err).Fatal("Failed to initialize Hyprland IPC")
	}

	fswatcher := filewatcher.NewService(cfg, disableAutoHotReload)

	var conn *dbus.Conn
	if !*disablePowerEvents {
		conn, err = dbus.ConnectSystemBus()
		if err != nil {
			logrus.WithError(err).Fatal("Cant connect to system bus")
		}
	}
	powerDetector, err := power.NewPowerDetector(ctx, cfg, conn, *disablePowerEvents)
	if err != nil {
		logrus.WithError(err).Fatal("Failed to initialize PowerDetector")
	}

	matcher := matchers.NewMatcher()

	generator := generators.NewConfigGenerator(cfg)
	notifications := notifications.NewService(cfg)

	svc := userconfigupdater.NewService(cfg, hyprIPC, powerDetector, &userconfigupdater.Config{
		DryRun: *dryRun,
	}, matcher, generator, notifications)

	reloader := reloader.NewService(cfg, fswatcher, powerDetector, svc, *disableAutoHotReload)

	signalHandler := signal.NewHandler(cancel, reloader, svc)

	if err := run(ctx, svc, hyprIPC, powerDetector, signalHandler, fswatcher, reloader); err != nil {
		return err
	}

	return nil
}

func run(ctx context.Context, svc *userconfigupdater.Service, hyprIPC *hypr.IPC,
	powerDetector *power.PowerDetector, signalHandler *signal.Handler, fswatcher *filewatcher.Service,
	reloader *reloader.Service,
) error {
	eg, ctx := errgroup.WithContext(ctx)

	backgroundGoroutines := []struct {
		Fun  func(context.Context) error
		Name string
	}{
		{Fun: signalHandler.Run, Name: "signal handler"},
		{Fun: fswatcher.Run, Name: "filewatcher"},
		{Fun: hyprIPC.RunEventLoop, Name: "hypr ipc"},
		{Fun: powerDetector.Run, Name: "power detector dbus"},
		{Fun: reloader.Run, Name: "reloader"},
		{Fun: svc.Run, Name: "main service"},
	}
	for _, bg := range backgroundGoroutines {
		eg.Go(func() error {
			fields := logrus.Fields{"name": bg.Name, "fun": utils.GetFunctionName(bg.Fun)}
			logrus.WithFields(fields).Debug("Starting")
			if err := bg.Fun(ctx); err != nil {
				logrus.WithFields(fields).Debug("Exited with error")
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
