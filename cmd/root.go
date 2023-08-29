package cmd

import (
	"fmt"
	"github.com/G-PORTAL/gpcloud-cli/pkg/config"
	"github.com/G-PORTAL/gpcloud-cli/pkg/consts"
	"github.com/G-PORTAL/gpcloud-go/pkg/gpcloud/client"
	"github.com/spf13/cobra"
)

var printVersion = false

func New() *cobra.Command {
	rootCmd := cobra.Command{
		Use:   consts.BinaryName,
		Short: fmt.Sprintf("%s is the command line tool for interacting with the GP Cloud API", consts.BinaryName),
		Long:  fmt.Sprintf("%s is the command line tool for interacting with the GPCore API\nAuthenticate using the '%s auth' command.", consts.BinaryName, consts.BinaryName),
		RunE: func(cobraCmd *cobra.Command, args []string) error {
			if printVersion {
				cobraCmd.Print(GetVersionDisplay())
				return nil
			}
			cobraCmd.Println(cobraCmd.UsageString())
			return nil
		},
	}

	// Application information
	rootCmd.Flags().BoolVarP(&printVersion, "version", "V", false, "print version information and quit")

	// GPCloud API
	// TODO: Will set on first run (when agent starts),the following client calls will ignore these, so, move this to the agent only or reconnect the API on every change
	rootCmd.PersistentFlags().StringVarP(&config.Endpoint, "endpoint", "e", client.DefaultEndpoint, "set API endpoint")

	// Output formats and verbosity
	rootCmd.PersistentFlags().BoolVarP(&config.Verbose, "verbose", "v", false, "verbose mode")
	rootCmd.PersistentFlags().BoolVarP(&config.JSONOutput, "json", "j", false, "output as JSON")
	rootCmd.PersistentFlags().BoolVarP(&config.CSVOutput, "csv", "x", false, "output as CSV")

	InteractiveCLICommand(&rootCmd)
	SelfupdateCommand(&rootCmd)
	LiveLogCommand(&rootCmd)
	AddGeneratedCommands(&rootCmd)

	return &rootCmd
}
