package node

// TODO: File needs to be generated instead of being statically defined

import (
	cloudv1 "buf.build/gen/go/gportal/gportal-cloud/protocolbuffers/go/gpcloud/api/cloud/v1"
	"encoding/json"
	"github.com/G-PORTAL/gpcloud-go/pkg/gpcloud/client"
	"github.com/spf13/cobra"
	"gpcloud-cli/pkg/config"
)

var listCmd = &cobra.Command{
	Use:                   "list",
	Short:                 "Prints a list of supported nodes",
	Long:                  `Prints a list of supported nodes`,
	DisableFlagsInUseLine: true,
	Args:                  cobra.MatchAll(cobra.ExactArgs(0), cobra.OnlyValidArgs),
	RunE: func(cmd *cobra.Command, args []string) error {
		conn := cmd.Context().Value("conn").(*client.Client)
		session := cmd.Context().Value("session").(*config.SessionConfig)
		resp, err := conn.CloudClient().ListNodes(cmd.Context(), &cloudv1.ListNodesRequest{
			Id: *session.CurrentProject,
		})
		if err != nil {
			return err
		}
		jsonData, err := json.MarshalIndent(resp, "", "  ")
		if err != nil {
			return err
		}
		cmd.Println(string(jsonData))
		return nil
	},
}

func init() {
	RootNodesCommand.AddCommand(listCmd)
}
