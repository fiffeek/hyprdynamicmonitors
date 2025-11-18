package cmd

import (
	"fmt"

	"github.com/fiffeek/hyprdynamicmonitors/internal/config"
	"github.com/fiffeek/hyprdynamicmonitors/internal/prepare"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var prepareCmd = &cobra.Command{
	Use:   "prepare",
	Short: "Clean up monitor configuration before daemon start",
	Long: `Prepares the monitor configuration file by removing monitor disable commands.

This command should be run before starting Hyprland to ensure
a clean state. It removes any 'monitor=...,disable' lines from the destination file,
which prevents an issue with no active displays (Hyprland does not launch).

This is particularly useful when running with systemd, where you want to ensure the
monitor configuration is reset before the daemon starts managing profiles.

Example:
  hyprdynamicmonitors prepare`,
	RunE: func(cmd *cobra.Command, args []string) error {
		logrus.WithField("version", Version).Debug("Starting Hyprland Dynamic Monitor Manager")

		cfg, err := config.NewConfig(configPath)
		if err != nil {
			return fmt.Errorf("cant get config: %w", err)
		}

		service := prepare.NewService(cfg)
		err = service.TruncateDestination()
		if err != nil {
			return fmt.Errorf("cant prepare environment prior to the hdm run: %w", err)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(prepareCmd)
}
