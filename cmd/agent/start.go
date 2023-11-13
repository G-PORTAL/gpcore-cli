package agent

import (
	"buf.build/gen/go/gportal/gportal-cloud/grpc/go/gpcloud/api/auth/v1/authv1grpc"
	authv1 "buf.build/gen/go/gportal/gportal-cloud/protocolbuffers/go/gpcloud/api/auth/v1"
	cloudv1 "buf.build/gen/go/gportal/gportal-cloud/protocolbuffers/go/gpcloud/api/cloud/v1"
	"context"
	"errors"
	"fmt"
	"github.com/G-PORTAL/gpcloud-go/pkg/gpcloud/client"
	"github.com/G-PORTAL/gpcloud-go/pkg/gpcloud/client/auth"
	"github.com/G-PORTAL/gpcore-cli/pkg/api"
	"github.com/G-PORTAL/gpcore-cli/pkg/config"
	"github.com/G-PORTAL/gpcore-cli/pkg/consts"
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

var DoneChan = make(chan os.Signal, 1)
var IsRunning = false

func (s *Session) ContextWithSession(ctx context.Context) context.Context {
	ctx = context.WithValue(ctx, "config", s.config)
	ctx = context.WithValue(ctx, "user", s.user)
	ctx = context.WithValue(ctx, "ssh", s.ssh)
	ctx = context.WithValue(ctx, "conn", s.conn)
	return ctx
}

var startCmd = &cobra.Command{
	Use:                   "start",
	Short:                 "Start the agent",
	Long:                  "Start the agent",
	DisableFlagsInUseLine: true,
	Args:                  cobra.MatchAll(cobra.ExactArgs(0), cobra.OnlyValidArgs),
	RunE: func(cobraCmd *cobra.Command, args []string) error {
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
		if os.Getenv("GPCORE_ENDPOINT") != "" {
			endpoint = os.Getenv("GPCORE_ENDPOINT")
		}

		// Credentials
		// We need the user/pass auth to use admin endpoints
		credentials := &auth.ProviderKeycloakUserPassword{
			Username:     session.config.Username,
			Password:     session.config.Password,
			ClientID:     session.config.ClientID,
			ClientSecret: session.config.ClientSecret,
		}

		// Open new connection
		session.conn, err = api.NewGRPCConnection(
			credentials,
			client.EndpointOverrideOption(endpoint),
		)
		if err != nil {
			log.Errorf("Can not connect to GPCloud API: %v", err)
			log.Fatal("Check your config file and/or reset it with \"reset-config\"")
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
						rootCmd = New()

						rootCmd.SetArgs(s.Command())
						rootCmd.SetOut(s)
						rootCmd.SetIn(s)
						rootCmd.SetErr(s.Stderr())
						rootCmd.CompletionOptions.DisableDefaultCmd = true

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
						log.Info("User logged in")
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

			IsRunning = true

			if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
				log.Fatalf("Can not start ssh server: %v", err)
				DoneChan <- nil
				IsRunning = false
			}
		}()

		// External interrupt signal also can close the server
		signal.Notify(DoneChan, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

		// Waiting for the signal
		<-DoneChan
		IsRunning = false

		log.Info("Stopping server")
		ctx, cancel := context.WithTimeout(context.Background(), 5)
		defer cancel()
		err = server.Shutdown(ctx)
		if err != nil && !errors.Is(err, ssh.ErrServerClosed) && !errors.Is(err, context.DeadlineExceeded) {
			log.Fatalf("Can not shutdown ssh server: %v", err)
		}
		return nil
	},
}

func init() {
	AgentCommand.AddCommand(startCmd)
}
