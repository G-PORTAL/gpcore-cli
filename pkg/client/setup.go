package client

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"github.com/G-PORTAL/gpcloud-cli/pkg/config"
	"github.com/charmbracelet/log"
	"golang.org/x/crypto/ssh"
	"os"
)

const SSHKeyFile = "gpcloud"

// SSHKeyPassword is a random generated password for the private key. This is
// shared between client and agent. This is not the ultimate security solution,
// but to ask for a password on every client connection is not very ergonomic.
// We could use the ssh-agent, but we want to be able to use the client without
// it to be compatible with windows.

// Setup generates a new SSH keypair if it does not exist yet. The private key
// is stored in the users home directory and the public key is stored in the
// config. The private key is secured with the SSHKeyPassword password.
func Setup() {
	cfg, err := config.GetSessionConfig()
	if err != nil {
		log.Errorf("Can not get session config: %v", err)
		return
	}

	err = cfg.CreateConfigDirectory()
	if err != nil {
		log.Errorf("Can not create config directory: %v", err)
		return
	}

	// Check if we already created the keypair
	if _, err := os.Stat(config.SSHKeyFilePath); err == nil {
		if err == nil {
			log.Infof("SSH keypair already exists (%s)", config.SSHKeyFilePath)
			return
		}

		// Generate keypair
		log.Debug("Generating new SSH keypair ...")
		privateKey, err := rsa.GenerateKey(rand.Reader, 4096)
		if err != nil {
			panic(err)
		}

		block := &pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
		}

		// We use the user password to also encrypt the private key
		sessionConfig, err := config.GetSessionConfig()
		if err != nil {
			panic(err)
		}

		block, err = x509.EncryptPEMBlock(rand.Reader, block.Type, block.Bytes, []byte(sessionConfig.Password), x509.PEMCipherAES256)
		if err != nil {
			panic(err)
		}

		// Write private key to disk
		log.Debug("Store private key to disk (%s) ...", config.SSHKeyFilePath)

		err = os.WriteFile(config.SSHKeyFilePath, pem.EncodeToMemory(block), 0600)
		if err != nil {
			panic(err)
		}

		// Store public key in config
		log.Debug("Storing public key in config ...")
		publicKey, err := ssh.NewPublicKey(&privateKey.PublicKey)
		if err != nil {
			panic(err)
		}
		cfg.PublicKey = string(ssh.MarshalAuthorizedKey(publicKey))
		err = cfg.Write()
		if err != nil {
			return
		}
	}
}
