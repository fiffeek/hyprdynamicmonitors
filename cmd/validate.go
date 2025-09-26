package cmd

import (
	"flag"
	"fmt"
	"strings"

	"github.com/fiffeek/hyprdynamicmonitors/internal/config"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate configuration file",
	Long:  `Validate the configuration file for syntax errors and logical consistency.`,
	Run: func(cmd *cobra.Command, args []string) {
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
	},
}

func init() {
	rootCmd.AddCommand(validateCmd)
}
