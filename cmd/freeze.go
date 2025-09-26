package cmd

import (
	"context"
	"errors"
	"fmt"

	"github.com/fiffeek/hyprdynamicmonitors/internal/config"
	"github.com/fiffeek/hyprdynamicmonitors/internal/hypr"
	"github.com/fiffeek/hyprdynamicmonitors/internal/profilemaker"
	"github.com/spf13/cobra"
)

var (
	profileName         string
	profileFileLocation string
)

var freezeCmd = &cobra.Command{
	Use:   "freeze",
	Short: "Freeze current monitor configuration as a new profile template",
	Long: `Freeze the current Hyprland monitor configuration and save it as a new profile template.

This command captures your current monitor setup and creates two artifacts:
1. A Go template file containing the Hyprland configuration
2. A new profile entry in your configuration file that references this template

TEMPLATE FILE:
The Go template will be saved to hyprconfigs/{profile-name}.go.tmpl by default, or to a
custom location specified with --config-file-location. This template can be edited after
creation to customize the configuration.

PROFILE ENTRY:
A new profile with the specified name will be appended to your configuration file. The
profile will automatically require monitors by description (not name) to ensure better
portability across different systems.

PREREQUISITES:
- The profile name must not already exist in your configuration (it will be checked)
- The template file location must not exist (it will be created)
- Hyprland must be running with a valid monitor configuration

This is useful for quickly creating new profiles based on your current working setup.`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if profileFileLocation == "" && profileName != "" {
			profileFileLocation = fmt.Sprintf("hyprconfigs/%s.go.tmpl", profileName)
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		if profileName == "" {
			return errors.New("profile-name can't be empty")
		}
		ctx, cancel := context.WithCancelCause(context.Background())
		defer cancel(context.Canceled)

		cfg, err := config.NewConfig(configPath)
		if err != nil {
			return fmt.Errorf("the current config is not valid: %w", err)
		}

		hyprIPC, err := hypr.NewIPC(ctx)
		if err != nil {
			return fmt.Errorf("failed to initialize Hyprland IPC: %w", err)
		}

		profile := profilemaker.NewService(cfg, hyprIPC)
		if err = profile.FreezeCurrentAs(profileName, profileFileLocation); err != nil {
			return fmt.Errorf("cant freeze the current settings as a new profile: %w", err)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(freezeCmd)

	freezeCmd.Flags().StringVar(
		&profileName,
		"profile-name",
		"",
		"What profile name to set the frozen profile to.",
	)
	_ = freezeCmd.MarkFlagRequired("profile-name")
	freezeCmd.Flags().StringVar(
		&profileFileLocation,
		"config-file-location",
		"",
		"Where to put the generated config file template (defaults to hyprconfigs/$PROFILE_NAME.go.tmpl)",
	)
}
