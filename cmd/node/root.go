package node

import (
	"github.com/spf13/cobra"
)

// Note: node subcommands no longer require a globally selected project here.
// Each subcommand accepts an optional --project-id that falls back to the
// project selected via "project use" (and errors if neither is set). This
// avoids forcing "project use" when the project is passed explicitly.
var RootNodeCommand = &cobra.Command{
	Use:                   "node",
	Short:                 "Utility to combine multiple nodes api actions",
	Long:                  `Utility to combine multiple nodes api actions`,
	GroupID:               "resources",
	DisableFlagsInUseLine: true,
	Args:                  cobra.MatchAll(cobra.ExactArgs(0), cobra.OnlyValidArgs),
	RunE: func(cobraCmd *cobra.Command, args []string) error {
		return cobraCmd.Usage()
	},
}
