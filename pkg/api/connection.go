package api

import (
	"context"
	"crypto/tls"
	"fmt"
	"math"
	"os"

	"github.com/G-PORTAL/gpcore-cli/pkg/config"
	"github.com/G-PORTAL/gpcore-go/pkg/gpcore/client"
	"github.com/G-PORTAL/gpcore-go/pkg/gpcore/client/auth"
	"github.com/Nerzal/gocloak/v13"
	"github.com/charmbracelet/log"
	"github.com/charmbracelet/ssh"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type Session struct {
	config *config.SessionConfig
	conn   *grpc.ClientConn
	SSH    *ssh.Session
}

func (s *Session) ContextWithSession(ctx context.Context) context.Context {
	ctx = context.WithValue(ctx, "config", s.config)
	ctx = context.WithValue(ctx, "ssh", s.SSH)
	ctx = context.WithValue(ctx, "conn", s.conn)
	return ctx
}

// ActiveSession hold the currently active session. This can change when the user
// impersonate a other user. In this case, the connection will be replaced and the
// access token will also be changed to perform action on behalf of the
// impersonated user. The command `user logout` will also replace the active
// session back to the original user.
var ActiveSession Session

// NewGRPCConnection creates a new gRPC connection. We can not use the NewClient
// function from the client package, because we need the grpc.ClientConn object
// in order to open a new AdminClient. If we decide to expose the AdminClient
// in the client package, we can use the NewClient function and this function
// gets obsolete.
func NewGRPCConnection(extraOptions ...interface{}) (*grpc.ClientConn, error) {
	var options []grpc.DialOption
	// Certificate pinning
	options = append(options, grpc.WithTransportCredentials(
		credentials.NewTLS(&tls.Config{
			MinVersion: tls.VersionTLS12,
		})))

	// User Agent
	options = append(options, grpc.WithUserAgent(fmt.Sprintf("GPCORE CLI [%s]", grpc.Version)))
	endpoint := client.DefaultEndpoint

	for _, option := range extraOptions {
		if opt, ok := option.(grpc.DialOption); ok {
			options = append(options, opt)
			continue
		}
		if opt, ok := option.(client.EndpointOverrideOption); ok {
			endpoint = string(opt)
			continue
		}
		if opt, ok := option.(client.AuthProviderOption); ok {
			options = append(options, grpc.WithPerRPCCredentials(&client.AuthOption{
				Provider: &opt,
			}))
			continue
		}
		log.Printf("Unknown option type: %T", option)
	}

	clientConn, err := grpc.NewClient(endpoint, options...)
	if err != nil {
		return nil, err
	}

	return clientConn, nil
}

// ConnectToAPI connects to the API with the given credentials, depending
// on what credentials we have.
func ConnectToAPI() (*grpc.ClientConn, error) {
	// Endpoint
	endpoint := config.Endpoint
	if os.Getenv("GPCORE_ENDPOINT") != "" {
		endpoint = os.Getenv("GPCORE_ENDPOINT")
	}

	var credOptions client.AuthProviderOption

	// We have two different connection methods available, depending on the
	// type of credentials we get. For "normal" usage, we need the ClientID
	// and the ClientSecret, which can be used by every user.
	// Some endpoints need admin privileges, tho. For that, we need the
	// username and password of the user. We can not use the same connection
	// for that, so we need to reconnect with admin credentials.

	// First, we check if we have user/pass for admin login. If we have the
	// credentials, we use it for login.
	if config.HasAdminConfig() {
		log.Info("Using admin credentials")
		credOptions = &auth.ProviderKeycloakUserPassword{
			ClientID:     ActiveSession.config.ClientID,
			ClientSecret: ActiveSession.config.ClientSecret,
			Username:     *ActiveSession.config.Username,
			Password:     *ActiveSession.config.Password,
		}

		// If we currently impersonate a user, we need to add the access token
		// to the credentials, so we can access the resources of that user.
		if ActiveSession.config.ImpersonateAccessToken != nil {
			credOptions.Impersonate(&gocloak.JWT{
				AccessToken: *ActiveSession.config.ImpersonateAccessToken,
				ExpiresIn:   *ActiveSession.config.ImpersonateExpiresIn,
			})
		}

		return NewGRPCConnection(
			credOptions,
			grpc.WithDefaultCallOptions(
				grpc.MaxCallRecvMsgSize(math.MaxInt32),
				grpc.MaxCallSendMsgSize(math.MaxInt32),
			),
			client.EndpointOverrideOption(endpoint),
		)
	}

	// Otherwise, we just use the client credentials. With this login, the
	// admin endpoints will not work and result in an error.
	credOptions = &auth.ProviderKeycloakClientAuth{
		ClientID:     ActiveSession.config.ClientID,
		ClientSecret: ActiveSession.config.ClientSecret,
	}
	return NewGRPCConnection(
		credOptions,
		grpc.WithDefaultCallOptions(
			grpc.MaxCallRecvMsgSize(math.MaxInt32),
			grpc.MaxCallSendMsgSize(math.MaxInt32),
		),
		client.EndpointOverrideOption(endpoint),
	)
}

// RenewAPISession restarts the API session with the currently set config from
// the config file and the keystore. Call this function if you change the session
// config.
func RenewAPISession() *grpc.ClientConn {
	// Initialize a new session
	err := config.RefreshSessionConfig()
	if err != nil {
		log.Fatal(err)
		panic(err)
	}
	sessionConfig, err := config.GetSessionConfig()
	if err != nil {
		panic(err)
	}

	ActiveSession = Session{
		config: sessionConfig,
	}

	// Open new connection
	newConnection, err := ConnectToAPI()
	if err != nil {
		log.Errorf("Can not connect to GPCORE API: %v", err)
		log.Fatal("Check your config file and/or reset it with \"gpcore agent setup\"")
		panic(err)
	}

	ActiveSession.conn = newConnection
	log.Debugf("Renewed GPCORE API session")

	return newConnection
}
