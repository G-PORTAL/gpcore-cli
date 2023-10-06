package client

import (
	"github.com/charmbracelet/log"
	"github.com/melbahja/goph"
	"golang.org/x/crypto/ssh"
	"net"
	"os"
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
		Addr:    "localhost",
		Port:    9001,
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

// Execute executes a command on the agent and prints the result to stdout. If
// there is an error, it will panic.
func Execute(client *goph.Client, command string) {
	res, err := client.Run(command)
	if err != nil {
		panic(err)
	}

	_, err = os.Stdout.Write(res)
	if err != nil {
		return
	}
}
