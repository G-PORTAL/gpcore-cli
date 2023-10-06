package cmd

import (
	"buf.build/gen/go/gportal/gportal-cloud/grpc/go/gpcloud/api/admin/v1/adminv1grpc"
	adminv1 "buf.build/gen/go/gportal/gportal-cloud/protocolbuffers/go/gpcloud/api/admin/v1"
	"crypto/tls"
	"fmt"
	"github.com/G-PORTAL/gpcloud-go/pkg/gpcloud/client"
	"github.com/G-PORTAL/gpcloud-go/pkg/gpcloud/client/auth"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"gpcloud-cli/pkg/config"
)

// TODO: Stolen from the grpc lib
func getTLSOptions() *tls.Config {
	return &tls.Config{
		MinVersion: tls.VersionTLS12,
	}
}

// TODO: Not ideal.The grpc lib should provide this
func AdminClient(session *config.SessionConfig) adminv1grpc.AdminServiceClient {
	var grpcClient *grpc.ClientConn
	var options []grpc.DialOption
	// Certificate pinning
	options = append(options, grpc.WithTransportCredentials(credentials.NewTLS(getTLSOptions())))

	// User Agent
	options = append(options, grpc.WithUserAgent(fmt.Sprintf("GPCloud Golang Client [%s]", Version)))

	endpoint := client.DefaultEndpoint
	authenticationDefined := false
	var extraOptions []interface{}
	extraOptions = append(extraOptions, &auth.ProviderKeycloakClientAuth{
		ClientID:     session.ClientID,
		ClientSecret: session.ClientSecret,
	})
	extraOptions = append(extraOptions, client.EndpointOverrideOption(endpoint))

	for _, option := range extraOptions {
		if opt, ok := option.(grpc.DialOption); ok {
			options = append(options, opt)
			continue
		}
		if opt, ok := option.(client.EndpointOverrideOption); ok {
			endpoint = string(opt)
			continue
		}
		if opt, ok := option.(client.AuthProviderOption); ok && !authenticationDefined {
			options = append(options, grpc.WithPerRPCCredentials(&client.AuthOption{
				Provider: &opt,
			}))
			authenticationDefined = true
			continue
		}
	}

	grpcClient, err := grpc.Dial(endpoint, options...)
	if err != nil {
		return nil
	}
	return adminv1grpc.NewAdminServiceClient(grpcClient)
}

func LiveLogCommand(rootCmd *cobra.Command) {
	rootCmd.AddCommand(&cobra.Command{
		Use:                   "livelog",
		Short:                 "Live log stream",
		Long:                  "Live log stream",
		DisableFlagsInUseLine: true,
		Args:                  cobra.OnlyValidArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			session := cmd.Context().Value("session").(*config.SessionConfig)

			adminClient := AdminClient(session)
			res, err := adminClient.GetUser(cmd.Context(), &adminv1.GetUserRequest{})
			if err != nil {
				return err
			}
			cmd.Printf("User: %+v\n", res)

			//adminClient.SubscribeServerLogs(cmd.Context(), &adminv1.SubscribeServerLogsRequest{})
			////adminv1.SubscribeServerLogs
			//// TODO: Add filter
			//_ := adminv1.AdminLog{
			//	Id:         "",
			//	AdminUser:  nil,
			//	TargetUser: nil,
			//	Message:    "",
			//	CreatedAt:  nil,
			//	UpdatedAt:  nil,
			//}
			return nil
		},
	})
}
