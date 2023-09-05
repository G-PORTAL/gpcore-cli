package project

// TODO: File needs to be generated instead of being statically defined

import (
	cloudv1 "buf.build/gen/go/gportal/gportal-cloud/protocolbuffers/go/gpcloud/api/cloud/v1"
	"github.com/G-PORTAL/gpcloud-go/pkg/gpcloud/client"
	"github.com/spf13/cobra"
	"gpcloud-cli/pkg/config"
)

var listCmd = &cobra.Command{
	Use:                   "list",
	Short:                 "Prints all available projects",
	Long:                  "Prints all available projects",
	DisableFlagsInUseLine: true,
	Args:                  cobra.MatchAll(cobra.ExactArgs(0), cobra.OnlyValidArgs),
	RunE: func(cmd *cobra.Command, args []string) error {
		session := cmd.Context().Value("session").(*config.SessionConfig)
		conn := cmd.Context().Value("conn").(*client.Client)
		resp, err := conn.CloudClient().ListProjects(cmd.Context(), &cloudv1.ListProjectsRequest{})
		if err != nil {
			return err
		}

		// TODO: Put this into a hook or something
		indicator := "-"
		for _, project := range resp.Projects {
			if project.Id == *session.CurrentProject {
				indicator = "*"
			} else {
				indicator = "-"
			}
			cmd.Println(indicator + " " + project.Name)
		}

		//jsonData, err := json.MarshalIndent(resp, "", "  ")
		//if err != nil {
		//	return err
		//}
		//cmd.Println(string(jsonData))
		return nil
	},
}

func init() {
	RootProjectCommand.AddCommand(listCmd)
}
