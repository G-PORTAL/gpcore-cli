package agent

import (
	"context"
	"errors"
	"fmt"
	"github.com/G-PORTAL/gpcore-cli/pkg/api"
	"github.com/G-PORTAL/gpcore-cli/pkg/config"
	"github.com/G-PORTAL/gpcore-cli/pkg/consts"
	"github.com/G-PORTAL/gpcore-go/pkg/gpcore/client"
	"github.com/G-PORTAL/gpcore-go/pkg/gpcore/client/auth"
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
	conn   *grpc.ClientConn
	ssh    *ssh.Session
}

var session Session

var DoneChan = make(chan os.Signal, 1)
var IsRunning = false

func (s *Session) ContextWithSession(ctx context.Context) context.Context {
	ctx = context.WithValue(ctx, "config", s.config)
	ctx = context.WithValue(ctx, "ssh", s.ssh)
	ctx = context.WithValue(ctx, "conn", s.conn)
	return ctx
}

// ConnectToAPI connects to the API with the given credentials, depending
// on what credentials we have.
func ConnectToAPI(session *Session) (*grpc.ClientConn, error) {
	// Endpoint
	endpoint := config.Endpoint
	if os.Getenv("GPCORE_ENDPOINT") != "" {
		endpoint = os.Getenv("GPCORE_ENDPOINT")
	}

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
		credentials := &auth.ProviderKeycloakUserPassword{
			Username:     *session.config.Username,
			Password:     *session.config.Password,
			ClientID:     session.config.ClientID,
			ClientSecret: session.config.ClientSecret,
		}
		return api.NewGRPCConnection(
			credentials,
			client.EndpointOverrideOption(endpoint),
		)
	}

	// Otherwise, we just use the client credentials. With this login, the
	// admin endpoints will not work and result in an error.
	credentials := &auth.ProviderKeycloakClientAuth{
		ClientID:     session.config.ClientID,
		ClientSecret: session.config.ClientSecret,
	}
	return api.NewGRPCConnection(
		credentials,
		client.EndpointOverrideOption(endpoint),
	)
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

		// Open new connection
		session.conn, err = ConnectToAPI(&session)
		if err != nil {
			log.Errorf("Can not connect to GPCORE API: %v", err)
			log.Fatal("Check your config file and/or reset it with \"gpcore agent setup\"")
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
							log.Errorf("Error executing command: %v", err)
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

		go func() {
			log.Infof("Starting server on %s:%d", consts.AgentHost, consts.AgentPort)
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
	agentCommand.AddCommand(startCmd)
}
