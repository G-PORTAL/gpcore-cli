package project

import (
	"buf.build/gen/go/gportal/gpcore/grpc/go/gpcore/api/cloud/v1/cloudv1grpc"
	cloudv1 "buf.build/gen/go/gportal/gpcore/protocolbuffers/go/gpcore/api/cloud/v1"
	"fmt"
	"github.com/G-PORTAL/gpcore-cli/pkg/client"
	"github.com/G-PORTAL/gpcore-cli/pkg/config"
	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
)

var useId string
var useName string
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
			if (useName != "" && project.Name == useName) || (useId == "" && project.Id == useId) {
				cfg.CurrentProject = &project.Id
				if err := cfg.Write(); err != nil {
					return err
				}
				log.Info("Active project is now: " + project.Name)
				cobraCmd.Println("Active project is now: " + project.Name)
				return nil
			}
		}
		return fmt.Errorf("project not found")
	},
}

func init() {
	RootProjectCommand.AddCommand(useCmd)
	useCmd.PersistentFlags().StringVarP(&useId, "id", "", "", "Specify ID of Project to use")
	useCmd.PersistentFlags().StringVarP(&useName, "name", "", "", "Specify name of Project to use")
}
