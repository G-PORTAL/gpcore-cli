package node

import (
	"buf.build/gen/go/gportal/gpcore/grpc/go/gpcore/api/cloud/v1/cloudv1grpc"
	cloudv1 "buf.build/gen/go/gportal/gpcore/protocolbuffers/go/gpcore/api/cloud/v1"
	"github.com/G-PORTAL/gpcore-cli/pkg/client"
	"github.com/G-PORTAL/gpcore-cli/pkg/config"
	"github.com/G-PORTAL/gpcore-cli/pkg/protobuf"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
)

var changeRescueModeId string
var changeRescueModeProjectId string
var changeRescueModeEnabled bool
var changeRescueModePassword string

var changeRescueModeCmd = &cobra.Command{
	Args:                  cobra.OnlyValidArgs,
	DisableFlagsInUseLine: true,
	Long:                  "Change rescue mode",
	RunE: func(cobraCmd *cobra.Command, args []string) error {
		ctx := client.ExtractContext(cobraCmd)
		grpcConn := ctx.Value("conn").(*grpc.ClientConn)
		client := cloudv1grpc.NewCloudServiceClient(grpcConn)
		resp, err := client.ChangeNodeRescueMode(cobraCmd.Context(), &cloudv1.ChangeNodeRescueModeRequest{
			Id:        changeRescueModeId,
			ProjectId: changeRescueModeProjectId,
			RescueMode: &cloudv1.RescueMode{
				Enabled:  changeRescueModeEnabled,
				Password: changeRescueModePassword,
			},
		})
		if err != nil {
			return err
		}
		respData := resp
		if config.JSONOutput {
			jsonData, err := protobuf.MarshalIndent(respData)
			if err != nil {
				return err
			}
			cobraCmd.Println(string(jsonData))
		}
		return nil
	},
	Short:     "Change rescue mode",
	Use:       "change-rescue-mode",
	ValidArgs: []string{"id", "project-id"},
}

func init() {
	changeRescueModeCmd.Flags().StringVar(&changeRescueModeId, "id", "", "Node ID (required)")
	changeRescueModeCmd.Flags().StringVar(&changeRescueModeProjectId, "project-id", "", "Project ID (required)")
	changeRescueModeCmd.Flags().BoolVar(&changeRescueModeEnabled, "enabled", false, "Enable or disable rescue mode (required)")
	changeRescueModeCmd.Flags().StringVar(&changeRescueModePassword, "password", "", "Password for rescue mode (required)")

	changeRescueModeCmd.MarkFlagRequired("id")
	changeRescueModeCmd.MarkFlagRequired("project-id")

	RootNodeCommand.AddCommand(changeRescueModeCmd)
}
