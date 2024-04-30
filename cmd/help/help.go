package help

import (
	"fmt"
	"github.com/spf13/cobra"
)

// UnknownSubcommandAction is a cobra.Command.RunE function that prints the
// help message for unknown subcommands. If there are suggestions for similar
// subcommands, they will be printed as well.
func UnknownSubcommandAction(cobraCmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return cobraCmd.Help()
	}

	err := fmt.Sprintf("Unknown subcommand %q for %q", args[0], cobraCmd.Name())
	if suggestions := cobraCmd.SuggestionsFor(args[0]); len(suggestions) > 0 {
		err += "\n\nDid you mean this??\n"
		for _, s := range suggestions {
			err += fmt.Sprintf("\t%v\n", s)
		}
	}
	cobraCmd.Println(err)

	return fmt.Errorf(err)
}
