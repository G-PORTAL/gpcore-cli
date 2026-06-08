package agent

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/G-PORTAL/gpcore-cli/pkg/api"
	"github.com/G-PORTAL/gpcore-cli/pkg/client"
	"github.com/G-PORTAL/gpcore-cli/pkg/config"
	"github.com/G-PORTAL/gpcore-cli/pkg/consts"
	"github.com/charmbracelet/log"
	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
	"github.com/spf13/cobra"
)

var rootCmd *cobra.Command

var DoneChan = make(chan os.Signal, 1)
var IsRunning = false

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the agent",
	Long:  "Start the agent",
	Args:  cobra.MatchAll(cobra.ExactArgs(0), cobra.OnlyValidArgs),
	RunE: func(cobraCmd *cobra.Command, args []string) error {
		// Initialize a new session
		sessionConfig, err := config.GetSessionConfig()
		if err != nil {
			panic(err)
		}

		// If we have impersonated a user before the agent was stopped, we remove
		// the access token and the expiry time, so the user start with his own
		// user.
		sessionConfig.ImpersonateAccessToken = nil
		sessionConfig.ImpersonateExpiresIn = nil
		err = sessionConfig.Write()
		if err != nil {
			panic(err)
		}

		api.RenewAPISession()

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

						api.ActiveSession.SSH = &s
						ctx := api.ActiveSession.ContextWithSession(context.Background())

						if err := rootCmd.ExecuteContext(ctx); err != nil {
							log.Errorf("Error executing command on agent: %v", err)
							rootCmd.Printf("Error: %s\n", client.FormatCommandError(err))
							_ = s.Exit(1) // send cmd exit code to the client
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
