package project

import (
	"github.com/G-PORTAL/gpcloud-cli/pkg/client"
	"github.com/G-PORTAL/gpcloud-cli/pkg/config"
	"github.com/spf13/cobra"
)

var RootProjectCommand = &cobra.Command{
	Use:                   "project",
	Short:                 "Utility to combine multiple project api actions",
	Long:                  `Utility to combine multiple project api actions`,
	DisableFlagsInUseLine: true,
	Args:                  cobra.MatchAll(cobra.ExactArgs(0), cobra.OnlyValidArgs),
	RunE: func(cobraCmd *cobra.Command, args []string) error {
		ctx := client.ExtractContext(cobraCmd)
		config := ctx.Value("config").(*config.SessionConfig)
		if config.CurrentProject != nil {
			cobraCmd.Printf("Current project: %s\n\n", *config.CurrentProject)
		} else {
			cobraCmd.Printf("No project selected\n\n")
		}
		return cobraCmd.Usage()
	},
}
