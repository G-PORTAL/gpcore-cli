package user

import (
	"github.com/G-PORTAL/gpcore-cli/pkg/api"
	"github.com/G-PORTAL/gpcore-cli/pkg/client"
	"github.com/G-PORTAL/gpcore-cli/pkg/config"
	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"
)

var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Logout a previously impersonated user",
	Long:  "Logout a previously impersonated user",
	RunE: func(cobraCmd *cobra.Command, args []string) error {
		ctx := client.ExtractContext(cobraCmd)
		user := client.GetUserFromContext(ctx)
		sessionConfig, err := config.GetSessionConfig()
		if err != nil {
			return err
		}

		// If the impersonate access token is already nil, we no longer
		// impersonating anybody.
		isImpersonated, err := client.IsImpersonated()
		if err != nil {
			return err
		}

		if !isImpersonated {
			// Even when not impersonating, clear any lingering active project so
			// a selection left over from a previous (expired) impersonation does
			// not leak into the current user context.
			if sessionConfig.CurrentProject != nil {
				sessionConfig.CurrentProject = nil
				if err = sessionConfig.Write(); err != nil {
					return err
				}
				if err = config.RefreshSessionConfig(); err != nil {
					return err
				}
				cobraCmd.Println("Cleared a lingering active project from a previous session.")
			}
			cobraCmd.Println("No need to logout, you do not impersonate anybody.")
			return nil
		}

		sessionConfig.ImpersonateAccessToken = nil
		sessionConfig.ImpersonateExpiresIn = nil
		sessionConfig.CurrentProject = nil

		err = sessionConfig.Write()
		if err != nil {
			return err
		}

		api.RenewAPISession()

		log.Infof("No longer impersonating (id: %s) %s", user.GetId(), user.GetUsername())
		cobraCmd.Printf("No longer impersonating user %s\n", user.GetUsername())

		return nil
	},
}

func init() {
	if config.HasAdminConfig() {
		RootUserCommand.AddCommand(logoutCmd)
	}
}
