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

var useClear bool

var useCmd = &cobra.Command{
	Use:   "use [project-name-or-id]",
	Short: "Select, show or clear the active project",
	Long: "Manage the active project used as the default --project-id for\n" +
		"project-scoped commands (e.g. \"node get\").\n\n" +
		"  - With an argument: selects the given project (by name or UUID).\n" +
		"  - Without arguments: shows the currently active project.\n" +
		"  - With --clear: clears the active project selection.",
	Example: "  gpcore project use my-project\n" +
		"  gpcore project use            # show the active project\n" +
		"  gpcore project use --clear    # clear the active project",
	DisableFlagsInUseLine: true,
	Args:                  cobra.MaximumNArgs(1),
	RunE: func(cobraCmd *cobra.Command, args []string) error {
		ctx := client.ExtractContext(cobraCmd)
		cfg := ctx.Value("config").(*config.SessionConfig)

		// Clear the active project.
		if useClear {
			if len(args) > 0 {
				return fmt.Errorf("--clear cannot be combined with a project argument")
			}
			if cfg.CurrentProject == nil {
				cobraCmd.Println("No active project was set.")
				return nil
			}
			cfg.CurrentProject = nil
			if err := cfg.Write(); err != nil {
				return err
			}
			if err := config.RefreshSessionConfig(); err != nil {
				return err
			}
			log.Info("Cleared active project")
			cobraCmd.Println("Active project cleared.")
			return nil
		}

		grpcConn := ctx.Value("conn").(*grpc.ClientConn)
		grpcClient := cloudv1grpc.NewCloudServiceClient(grpcConn)

		// No argument: show the currently active project.
		if len(args) == 0 {
			if cfg.CurrentProject == nil {
				cobraCmd.Println("No active project selected.")
				cobraCmd.Println("Select one with \"project use <project-name-or-id>\".")
				return nil
			}

			// Resolve the project to show its name and to verify it is still
			// accessible in the current (possibly impersonated) user context.
			projectID := *cfg.CurrentProject
			resp, err := grpcClient.GetProject(cobraCmd.Context(), &cloudv1.GetProjectRequest{
				Id: projectID,
			})
			if err != nil || resp.GetProject() == nil {
				// The active project is not accessible in the current context
				// (e.g. it was selected while impersonating another user). Clear
				// the stale selection so it cannot leak across contexts.
				cfg.CurrentProject = nil
				if werr := cfg.Write(); werr != nil {
					return werr
				}
				if rerr := config.RefreshSessionConfig(); rerr != nil {
					return rerr
				}
				cobraCmd.Printf("The previously active project (%s) is not accessible "+
					"in the current context and has been cleared.\n", projectID)
				cobraCmd.Println("Select one with \"project use <project-name-or-id>\".")
				return nil
			}
			cobraCmd.Printf("Active project: %s (%s)\n", resp.GetProject().GetName(), projectID)
			return nil
		}

		// Argument given: select the project.
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
	useCmd.Flags().BoolVar(&useClear, "clear", false, "Clear the active project selection")

	RootProjectCommand.AddCommand(useCmd)
}
