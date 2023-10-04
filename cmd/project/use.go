package project

// TODO: File needs to be generated instead of being statically defined

import (
	cloudv1 "buf.build/gen/go/gportal/gportal-cloud/protocolbuffers/go/gpcloud/api/cloud/v1"
	"fmt"
	api "github.com/G-PORTAL/gpcloud-go/pkg/gpcloud/client"
	"github.com/spf13/cobra"
	"gpcloud-cli/pkg/client"
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
	RunE: func(cobraCmd *cobra.Command, args []string) error {
		ctx := client.ExtractContext(cobraCmd)
		conn := ctx.Value("conn").(*api.Client)
		config := ctx.Value("config").(*config.SessionConfig)
		resp, err := conn.CloudClient().ListProjects(cobraCmd.Context(), &cloudv1.ListProjectsRequest{})
		if err != nil {
			return err
		}
		for _, project := range resp.Projects {
			if project.Id == id || project.Name == name {
				config.CurrentProject = &id
				if err := config.Write(); err != nil {
					return err
				}
				cobraCmd.Println("Active project is now: " + project.Name)
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
