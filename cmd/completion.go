package cmd

import "github.com/spf13/cobra"

func CompletionCommand(rootCmd *cobra.Command) {
	rootCmd.AddCommand(&cobra.Command{
		Use:       "completion [SHELL]",
		Short:     "Prints shell completion scripts",
		Long:      "Prints shell completion scripts",
		ValidArgs: []string{"bash", "zsh", "fish", "powershell"},
		Annotations: map[string]string{
			"commandType": "main",
		},
		Args: cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
		RunE: func(cmd *cobra.Command, args []string) error {
			switch args[0] {
			case "bash":
				_ = rootCmd.GenBashCompletion(cmd.OutOrStdout())
			case "zsh":
				_ = rootCmd.GenZshCompletion(cmd.OutOrStdout())
			case "fish":
				_ = rootCmd.GenFishCompletion(cmd.OutOrStdout(), true)
			case "powershell":
				_ = rootCmd.GenPowerShellCompletion(cmd.OutOrStdout())
			}

			return nil
		},
	})
}
