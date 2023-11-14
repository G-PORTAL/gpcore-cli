package agent

import (
	"github.com/G-PORTAL/gpcore-cli/pkg/config"
	"github.com/spf13/cobra"
)

var admin bool

var setupCommand = &cobra.Command{
	Use:                   "setup",
	Short:                 "Setup config (the old config will be overwritten)",
	Long:                  "Setup config (the old config will be overwritten)",
	DisableFlagsInUseLine: true,
	Args:                  cobra.MatchAll(cobra.ExactArgs(0), cobra.OnlyValidArgs),
	RunE: func(cmd *cobra.Command, args []string) error {
		err := config.SetupConfig()
		if err != nil {
			return err
		}

		// If we've got the admin flag, ask for admin credentials as well
		if admin {
			return config.SetupAdminConfig()
		}

		return nil
	},
}

func init() {
	setupCommand.PersistentFlags().BoolVarP(&admin, "admin", "", false, "Setup admin credentials as well")
	agentCommand.AddCommand(setupCommand)
}
