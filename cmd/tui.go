package cmd

import (
	"context"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/fiffeek/hyprdynamicmonitors/internal/app"
	"github.com/muesli/termenv"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	mockedHyprMonitors string
	runningUnderTest   bool
)

var tuiCmd = &cobra.Command{
	Use:   "tui",
	Short: "Launch interactive TUI for monitor configuration",
	Long:  `Launch an interactive terminal-based TUI for managing monitor configurations.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if debug {
			f, err := tea.LogToFile("debug.log", "debug")
			if err != nil {
				fmt.Println("fatal:", err)
				os.Exit(1)
			}
			logrus.SetOutput(f)
			defer f.Close()
		} else {
			// disable logging completely for tui unless run in the debug mode
			logrus.SetLevel(logrus.PanicLevel)
		}

		if runningUnderTest {
			lipgloss.SetColorProfile(termenv.Ascii)
		}

		ctx, cancel := context.WithCancelCause(context.Background())
		defer cancel(context.Canceled)

		app, err := app.NewTUI(ctx, configPath, mockedHyprMonitors, Version, disablePowerEvents,
			connectToSessionBus, enableLidEvents, runningUnderTest)
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
		&disablePowerEvents,
		"disable-power-events",
		false,
		"Disable power events (dbus)",
	)

	tuiCmd.Flags().BoolVar(
		&connectToSessionBus,
		"connect-to-session-bus",
		false,
		"Connect to session bus instead of system bus for power events: https://wiki.archlinux.org/title/D-Bus. You can switch as long as you expose power line events in your user session bus.",
	)

	tuiCmd.Flags().BoolVar(
		&enableLidEvents,
		"enable-lid-events",
		false,
		"Enable listening to dbus lid events",
	)

	tuiCmd.Flags().BoolVar(&runningUnderTest, "running-under-test", false,
		"Use test settings such as no styling etc.")
}
