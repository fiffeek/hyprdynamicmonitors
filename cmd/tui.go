package cmd

import (
	"context"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/fiffeek/hyprdynamicmonitors/internal/app"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	mockedHyprMonitors string
	enablePowerEvents  bool
)

var tuiCmd = &cobra.Command{
	Use:   "tui",
	Short: "Launch interactive TUI for monitor configuration",
	Long:  `Launch an interactive terminal-based TUI for managing monitor configurations.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if debug {
			logrus.Debug("Running a debug log for tea")
			f, err := tea.LogToFile("debug.log", "debug")
			if err != nil {
				fmt.Println("fatal:", err)
				os.Exit(1)
			}
			logrus.SetOutput(f)
			defer f.Close()
		}

		ctx, cancel := context.WithCancelCause(context.Background())
		defer cancel(context.Canceled)

		app, err := app.NewTUI(ctx, configPath, mockedHyprMonitors, Version, enablePowerEvents, connectToSessionBus)
		if err != nil {
			return fmt.Errorf("cant init tui: %w", err)
		}

		return app.Run(ctx, cancel)
	},
}

func init() {
	rootCmd.AddCommand(tuiCmd)

	tuiCmd.Flags().StringVar(
		&mockedHyprMonitors,
		"hypr-monitors-override",
		"",
		"When used it fill parse the given file as hyprland monitors spec, used for testing.",
	)

	tuiCmd.Flags().BoolVar(
		&enablePowerEvents,
		"enable-power-events",
		false,
		"Enable power events (dbus), needs proper configuration",
	)

	tuiCmd.Flags().BoolVar(
		&connectToSessionBus,
		"connect-to-session-bus",
		false,
		"Connect to session bus instead of system bus",
	)
}
