package cmd

import (
	"context"
	"github.com/G-PORTAL/gpcloud-go/pkg/gpcloud/client"
	"github.com/G-PORTAL/gpcloud-go/pkg/gpcloud/client/auth"
	"github.com/spf13/cobra"
	"gopkg.in/op/go-logging.v1"
	"gpcloud-cli/cmd/project"
	"gpcloud-cli/pkg/config"
	"os"
)

func New() *cobra.Command {
	var rootCmd = &cobra.Command{
		Use:   "gpc",
		Short: "gpc is the command line tool for interacting with the GPCore API",
		Long:  "gpc is the command line tool for interacting with the GPCore API\nAuthenticate using the 'gpc auth' command.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if version {
				cmd.Print(GetVersionDisplay())
				return nil
			}
			cmd.Println(cmd.UsageString())
			return nil
		},
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			cmd.SetOut(cmd.OutOrStdout())

			var format = logging.MustStringFormatter(
				`%{color}%{time:15:04:05} %{shortfunc} [%{level:.4s}]%{color:reset} %{message}`,
			)
			var backend = logging.AddModuleLevel(
				logging.NewBackendFormatter(logging.NewLogBackend(os.Stderr, "", 0), format))

			if verbose {
				backend.SetLevel(logging.DEBUG, "")
			} else {
				backend.SetLevel(logging.ERROR, "")
			}

			logging.SetBackend(backend)

			session, err := config.ReadSessionConfig()
			if err != nil {
				panic(err)
			}

			conn, err := client.NewClient(
				&auth.ProviderKeycloakClientAuth{
					ClientID:     session.ClientID,     // Set your Client ID
					ClientSecret: session.ClientSecret, // Set your Client Secret
				},
			)
			if err != nil {
				panic(err)
			}
			cmd.SetContext(context.WithValue(cmd.Context(), "conn", conn))

		},
	}
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose mode")
	rootCmd.PersistentFlags().StringVarP(&configPath, "config", "c", "./config.yaml", "Define config file location")
	rootCmd.Flags().BoolVarP(&version, "version", "V", false, "Print version information and quit")
	rootCmd.AddCommand(
		project.RootProjectCommand,
	)
	return rootCmd
}
