package cmd

import (
	"github.com/G-PORTAL/gpcloud-go/pkg/gpcloud/client"
	"github.com/spf13/cobra"
	"gpcloud-cli/pkg/config"
)

func New() *cobra.Command {
	rootCmd := cobra.Command{
		Use:   "gpc",
		Short: "gpc is the command line tool for interacting with the GPCore API",
		Long:  "gpc is the command line tool for interacting with the GPCore API\nAuthenticate using the 'gpc auth' command.",
		RunE: func(cobraCmd *cobra.Command, args []string) error {
			if version {
				cobraCmd.Print(GetVersionDisplay())
				return nil
			}
			cobraCmd.Println(cobraCmd.UsageString())
			return nil
		},
	}

	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose mode")
	rootCmd.PersistentFlags().StringVarP(&endpoint, "endpoint", "e", client.DefaultEndpoint, "set API endpoint")

	//dirname, _ := os.UserHomeDir()
	//rootCmd.PersistentFlags().StringVarP(&config.Path, "config", "c", dirname+"/.gpc.yaml", "define config file location")

	rootCmd.PersistentFlags().BoolVarP(&config.JSONOutput, "json", "j", false, "output as JSON")
	rootCmd.PersistentFlags().BoolVarP(&config.CSVOutput, "csv", "x", false, "output as CSV")

	rootCmd.Flags().BoolVarP(&version, "version", "V", false, "print version information and quit")
	InteractiveCLICommand(&rootCmd)
	SelfupdateCommand(&rootCmd)
	LiveLogCommand(&rootCmd)
	AddGeneratedCommands(&rootCmd)

	return &rootCmd
}
