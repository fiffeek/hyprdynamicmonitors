package cmd

import (
	"context"
	"fmt"

	"github.com/fiffeek/hyprdynamicmonitors/internal/app"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	runOnce              bool
	dryRun               bool
	disableAutoHotReload bool
	connectToSessionBus  bool
	disablePowerEvents   bool
	enableLidEvents      bool
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run the monitor configuration service",
	Long:  `Run the HyprDynamicMonitors service to continuously monitor for display changes and automatically apply matching configuration profiles.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		logrus.WithField("version", Version).Debug("Starting Hyprland Dynamic Monitor Manager")
		ctx, cancel := context.WithCancelCause(context.Background())
		app, err := app.NewApplication(&configPath, &dryRun, ctx, cancel, &disablePowerEvents,
			&disableAutoHotReload, &connectToSessionBus, &enableLidEvents)
		if err != nil {
			return fmt.Errorf("cant create application: %w", err)
		}

		if runOnce {
			if err := app.RunOnce(ctx); err != nil {
				return fmt.Errorf("error while running: %w", err)
			}
			return nil
		}

		return app.Run(ctx)
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
	runCmd.Flags().BoolVar(
		&dryRun,
		"dry-run",
		false,
		"Show what would be done without making changes",
	)
	runCmd.Flags().BoolVar(
		&runOnce,
		"run-once",
		false,
		"Run once and exit immediately",
	)
	runCmd.Flags().BoolVar(
		&disablePowerEvents,
		"disable-power-events",
		false,
		"Disable power events (dbus)",
	)
	runCmd.Flags().BoolVar(
		&connectToSessionBus,
		"connect-to-session-bus",
		false,
		"Connect to session bus instead of system bus for power events: https://wiki.archlinux.org/title/D-Bus. You can switch as long as you expose power line events in your user session bus.",
	)
	runCmd.Flags().BoolVar(
		&disableAutoHotReload,
		"disable-auto-hot-reload",
		false,
		"Disable automatic hot reload (no file watchers)",
	)
	runCmd.Flags().BoolVar(
		&enableLidEvents,
		"enable-lid-events",
		false,
		"Enable listening to dbus lid events",
	)
}
