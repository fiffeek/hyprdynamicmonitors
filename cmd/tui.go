package cmd

import (
	"context"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/fiffeek/hyprdynamicmonitors/internal/config"
	"github.com/fiffeek/hyprdynamicmonitors/internal/hypr"
	"github.com/fiffeek/hyprdynamicmonitors/internal/profilemaker"
	"github.com/fiffeek/hyprdynamicmonitors/internal/tui"
	"github.com/fiffeek/hyprdynamicmonitors/internal/utils"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var mockedHyprMonitors string

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

		cfg, err := config.NewConfig(configPath)
		if err != nil {
			logrus.WithError(err).Error("cant read config, ignoring")
		}

		monitors, maker, err := getDeps(ctx, cfg)
		if err != nil {
			return fmt.Errorf("cant get the current monitors spec: %w", err)
		}

		// TODO listen to monitor, config events
		model := tui.NewModel(cfg, monitors, maker)
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

func getDeps(ctx context.Context, cfg *config.Config) (hypr.MonitorSpecs, *profilemaker.Service, error) {
	if mockedHyprMonitors != "" {
		//nolint:gosec
		contents, err := os.ReadFile(mockedHyprMonitors)
		if err != nil {
			return nil, nil, fmt.Errorf("cant read the mocked hypr monitors file: %w", err)
		}

		var res hypr.MonitorSpecs
		if err := utils.UnmarshalResponse(contents, &res); err != nil {
			return nil, nil, fmt.Errorf("failed to parse contents: %w", err)
		}

		profileMaker := profilemaker.NewService(cfg, nil)

		return res, profileMaker, nil
	}
	hyprIPC, err := hypr.NewIPC(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to connect to Hyprland IPC: %w", err)
	}

	monitors := hyprIPC.GetConnectedMonitors()
	if err := monitors.Validate(); err != nil {
		return nil, nil, fmt.Errorf("failed to get valid monitor information: %w", err)
	}

	return monitors, profilemaker.NewService(cfg, hyprIPC), nil
}

func init() {
	rootCmd.AddCommand(tuiCmd)

	tuiCmd.Flags().StringVar(
		&mockedHyprMonitors,
		"hypr-monitors-override",
		"",
		"When used it fill parse the given file as hyprland monitors spec, used for testing.",
	)
}
