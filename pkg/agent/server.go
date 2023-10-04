package agent

import (
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
	"gopkg.in/op/go-logging.v1"
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

// TODO: Make it possible to use a different session for different connections
var session Session

func (s *Session) ContextWithSession(ctx context.Context) context.Context {
	ctx = context.WithValue(ctx, "config", s.config)
	ctx = context.WithValue(ctx, "user", s.user)
	ctx = context.WithValue(ctx, "conn", s.conn)
	ctx = context.WithValue(ctx, "ssh", s.ssh)
	return ctx
}

func StartServer() {
	// Initialize logger
	var format = logging.MustStringFormatter(`%{color}%{time:15:04:05} %{shortfunc} [%{level:.4s}]%{color:reset} %{message}`)
	var backend = logging.AddModuleLevel(logging.NewBackendFormatter(logging.NewLogBackend(os.Stderr, "", 0), format))
	backend.SetLevel(logging.ERROR, "")
	logging.SetBackend(backend)

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

	log.Infof("Connect to GPClout API ...")
	conn, err := client.NewClient(
		&auth.ProviderKeycloakClientAuth{
			ClientID:     session.config.ClientID,     // Set your Client ID
			ClientSecret: session.config.ClientSecret, // Set your Client Secret
		},
		client.EndpointOverrideOption(endpoint),
	)
	if err != nil {
		panic(err)
	}
	session.conn = conn

	// Set user
	//resp, err := conn.AuthClient().GetUser(rootCmd.Context(), &authv1.GetUserRequest{})
	//if err != nil {
	//	panic(err)
	//}
	//session.user = resp.GetUser()

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
