package admin_project

import (
	"buf.build/gen/go/gportal/gpcore/grpc/go/gpcore/api/admin/v1/adminv1grpc"
	adminv1 "buf.build/gen/go/gportal/gpcore/protocolbuffers/go/gpcore/api/admin/v1"
	cloudv1 "buf.build/gen/go/gportal/gpcore/protocolbuffers/go/gpcore/api/cloud/v1"
	"github.com/G-PORTAL/gpcore-cli/pkg/client"
	"github.com/G-PORTAL/gpcore-cli/pkg/config"
	"github.com/G-PORTAL/gpcore-cli/pkg/protobuf"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
)

// update is implemented manually (not generated) because
// admin.UpdateProjectRequest carries a nested BillingProfile message which the
// YAML generator cannot express as command flags. The billing profile is
// referenced by its UUID; the backend resolves the full profile from the ID.

var updateId string
var updateName string
var updateAvatarUrl string
var updateBillingProfileId string
var updateServerPoolIds []string

var updateCmd = &cobra.Command{
	Args:                  cobra.OnlyValidArgs,
	DisableFlagsInUseLine: true,
	Long:                  "Update project details (admin)",
	RunE: func(cobraCmd *cobra.Command, args []string) error {
		ctx := client.ExtractContext(cobraCmd)
		grpcConn := ctx.Value("conn").(*grpc.ClientConn)
		grpcClient := adminv1grpc.NewAdminServiceClient(grpcConn)

		req := &adminv1.UpdateProjectRequest{
			Id:            updateId,
			Name:          updateName,
			AvatarUrl:     updateAvatarUrl,
			ServerPoolIds: updateServerPoolIds,
		}
		if updateBillingProfileId != "" {
			req.BillingProfile = cloudv1.BillingProfile_builder{
				Id: updateBillingProfileId,
			}.Build()
		}

		resp, err := grpcClient.UpdateProject(cobraCmd.Context(), req)
		if err != nil {
			return err
		}
		if config.JSONOutput {
			jsonData, err := protobuf.MarshalIndent(resp)
			if err != nil {
				return err
			}
			cobraCmd.Println(string(jsonData))
		}
		return nil
	},
	Short:     "Update project details (admin)",
	Use:       "update",
	ValidArgs: []string{"id", "name", "avatar-url", "billing-profile-id", "server-pool-ids"},
}

func init() {
	updateCmd.Flags().StringVar(&updateId, "id", "", "Project UUID (required)")
	updateCmd.Flags().StringVar(&updateName, "name", "", "Project name (required)")
	updateCmd.Flags().StringVar(&updateAvatarUrl, "avatar-url", "", "Avatar URL")
	updateCmd.Flags().StringVar(&updateBillingProfileId, "billing-profile-id", "", "Billing profile UUID")
	updateCmd.Flags().StringSliceVar(&updateServerPoolIds, "server-pool-ids", nil, "Server pool UUIDs")

	updateCmd.MarkFlagRequired("id")
	updateCmd.MarkFlagRequired("name")

	RootAdminProjectCommand.AddCommand(updateCmd)
}
