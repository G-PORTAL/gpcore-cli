package client

import (
	"context"
	"github.com/spf13/cobra"
)

// https://github.com/spf13/cobra/issues/1469
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
