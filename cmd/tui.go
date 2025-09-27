package cmd

import (
	"context"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/fiffeek/hyprdynamicmonitors/internal/config"
	"github.com/fiffeek/hyprdynamicmonitors/internal/hypr"
	"github.com/fiffeek/hyprdynamicmonitors/internal/tui"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
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
			defer f.Close()
		}

		ctx, cancel := context.WithCancelCause(context.Background())
		defer cancel(context.Canceled)

		cfg, err := config.NewConfig(configPath)
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		hyprIPC, err := hypr.NewIPC(ctx)
		if err != nil {
			return fmt.Errorf("failed to connect to Hyprland IPC: %w", err)
		}

		monitors := hyprIPC.GetConnectedMonitors()
		if err := monitors.Validate(); err != nil {
			return fmt.Errorf("failed to get valid monitor information: %w", err)
		}

		model := tui.NewModel(cfg, monitors)
		program := tea.NewProgram(
			model,
			tea.WithAltScreen(),
			tea.WithMouseCellMotion(),
		)

		if _, err := program.Run(); err != nil {
			return fmt.Errorf("failed to run TUI: %w", err)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(tuiCmd)
}
