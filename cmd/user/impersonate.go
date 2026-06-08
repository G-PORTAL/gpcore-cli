package user

import (
	"buf.build/gen/go/gportal/gpcore/grpc/go/gpcore/api/admin/v1/adminv1grpc"
	adminv1 "buf.build/gen/go/gportal/gpcore/protocolbuffers/go/gpcore/api/admin/v1"
	"github.com/G-PORTAL/gpcore-cli/pkg/api"
	"github.com/G-PORTAL/gpcore-cli/pkg/client"
	"github.com/G-PORTAL/gpcore-cli/pkg/config"
	"github.com/G-PORTAL/gpcore-cli/pkg/protobuf"
	"github.com/charmbracelet/log"
	"github.com/charmbracelet/ssh"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
)

var userID string

var impersonateCmd = &cobra.Command{
	Use:       "impersonate",
	Short:     "Impersonate a user",
	Long:      "Impersonate a user, to access the resources from that user",
	ValidArgs: []string{"id"},
	Args:      cobra.OnlyValidArgs,
	RunE: func(cobraCmd *cobra.Command, args []string) error {
		ctx := client.ExtractContext(cobraCmd)
		grpcConn := ctx.Value("conn").(*grpc.ClientConn)
		grpcClient := adminv1grpc.NewAdminServiceClient(grpcConn)

		resp, err := grpcClient.ImpersonateUser(cobraCmd.Context(), &adminv1.ImpersonateUserRequest{
			UserId: userID,
		})
		if err != nil {
			return err
		}

		// Store new access token in config
		sessionConfig, err := config.GetSessionConfig()
		if err != nil {
			return err
		}

		accessToken := resp.GetToken().GetAccessToken()
		sessionConfig.ImpersonateAccessToken = &accessToken

		expiresIn := int(resp.GetToken().GetExpiresAt().GetSeconds())
		sessionConfig.ImpersonateExpiresIn = &expiresIn

		// The active project belongs to the previous user context and is not
		// valid for the impersonated user. Clear it so project-scoped commands
		// don't operate on a project that does not belong to the new context.
		sessionConfig.CurrentProject = nil

		err = sessionConfig.Write()
		if err != nil {
			return err
		}

		err = config.RefreshSessionConfig()
		if err != nil {
			return err
		}

		newConnection := api.RenewAPISession()

		user := client.GetUser(newConnection)
		log.Infof("Impersonating (id: %s) %s", user.GetId(), user.GetUsername())

		if config.JSONOutput {
			jsonData, err := protobuf.MarshalIndent(resp)
			if err != nil {
				return err
			}
			cobraCmd.Println(string(jsonData))
		} else {
			tbl := table.NewWriter()
			tbl.SetStyle(table.StyleRounded)
			sshSession := ctx.Value("ssh").(*ssh.Session)
			tbl.SetOutputMirror(*sshSession)
			tbl.AppendRow([]interface{}{"ID", user.GetId()})
			tbl.AppendRow([]interface{}{"Email: ", user.GetEmail()})

			cobraCmd.Printf("Impersonating user %v\n", user.GetUsername())
			cobraCmd.Println("All actions are now performed on behalf of the following user.")
			cobraCmd.Println("Use 'user logout' to stop impersonating the user.")
			tbl.Render()
		}

		return nil
	},
}

func init() {
	impersonateCmd.Flags().StringVar(&userID, "id", "", "User ID to impersonate (required)")
	err := impersonateCmd.MarkFlagRequired("id")
	if err != nil {
		return
	}

	if config.HasAdminConfig() {
		RootUserCommand.AddCommand(impersonateCmd)
	}
}
