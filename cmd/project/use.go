package project

// TODO: File needs to be generated instead of being statically defined

import (
	cloudv1 "buf.build/gen/go/gportal/gportal-cloud/protocolbuffers/go/gpcloud/api/cloud/v1"
	"fmt"
	"github.com/G-PORTAL/gpcloud-go/pkg/gpcloud/client"
	"github.com/spf13/cobra"
	"gpcloud-cli/pkg/config"
)

var id string
var name string
var useCmd = &cobra.Command{
	Use:                   "use",
	Short:                 "Selects a project to use",
	Long:                  "Selects a project to use",
	DisableFlagsInUseLine: true,
	Args:                  cobra.MatchAll(cobra.ExactArgs(0), cobra.OnlyValidArgs),
	RunE: func(cmd *cobra.Command, args []string) error {
		conn := cmd.Context().Value("conn").(*client.Client)
		session := cmd.Context().Value("session").(*config.SessionConfig)
		resp, err := conn.CloudClient().ListProjects(cmd.Context(), &cloudv1.ListProjectsRequest{})
		if err != nil {
			return err
		}
		for _, project := range resp.Projects {
			if project.Id == id || project.Name == name {
				session.CurrentProject = &id
				if err := session.Write(); err != nil {
					return err
				}
				cmd.Println("Active project is now: " + project.Name)
				return nil
			}
		}
		return fmt.Errorf("project with id %s not found", id)
	},
}

func init() {
	RootProjectCommand.AddCommand(useCmd)
	useCmd.PersistentFlags().StringVarP(&id, "id", "", "", "Specify ID of Project to use")
	useCmd.PersistentFlags().StringVarP(&name, "name", "", "", "Specify name of Project to use")
}
