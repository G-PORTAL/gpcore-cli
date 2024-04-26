package cmd

import (
	"github.com/G-PORTAL/gpcore-cli/pkg/config"
	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"
	"strings"
)

func SetLogLevelCommand(rootCmd *cobra.Command) {
	rootCmd.AddCommand(&cobra.Command{
		Use:       "loglevel",
		Short:     "Set the log level",
		Long:      "Set the log level",
		ValidArgs: []string{"trace", "debug", "info", "warn", "error", "fatal"},
		Args:      cobra.OnlyValidArgs,
		RunE: func(cobraCmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return cobraCmd.Help()
			}

			level := strings.ToUpper(args[0])

			err := config.WriteLogLevel(level)
			if err != nil {
				return err
			}
			config.ActivateLogLevel()

			log.Info("Log level set to: " + level)
			return nil
		},
	})
}
