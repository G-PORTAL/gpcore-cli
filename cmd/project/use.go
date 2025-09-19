package project

import (
	"buf.build/gen/go/gportal/gpcore/grpc/go/gpcore/api/cloud/v1/cloudv1grpc"
	cloudv1 "buf.build/gen/go/gportal/gpcore/protocolbuffers/go/gpcore/api/cloud/v1"
	"github.com/G-PORTAL/gpcore-cli/pkg/client"
	"github.com/G-PORTAL/gpcore-cli/pkg/config"
	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
)

var useCmd = &cobra.Command{
	Use:                   "use",
	Short:                 "Selects a project to use",
	Long:                  "Selects a project to use",
	DisableFlagsInUseLine: true,
	Args:                  cobra.MatchAll(cobra.ExactArgs(1)),
	RunE: func(cobraCmd *cobra.Command, args []string) error {
		ctx := client.ExtractContext(cobraCmd)
		grpcConn := ctx.Value("conn").(*grpc.ClientConn)
		grpcClient := cloudv1grpc.NewCloudServiceClient(grpcConn)
		cfg := ctx.Value("config").(*config.SessionConfig)

		var newProject *cloudv1.Project

		resp, err := grpcClient.ListProjects(cobraCmd.Context(), &cloudv1.ListProjectsRequest{})
		if err != nil {
			return err
		}

		for _, project := range resp.Projects {
			if (project.Name == args[0]) || (project.Id == args[0]) {
				newProject = project
				break
			}
		}

		// If there is no project found in the list of the current user (or the
		// user currently impersonating), we need to raise an error.
		if newProject == nil {
			cobraCmd.Println("Project not found or not accessible.")
			cobraCmd.Println("If you are an admin user, try impersonate first.")
			return nil
		}

		// Set the new project as the current project in the config and save
		// the config. This will be used for all subsequent commands.
		log.Info("Selecting project: " + args[0])
		cfg.CurrentProject = &newProject.Id
		if err := cfg.Write(); err != nil {
			return err
		}
		if err := config.RefreshSessionConfig(); err != nil {
			return err
		}
		log.Info("Active project is now: " + newProject.Name)
		cobraCmd.Println("Active project is now: " + newProject.Name)

		return nil
	},
}

func init() {
	RootProjectCommand.AddCommand(useCmd)
}
