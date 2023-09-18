package cmd

import (
	authv1 "buf.build/gen/go/gportal/gportal-cloud/protocolbuffers/go/gpcloud/api/auth/v1"
	cloudv1 "buf.build/gen/go/gportal/gportal-cloud/protocolbuffers/go/gpcloud/api/cloud/v1"
	"context"
	"github.com/G-PORTAL/gpcloud-go/pkg/gpcloud/client"
	"github.com/G-PORTAL/gpcloud-go/pkg/gpcloud/client/auth"
	"github.com/spf13/cobra"
	"gopkg.in/op/go-logging.v1"
	"gpcloud-cli/pkg/config"
	"log"
	"os"
)

type interactiveSession struct {
	config *config.SessionConfig
	user   *cloudv1.User
	conn   *client.Client
}

// ContextWithSession returns a context with the session, connection and user set
func (i *interactiveSession) ContextWithSession(ctx context.Context) context.Context {
	ctx = context.WithValue(ctx, "session", i.config)
	ctx = context.WithValue(ctx, "user", i.user)
	ctx = context.WithValue(ctx, "conn", i.conn)
	return ctx

}

var activeSession *interactiveSession

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
		PersistentPreRun: func(cobraCmd *cobra.Command, args []string) {
			if activeSession != nil {
				cobraCmd.SetContext(activeSession.ContextWithSession(cobraCmd.Context()))
				return
			}
			activeSession = &interactiveSession{}
			cobraCmd.SetOut(cobraCmd.OutOrStdout())

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

			session, err := config.GetSessionConfig()
			if err != nil {
				panic(err)
			}
			activeSession.config = session

			// Override endpoint if GPCLOUD_ENDPOINT is set
			if os.Getenv("GPCLOUD_ENDPOINT") != "" {
				endpoint = os.Getenv("GPCLOUD_ENDPOINT")
			}

			conn, err := client.NewClient(
				&auth.ProviderKeycloakClientAuth{
					ClientID:     session.ClientID,     // Set your Client ID
					ClientSecret: session.ClientSecret, // Set your Client Secret
				},
				client.EndpointOverrideOption(endpoint),
			)
			if err != nil {
				panic(err)
			}
			activeSession.conn = conn
			// Set user
			resp, err := conn.AuthClient().GetUser(cobraCmd.Context(), &authv1.GetUserRequest{})
			if err != nil {
				panic(err)
			}
			activeSession.user = resp.GetUser()
			if verbose {
				if resp.GetUser().Type == cloudv1.UserType_USER_TYPE_SERVICE_ACCOUNT {
					log.Println("Logged in with a service account")
				} else {
					// TODO: Print user metadata
				}
			}
			// Set context from activeSession
			cobraCmd.SetContext(activeSession.ContextWithSession(cobraCmd.Context()))
		},
	}

	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose mode")
	rootCmd.PersistentFlags().StringVarP(&endpoint, "endpoint", "e", client.DefaultEndpoint, "set API endpoint")

	dirname, _ := os.UserHomeDir()
	rootCmd.PersistentFlags().StringVarP(&config.Path, "config", "c", dirname+"/.gpc.yaml", "define config file location")

	rootCmd.PersistentFlags().BoolVarP(&config.JSONOutput, "json", "j", false, "output as JSON")
	rootCmd.PersistentFlags().BoolVarP(&config.CSVOutput, "csv", "x", false, "output as CSV")

	rootCmd.Flags().BoolVarP(&version, "version", "V", false, "print version information and quit")
	InteractiveCLICommand(&rootCmd)
	AddGeneratedCommands(&rootCmd)

	return &rootCmd
}
