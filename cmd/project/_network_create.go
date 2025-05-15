package project

import (
	"buf.build/gen/go/gportal/gpcore/grpc/go/gpcore/api/admin/v1/adminv1grpc"
	adminv1 "buf.build/gen/go/gportal/gpcore/protocolbuffers/go/gpcore/api/admin/v1"
	cloudv1 "buf.build/gen/go/gportal/gpcore/protocolbuffers/go/gpcore/api/cloud/v1"
	"encoding/json"
	"github.com/G-PORTAL/gpcore-cli/pkg/client"
	"github.com/G-PORTAL/gpcore-cli/pkg/config"
	"github.com/G-PORTAL/gpcore-cli/pkg/protobuf"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
)

// This command is disabled at the moment because the ListSubnets endpoint is
// missing on gRPC (which is needed to get the subnets from the IDs).

var networkCreateProjectId string
var networkCreateName string
var networkCreateType string
var networkCreateSubnets []string
var networkCreateVlanId int32
var networkCreateDatacenter string

var networkCreateCmd = &cobra.Command{
	Args:                  cobra.OnlyValidArgs,
	DisableFlagsInUseLine: true,
	Long:                  "",
	RunE: func(cobraCmd *cobra.Command, args []string) error {
		ctx := client.ExtractContext(cobraCmd)
		grpcConn := ctx.Value("conn").(*grpc.ClientConn)
		client := adminv1grpc.NewAdminServiceClient(grpcConn)

		networkCreateSubnetStructs := make([]*cloudv1.Subnet, 0)
		// TODO: Get subnets from networkCreateSubnetUUIDList IDs
		// TODO: ListSubnets endpoint missing on gRPC

		networkCreateDatacenterStruct := &cloudv1.Datacenter{
			Id: networkCreateDatacenter,
		}

		resp, err := client.CreateProjectNetwork(cobraCmd.Context(), &adminv1.CreateProjectNetworkRequest{
			Name:       networkCreateName,
			ProjectId:  networkCreateProjectId,
			Subnets:    networkCreateSubnetStructs,
			Datacenter: networkCreateDatacenterStruct,
			Type:       protobuf.NetworkTypeToProto(networkCreateType),
			VlanId:     &networkCreateVlanId,
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
	Short:     "",
	Use:       "network-create",
	ValidArgs: []string{"project-id", "name", "type", "subnet-ids", "datacenter-id", "vlan-id"},
}

func init() {
	networkCreateCmd.Flags().StringVar(&networkCreateProjectId, "project-id", "", "Project ID (required)")
	networkCreateCmd.Flags().StringVar(&networkCreateName, "name", "", "Network name (required)")
	networkCreateCmd.Flags().StringVar(&networkCreateType, "type", "PRIVATE", "Network type (default:\"cloudv1.NETWORK_TYPE_PRIVATE\")")
	networkCreateCmd.Flags().StringSliceVar(&networkCreateSubnets, "subnet-ids", []string{}, "Subnets (required)")
	networkCreateCmd.Flags().StringVar(&networkCreateDatacenter, "datacenter-id", "", "Datacenter ID (required)")
	networkCreateCmd.Flags().Int32Var(&networkCreateVlanId, "vlan-id", int32(0), "VLAN ID")

	networkCreateCmd.MarkFlagRequired("project-id")
	networkCreateCmd.MarkFlagRequired("name")
	networkCreateCmd.MarkFlagRequired("type")
	networkCreateCmd.MarkFlagRequired("subnets")

	if config.HasAdminConfig() {
		RootProjectCommand.AddCommand(networkCreateCmd)
	}
}
