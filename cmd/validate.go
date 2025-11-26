package cmd

import (
	"github.com/fiffeek/hyprdynamicmonitors/internal/config"
	"github.com/fiffeek/hyprdynamicmonitors/internal/generators"
	"github.com/fiffeek/hyprdynamicmonitors/internal/utils"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate configuration file",
	Long:  `Validate the configuration file for syntax errors and logical consistency.`,
	Run: func(cmd *cobra.Command, args []string) {
		logrus.WithField("config_path", configPath).Debug("Validating configuration")

		cfg, err := config.NewConfig(configPath)
		if err != nil {
			utils.PrettyPrintError(err)
			logrus.Fatal("Configuration validation failed")
			return
		}

		_, err = generators.NewConfigGenerator(cfg)
		if err != nil {
			utils.PrettyPrintError(err)
			logrus.Fatal("Configuration validation failed")
			return
		}

		logrus.Info("Configuration is valid")
	},
}

func init() {
	rootCmd.AddCommand(validateCmd)
}
