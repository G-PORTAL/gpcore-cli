package client

import (
	"context"

	"buf.build/gen/go/gportal/gpcore/grpc/go/gpcore/api/auth/v1/authv1grpc"
	authv1 "buf.build/gen/go/gportal/gpcore/protocolbuffers/go/gpcore/api/auth/v1"
	cloudv1 "buf.build/gen/go/gportal/gpcore/protocolbuffers/go/gpcore/api/cloud/v1"
	"github.com/G-PORTAL/gpcore-cli/pkg/config"
	"github.com/charmbracelet/log"
	"google.golang.org/grpc"
)

// GetUserFromContext return the gRPC user, which will be extracted from the
// context.
func GetUserFromContext(ctx context.Context) *cloudv1.User {
	conn := ctx.Value("conn").(*grpc.ClientConn)
	return GetUser(conn)
}

// GetUser return the gRPC user from a connection.
func GetUser(conn *grpc.ClientConn) *cloudv1.User {
	authClient := authv1grpc.NewAuthServiceClient(conn)
	resp, err := authClient.GetUser(context.Background(), &authv1.GetUserRequest{})
	if err != nil {
		log.Fatalf("Can not get user: %v", err)
		return nil
	}

	return resp.GetUser()
}

// IsImpersonated return true if the current acting user is an impersonated
// user and not the real user.
func IsImpersonated() (bool, error) {
	cfg, err := config.GetSessionConfig()
	if err != nil {
		return false, err
	}
	return cfg.ImpersonateAccessToken != nil, nil
}
