package client

import (
	"buf.build/gen/go/gportal/gportal-cloud/grpc/go/gpcloud/api/auth/v1/authv1grpc"
	authv1 "buf.build/gen/go/gportal/gportal-cloud/protocolbuffers/go/gpcloud/api/auth/v1"
	cloudv1 "buf.build/gen/go/gportal/gportal-cloud/protocolbuffers/go/gpcloud/api/cloud/v1"
	"context"
	"github.com/charmbracelet/log"
	"google.golang.org/grpc"
)

func GetUser(ctx context.Context) *cloudv1.User {
	conn := ctx.Value("conn").(*grpc.ClientConn)

	authClient := authv1grpc.NewAuthServiceClient(conn)
	resp, err := authClient.GetUser(context.Background(), &authv1.GetUserRequest{})
	if err != nil {
		log.Fatalf("Can not get user: %v", err)
		return nil
	}

	return resp.GetUser()

}
