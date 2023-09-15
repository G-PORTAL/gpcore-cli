package cmd

import (
	"bufio"
	"strings"

	"github.com/spf13/cobra"
	"os"
)

var currentPrefixArgs = []string{}

func printPathFormatted(cmd *cobra.Command) {
	if len(currentPrefixArgs) == 0 {
		cmd.Print("> ")
		return
	}
	currentPathPretty := ""
	for _, arg := range currentPrefixArgs {
		currentPathPretty += "\033[36m" + arg + "\033[0m > "
	}
	cmd.Print(currentPathPretty)
}

func RegisterCLICommand(rootCmd *cobra.Command) {
	rootCmd.AddCommand(&cobra.Command{
		Use:   "interactive",
		Short: "Interactive mode",
		Long: `This is the interactive version of the gpc command line tool
If you prefix commands you can change the current root level
Example:
CLI mode enabled
> ` + "\033[31m.project\033[0m" + `                      # set root level to "project"
project > ` + "\033[31m.list\033[0m" + `                 # add level "list" to the root level "project". "Enter" will now just call the same subcommand again
project > list > ` + "\033[31m..\033[0m" + `             # move up one level
project > ` + "\033[31m.list\033[0m" + `                 # root level "project" with subcommand "list", pressing enter will now just call the same subcommand again
project > list > ` + "\033[31m...\033[0m" + `            # move up two levels (each additional dot moves up one level)
> ` + "\033[31mproject list\033[0m" + `			# run the project list command`,
		DisableFlagsInUseLine: true,
		Args:                  cobra.MatchAll(cobra.ExactArgs(0), cobra.OnlyValidArgs),
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.Print("\033[33mCLI mode enabled\033[0m\n")
			cmd.Printf("> ")
			var reader = bufio.NewReader(os.Stdin)
			for {
				command, _ := reader.ReadString('\n')
				messageParts := append(currentPrefixArgs, strings.Split(strings.TrimSpace(command), " ")...)
				cmd.Printf("Current command: %+v\n", messageParts)

				if len(messageParts) == 0 || messageParts[0] == "" {
					rootCmd.Usage()
					printPathFormatted(cmd)
					continue
				}

				if strings.HasPrefix(command, ".") {
					upper := strings.Count(command, ".") - 1
					if upper > 0 {
						if len(currentPrefixArgs) < upper {
							upper = len(currentPrefixArgs)
						}
						currentPrefixArgs = currentPrefixArgs[:len(currentPrefixArgs)-upper]
					} else {
						newDir := strings.Split(strings.TrimSpace(command), " ")
						newDir[0] = strings.TrimPrefix(newDir[0], ".")
						currentPrefixArgs = append(currentPrefixArgs, newDir...)
					}
					printPathFormatted(cmd)
					continue
				}

				found := false
				for _, subCmd := range rootCmd.Commands() {
					if subCmd.Use == messageParts[0] {
						found = true
						rootCmd.SetArgs(append(messageParts, currentPrefixArgs...))
						if err := subCmd.Execute(); err != nil {
							cmd.Printf("Error executing command %s: %s\n", subCmd.Use, err)
						}
						break
					}
				}
				if !found {
					cmd.Printf("Unknown command %s\n", messageParts[0])
				}

				printPathFormatted(cmd)
			}
			return nil
		},
	})
}
