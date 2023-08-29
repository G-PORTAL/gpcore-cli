package project

// TODO: File needs to be generated instead of being statically defined

import (
	cloudv1 "buf.build/gen/go/gportal/gportal-cloud/protocolbuffers/go/gpcloud/api/cloud/v1"
	"encoding/json"
	"github.com/G-PORTAL/gpcloud-go/pkg/gpcloud/client"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:                   "list",
	Short:                 "Prints a list of supported games",
	Long:                  `Prints a list of supported games`,
	DisableFlagsInUseLine: true,
	Args:                  cobra.MatchAll(cobra.ExactArgs(0), cobra.OnlyValidArgs),
	RunE: func(cmd *cobra.Command, args []string) error {
		conn := cmd.Context().Value("conn").(*client.Client)
		resp, err := conn.CloudClient().ListProjects(cmd.Context(), &cloudv1.ListProjectsRequest{})
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
	RootProjectCommand.AddCommand(listCmd)
}
