package client

import (
	"github.com/G-PORTAL/gpcore-cli/pkg/config"
	"github.com/G-PORTAL/gpcore-cli/pkg/consts"
	"github.com/charmbracelet/log"
	"github.com/melbahja/goph"
	"golang.org/x/crypto/ssh"
	"io"
	"net"
	"os"
	"os/signal"
	"syscall"
)

// NewClient creates a new ssh client to execute commands. For key verification
// we use a fixed private key from the config.
func NewClient() (*goph.Client, error) {
	// We "could" use the ssh-agent, but we want to be able to use the client
	// without it. So we use the private key from the config and the public key
	// from the config to verify the host key. This is not the best solution,
	// but makes it compatible with windows.
	sessionConfig, err := config.GetSessionConfig()
	if err != nil {
		return nil, err
	}

	if sessionConfig.PrivateKey == nil {
		log.Errorf("No private key found in config, (re)run `gpcore agent setup`")
		return nil, err
	}

	auth, err := goph.RawKey(*sessionConfig.PrivateKey, *sessionConfig.PrivateKeyPassword)
	if err != nil {
		return nil, err
	}

	gophConfig := &goph.Config{
		Auth:    auth,
		User:    "cli",
		Addr:    consts.AgentHost,
		Port:    consts.AgentPort,
		Timeout: goph.DefaultTimeout,
		Callback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			// We do not need to verify the host key for now, because there is
			// no real risk. The agent is only listening on localhost and the
			// public key is stored in the config.
			return nil
		},
	}

	client, err := goph.NewConn(gophConfig)
	if err != nil {
		return nil, err
	}

	return client, nil
}

// Execute executes a command on the agent and prints the result to stdout. If
// there is an error, it will panic.
func Execute(client *goph.Client, command string) {
	session, err := client.NewSession()
	if err != nil {
		log.Errorf("Error creating session: %s", err)
		return
	}
	defer func(session *ssh.Session) {
		err := session.Close()
		if err != nil && err != io.EOF {
			log.Errorf("Error closing session: %s", err)
		}
	}(session)

	// Handle CTRL+D
	sig := make(chan os.Signal, 1)
	signal.Notify(sig)
	go func() {
		s := <-sig
		if s == syscall.SIGTERM || s == syscall.SIGINT {
			log.Printf("Closing connection ...")
			session.SendRequest("break", true, nil)
			client.Close()
			os.Exit(0)
			return
		}
	}()

	session.Stdout = os.Stdout
	session.Stderr = os.Stderr

	// Executing the command on the agent
	err = session.Run(command)

	if err != nil {
		log.Error(err)
	}
}
