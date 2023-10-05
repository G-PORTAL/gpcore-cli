package client

import (
	"context"
	"github.com/spf13/cobra"
)

// ExtractContext extracts the context from a cobra command in respect to the
// parent commands. This is needed because cobra does not pass the context
// down to the subcommands. See https://github.com/spf13/cobra/issues/1469
// for more detail.
func ExtractContext(cmd *cobra.Command) context.Context {
	if cmd == nil {
		return nil
	}

	var (
		ctx = cmd.Context()
		p   = cmd.Parent()
	)
	if cmd.Parent() == nil {
		return ctx
	}
	for {
		ctx = p.Context()
		p = p.Parent()
		if p == nil {
			break
		}
	}
	return ctx
}
