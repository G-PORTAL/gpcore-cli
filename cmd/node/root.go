package node

import (
	"fmt"
	"github.com/G-PORTAL/gpcloud-cli/pkg/config"
	"github.com/spf13/cobra"
)

var RootNodeCommand = &cobra.Command{
	Use:                   "node",
	Short:                 "Utility to combine multiple nodes api actions",
	Long:                  `Utility to combine multiple nodes api actions`,
	DisableFlagsInUseLine: true,
	Args:                  cobra.MatchAll(cobra.ExactArgs(0), cobra.OnlyValidArgs),
	PersistentPreRunE: func(cobraCmd *cobra.Command, args []string) error {
		// Context is not set in PersistentPreRunE, so we need to get the session config manually
		config, err := config.GetSessionConfig()
		if err != nil {
			return err
		}
		if config.CurrentProject == nil {
			return fmt.Errorf("no project selected")
		}
		return nil
	},
	RunE: func(cobraCmd *cobra.Command, args []string) error {
		return cobraCmd.Usage()
	},
}
