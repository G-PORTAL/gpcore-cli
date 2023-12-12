package client

import (
	"buf.build/gen/go/gportal/gpcore/grpc/go/gpcore/api/auth/v1/authv1grpc"
	authv1 "buf.build/gen/go/gportal/gpcore/protocolbuffers/go/gpcore/api/auth/v1"
	cloudv1 "buf.build/gen/go/gportal/gpcore/protocolbuffers/go/gpcore/api/cloud/v1"
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
