package project

import "github.com/spf13/cobra"

var RootProjectCommand = &cobra.Command{
	Use:                   "project",
	Short:                 "Utility to combine multiple project api actions",
	Long:                  `Utility to combine multiple project api actions`,
	DisableFlagsInUseLine: true,
	Args:                  cobra.MatchAll(cobra.ExactArgs(0), cobra.OnlyValidArgs),
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Usage()
	},
}
