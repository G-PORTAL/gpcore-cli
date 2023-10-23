package agent

import (
	"buf.build/gen/go/gportal/gportal-cloud/grpc/go/gpcloud/api/auth/v1/authv1grpc"
	authv1 "buf.build/gen/go/gportal/gportal-cloud/protocolbuffers/go/gpcloud/api/auth/v1"
	cloudv1 "buf.build/gen/go/gportal/gportal-cloud/protocolbuffers/go/gpcloud/api/cloud/v1"
	"context"
	"errors"
	"fmt"
	command "github.com/G-PORTAL/gpcloud-cli/cmd"
	"github.com/G-PORTAL/gpcloud-cli/pkg/config"
	"github.com/G-PORTAL/gpcloud-cli/pkg/consts"
	"github.com/G-PORTAL/gpcloud-go/pkg/gpcloud/client"
	"github.com/G-PORTAL/gpcloud-go/pkg/gpcloud/client/auth"
	"github.com/charmbracelet/log"
	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

var rootCmd *cobra.Command

type Session struct {
	config *config.SessionConfig
	user   *cloudv1.User
	conn   *grpc.ClientConn
	ssh    *ssh.Session
}

var session Session

func (s *Session) ContextWithSession(ctx context.Context) context.Context {
	ctx = context.WithValue(ctx, "config", s.config)
	ctx = context.WithValue(ctx, "user", s.user)
	ctx = context.WithValue(ctx, "ssh", s.ssh)
	ctx = context.WithValue(ctx, "conn", s.conn)
	return ctx
}

func StartServer() {
	// Initialize a new session
	sessionConfig, err := config.GetSessionConfig()
	if err != nil {
		panic(err)
	}
	session = Session{
		config: sessionConfig,
	}

	// Endpoint
	endpoint := config.Endpoint
	if os.Getenv("GPCLOUD_ENDPOINT") != "" {
		endpoint = os.Getenv("GPCLOUD_ENDPOINT")
	}

	// Credentials
	// TODO: Encrypt password or whole config file
	credentials := &auth.ProviderKeycloakClientAuth{
		ClientID:     session.config.ClientID,
		ClientSecret: session.config.ClientSecret,
	}

	// TODO: optional check if we need to use auth auth.ProviderKeycloakUserPassword{}

	// Open new connection
	session.conn, err = NewGRPCConnection(
		credentials,
		client.EndpointOverrideOption(endpoint),
	)
	if err != nil {
		log.Fatalf("Can not connect to GPCloud API: %v", err)
		panic(err)
	}

	server, err := wish.NewServer(
		wish.WithAddress(fmt.Sprintf("%s:%d", consts.AgentHost, consts.AgentPort)),
		wish.WithPublicKeyAuth(func(ctx ssh.Context, key ssh.PublicKey) bool {
			// We use a middleware handler for that
			return true
		}),
		wish.WithMiddleware(
			// Cobra I/O
			func(next ssh.Handler) ssh.Handler {
				return func(s ssh.Session) {
					log.Infof("Command: %s", s.Command())
					rootCmd = command.New()

					rootCmd.SetArgs(s.Command())
					rootCmd.SetOut(s)
					rootCmd.SetIn(s)
					rootCmd.SetErr(s.Stderr())
					rootCmd.CompletionOptions.DisableDefaultCmd = true

					//log.Printf("Verbose: %t", config.Verbose)
					//if config.Verbose {
					//	log.SetLevel(log.DebugLevel)
					//}

					session.ssh = &s
					ctx := session.ContextWithSession(context.Background())
					if err := rootCmd.ExecuteContext(ctx); err != nil {
						_ = s.Exit(1)
						return
					}

					next(s)
				}
			},
			// Auth
			func(next ssh.Handler) ssh.Handler {
				return func(s ssh.Session) {
					log.Infof("Logged in")
					publicKey, _, _, _, err := ssh.ParseAuthorizedKey([]byte(sessionConfig.PublicKey))
					if err != nil {
						log.Fatalf("Can not parse public key: %v", err)
						return
					}

					if !ssh.KeysEqual(publicKey, s.PublicKey()) {
						log.Fatalf("Invalid key")
						return
					}

					next(s)
				}
			},
		),
	)

	if err != nil {
		log.Fatalf("Can not create ssh server: %v", err)
	}

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	log.Infof("Starting server on %s:%d", consts.AgentHost, consts.AgentPort)
	go func() {
		// Set user
		authClient := authv1grpc.NewAuthServiceClient(session.conn)
		resp, err := authClient.GetUser(context.Background(), &authv1.GetUserRequest{})
		if err != nil {
			log.Fatalf("Can not get user: %v", err)
		}
		session.user = resp.GetUser()
		log.Infof("Logged in as user ID %q", resp.User.Id)

		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Can not start ssh server: %v", err)
			done <- nil
		}
	}()

	<-done
	log.Info("Stopping server")
	ctx, cancel := context.WithTimeout(context.Background(), 5)
	defer func() {
		cancel()
	}()
	if err := server.Shutdown(ctx); err != nil && !errors.Is(err, ssh.ErrServerClosed) {
		log.Fatalf("Can not shutdown ssh server: %v", err)
	}
}
