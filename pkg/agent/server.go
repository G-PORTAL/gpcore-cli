package agent

import (
	authv1 "buf.build/gen/go/gportal/gportal-cloud/protocolbuffers/go/gpcloud/api/auth/v1"
	cloudv1 "buf.build/gen/go/gportal/gportal-cloud/protocolbuffers/go/gpcloud/api/cloud/v1"
	"context"
	"errors"
	"fmt"
	"github.com/G-PORTAL/gpcloud-go/pkg/gpcloud/client"
	"github.com/G-PORTAL/gpcloud-go/pkg/gpcloud/client/auth"
	"github.com/charmbracelet/log"
	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
	"github.com/spf13/cobra"
	command "gpcloud-cli/cmd"
	"gpcloud-cli/pkg/config"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

const (
	host = "localhost"
	port = 9001
)

var rootCmd *cobra.Command

type Session struct {
	config *config.SessionConfig
	user   *cloudv1.User
	conn   *client.Client
	ssh    *ssh.Session
}

var session Session

func (s *Session) ContextWithSession(ctx context.Context) context.Context {
	ctx = context.WithValue(ctx, "config", s.config)
	ctx = context.WithValue(ctx, "user", s.user)
	ctx = context.WithValue(ctx, "conn", s.conn)
	ctx = context.WithValue(ctx, "ssh", s.ssh)
	return ctx
}

func StartServer() {
	// Initialize a new session
	config, err := config.GetSessionConfig()
	if err != nil {
		panic(err)
	}
	session = Session{
		config: config,
	}

	// Connect to the API
	endpoint := client.DefaultEndpoint
	if os.Getenv("GPCLOUD_ENDPOINT") != "" {
		endpoint = os.Getenv("GPCLOUD_ENDPOINT")
	}

	log.Infof("Connect to GPCloud API ...")
	conn, err := client.NewClient(
		&auth.ProviderKeycloakUserPassword{
			ClientID:     session.config.ClientID, // admin-cli
			ClientSecret: session.config.ClientSecret, // ???
			Username:     session.config.Username, // aaron.fischer@g-portal.cloud
			Password:     session.config.Password, // siehe config
		},
		client.EndpointOverrideOption(endpoint),
	)
	if err != nil {
		log.Fatalf("Can not connect to GPCloud API: %v", err)
		panic(err)
	}
	session.conn = conn

	server, err := wish.NewServer(
		wish.WithAddress(fmt.Sprintf("%s:%d", host, port)),
		wish.WithPublicKeyAuth(func(ctx ssh.Context, key ssh.PublicKey) bool {
			// We use a middleware handler for that
			return true
		}),
		wish.WithMiddleware(
			// Auth
			func(next ssh.Handler) ssh.Handler {
				return func(s ssh.Session) {
					log.Infof("Logged in as: %s", s.User())
					publicKey, _, _, _, err := ssh.ParseAuthorizedKey([]byte(config.PublicKey))
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

					session.ssh = &s
					ctx := session.ContextWithSession(context.Background())

					if err := rootCmd.ExecuteContext(ctx); err != nil {
						_ = s.Exit(1)
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
	log.Infof("Starting server on %s:%d", host, port)
	go func() {
		// Set user
		resp, err := conn.AuthClient().GetUser(context.Background(), &authv1.GetUserRequest{})
		if err != nil {
			log.Fatalf("Can not get user: %v", err)
		}
		session.user = resp.GetUser()
		log.Infof("Logged in as user: %+v", session.user.Username)

		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Can not start ssh server: %v", err)
			done <- nil
		}
	}()

	<-done
	log.Info("Stopping server")
	ctx, cancal := context.WithTimeout(context.Background(), 5)
	defer func() {
		cancal()
	}()
	if err := server.Shutdown(ctx); err != nil && !errors.Is(err, ssh.ErrServerClosed) {
		log.Fatalf("Can not shutdown ssh server: %v", err)
	}
}
