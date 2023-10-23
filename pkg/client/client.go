package client

import (
	"github.com/G-PORTAL/gpcloud-cli/pkg/consts"
	"github.com/charmbracelet/log"
	"github.com/melbahja/goph"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/terminal"
	"io"
	"net"
	"os"
	"os/signal"
	"syscall"
)

// NewClient creates a new ssh client to execute commands. For key verification
// we use a fixed private key from the config.
func NewClient() (*goph.Client, error) {
	keyPath, err := GetPrivateKeyFilepath()
	if _, err := os.Stat(keyPath); err != nil {
		log.Errorf("SSH keypair not found. Please run 'gpc agent start' first.")
		return nil, err
	}

	// We "could" use the ssh-agent, but we want to be able to use the client
	// without it. So we use the private key from the config and the public key
	// from the config to verify the host key. This is not the best solution,
	// but makes it compatible with windows.
	auth, err := goph.Key(keyPath, SSHKeyPassword)

	config := &goph.Config{
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

	client, err := goph.NewConn(config)
	if err != nil {
		return nil, err
	}

	return client, nil
}

func AskPassword() string {
	pass, err := terminal.ReadPassword(0)
	if err != nil {
		panic(err)
	}

	return string(pass)
}

// Execute executes a command on the agent and prints the result to stdout. If
// there is an error, it will panic.
func Execute(client *goph.Client, command string) {
	session, err := client.NewSession()
	if err != nil {
		log.Errorf("Error creating session: %s", err)
		panic(err)
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
	err = session.Run(command)
	if err != nil {
		log.Errorf("Error executing command: %s", err)
		panic(err)
	}
}
