package project

import (
	"github.com/spf13/cobra"
	"gpcloud-cli/pkg/config"
)

var RootProjectCommand = &cobra.Command{
	Use:                   "project",
	Short:                 "Utility to combine multiple project api actions",
	Long:                  `Utility to combine multiple project api actions`,
	DisableFlagsInUseLine: true,
	Args:                  cobra.MatchAll(cobra.ExactArgs(0), cobra.OnlyValidArgs),
	RunE: func(cmd *cobra.Command, args []string) error {
		session := cmd.Context().Value("session").(*config.SessionConfig)
		if session.CurrentProject != nil {
			cmd.Printf("Current project: %s\n\n", *session.CurrentProject)
		} else {
			cmd.Printf("No project selected\n\n")
		}
		return cmd.Usage()
	},
}
