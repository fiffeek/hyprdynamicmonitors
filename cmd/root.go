// Package cmd provides the entry point for the hyprdynamicmonitors application.
// It dynamically manages Hyprland monitor configurations based on connected displays.
package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"time"

	"github.com/fiffeek/hyprdynamicmonitors/internal/errs"
	"github.com/fiffeek/hyprdynamicmonitors/internal/signal"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var (
	Version    = "dev"
	Commit     = "none"
	BuildDate  = "unknown"
	BinaryName = "hyprdynamicmonitors"
)

var (
	debug                bool
	verbose              bool
	enableJSONLogsFormat bool
	configPath           string
	rootCmd              = &cobra.Command{
		Use:              BinaryName,
		Short:            "Dynamically manage Hyprland monitor configurations",
		Long:             "HyprDynamicMonitors is a service that automatically switches between predefined Hyprland monitor configuration profiles based on connected monitors and power state.",
		Version:          fmt.Sprintf("%s (commit %s, built %s)", Version, Commit, BuildDate),
		PersistentPreRun: setupLogger,
		SilenceErrors:    true,
		SilenceUsage:     true,
	}
)

func Execute() {
	cmd, _, err := rootCmd.Find(os.Args[1:])

	if err == nil && cmd.Use == rootCmd.Use && !errors.Is(cmd.Flags().Parse(os.Args[1:]), pflag.ErrHelp) &&
		!slices.Contains(os.Args[1:], "--version") && !slices.Contains(os.Args[1:], "-v") {
		args := append([]string{runCmd.Use}, os.Args[1:]...)
		rootCmd.SetArgs(args)
	}

	err = rootCmd.Execute()
	var target *signal.Interrupted
	if errors.As(err, &target) {
		logrus.WithError(err).Info("Service interrupted")
		os.Exit(target.ExitCode())
	}
	if errors.Is(err, context.Canceled) {
		logrus.WithError(err).Info("Context cancelled, exiting")
		return
	}
	if errors.Is(err, errs.ErrUPowerMisconfigured) {
		logrus.Warn(`If lid or power events are enabled, UPower is required to run. Start the service.
See: https://hyprdynamicmonitors.filipmikina.com/docs/#runtime-requirements.
Alternatively, disable the power/lid events if you do not expect to use them with --disable-power-events`)
		logrus.WithError(err).Fatal("Is UPower running?")
		return
	}
	if err != nil {
		logrus.WithError(err).Fatal("Service failed")
	}
	logrus.Debug("Exiting...")
}

func setupLogger(cmd *cobra.Command, args []string) {
	if debug {
		logrus.SetLevel(logrus.DebugLevel)
	} else {
		logrus.SetLevel(logrus.InfoLevel)
	}
	if verbose {
		logrus.SetReportCaller(true)
	}

	if enableJSONLogsFormat {
		logrus.SetFormatter(&logrus.JSONFormatter{
			DisableTimestamp: false,
			TimestampFormat:  time.RFC3339Nano,
		})
	} else {
		logrus.SetFormatter(&logrus.TextFormatter{
			DisableTimestamp: false,
			DisableColors:    false,
			TimestampFormat:  time.RFC3339Nano,
			FullTimestamp:    true,
			ForceQuote:       true,
			CallerPrettyfier: func(f *runtime.Frame) (string, string) {
				fn := filepath.Base(f.Function)
				file := fmt.Sprintf("%s:%d", filepath.Base(f.File), f.Line)
				return fn, file
			},
		})
	}
}

func init() {
	rootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "Enable debug logging")
	rootCmd.PersistentFlags().BoolVar(&verbose, "verbose", false, "Enable verbose logging")
	rootCmd.PersistentFlags().StringVar(
		&configPath,
		"config",
		"$HOME/.config/hyprdynamicmonitors/config.toml",
		"Path to configuration file",
	)
	rootCmd.PersistentFlags().BoolVar(&enableJSONLogsFormat, "enable-json-logs-format", false, "Enable structured logging")
}
