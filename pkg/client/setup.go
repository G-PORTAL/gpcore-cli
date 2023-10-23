package client

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"github.com/G-PORTAL/gpcloud-cli/pkg/config"
	"github.com/charmbracelet/log"
	"golang.org/x/crypto/ssh"
	"os"
	"path"
)

const SSHKeyFile = "gpcloud"

// SSHKeyPassword is a random generated password for the private key. This is
// shared between client and agent. This is not the ultimate security solution,
// but to ask for a password on every client connection is not very ergonomic.
// We could use the ssh-agent, but we want to be able to use the client without
// it to be compatible with windows.

// TODO: use https://github.com/99designs/keyring instead!
const SSHKeyPassword = "G^cSH@aGHz8*T74KC^!8mKj&#5iH6j%zvQwH" // Randomly generated

// Setup generates a new SSH keypair if it does not exist yet. The private key
// is stored in the users home directory and the public key is stored in the
// config. The private key is secured with the SSHKeyPassword password.
func Setup() {
	cfg, err := config.GetSessionConfig()
	if err != nil {
		log.Errorf("Can not get session config: %v", err)
		return
	}

	configDir, err := cfg.GetConfigDirectory()
	if err != nil {
		log.Errorf("Can not get config directory: %v", err)
		return
	}

	// Check if we already created the keypair
	filePath, err := GetPrivateKeyFilepath()
	if err == nil {
		log.Errorf("SSH keypair already exists (%s)", filePath)
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

	block, err = x509.EncryptPEMBlock(rand.Reader, block.Type, block.Bytes, []byte(SSHKeyPassword), x509.PEMCipherAES256)
	if err != nil {
		panic(err)
	}

	// Write private key to disk
	log.Debug("Store private key to disk (%s) ...", filePath)

	err = os.WriteFile(path.Join(configDir, "id_rsa"), pem.EncodeToMemory(block), 0600)
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

// GetPrivateKeyFilepath returns the path to the private key file. if it does
// not exist or the user home dir can not be found, an empty string is returned.
func GetPrivateKeyFilepath() (string, error) {
	homedir, err := os.UserHomeDir()
	if err != nil {
		log.Debugf("Can not find user home dir: %v", err)
		return "", err
	}

	filePath := path.Join(homedir, ".ssh", SSHKeyFile)

	if _, err := os.Stat(filePath); err != nil {
		log.Debugf("SSH keypair not found (%s)", filePath)
		return filePath, errors.New("SSH keypair not found")
	}

	return filePath, nil
}
