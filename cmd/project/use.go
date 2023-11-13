package project

// TODO: File needs to be generated instead of being statically defined

import (
	"buf.build/gen/go/gportal/gportal-cloud/grpc/go/gpcloud/api/cloud/v1/cloudv1grpc"
	cloudv1 "buf.build/gen/go/gportal/gportal-cloud/protocolbuffers/go/gpcloud/api/cloud/v1"
	"fmt"
	"github.com/G-PORTAL/gpcore-cli/pkg/client"
	"github.com/G-PORTAL/gpcore-cli/pkg/config"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
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
		grpcConn := ctx.Value("conn").(*grpc.ClientConn)
		client := cloudv1grpc.NewCloudServiceClient(grpcConn)
		cfg := ctx.Value("config").(*config.SessionConfig)
		resp, err := client.ListProjects(cobraCmd.Context(), &cloudv1.ListProjectsRequest{})
		if err != nil {
			return err
		}
		for _, project := range resp.Projects {
			if project.Id == id || project.Name == name {
				cfg.CurrentProject = &id
				if err := cfg.Write(); err != nil {
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
