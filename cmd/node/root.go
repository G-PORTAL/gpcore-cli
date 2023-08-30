package node

import (
	"fmt"
	"github.com/spf13/cobra"
	"gpcloud-cli/pkg/config"
)

var RootNodesCommand = &cobra.Command{
	Use:                   "node",
	Short:                 "Utility to combine multiple nodes api actions",
	Long:                  `Utility to combine multiple nodes api actions`,
	DisableFlagsInUseLine: true,
	Args:                  cobra.MatchAll(cobra.ExactArgs(0), cobra.OnlyValidArgs),
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Context is not set in PersistentPreRunE, so we need to get the session config manually
		session, err := config.GetSessionConfig()
		if err != nil {
			return err
		}
		if session.CurrentProject == nil {
			return fmt.Errorf("no project selected")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Usage()
	},
}
