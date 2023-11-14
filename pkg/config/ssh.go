package config

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"github.com/charmbracelet/log"
	"golang.org/x/crypto/ssh"
)

// SetupSSHKey generates a new SSH keypair if it does not exist yet. The private key
// is stored in the users home directory and the public key is stored in the
// config. The private key is secured with the SSHKeyPassword password.
func SetupSSHKey() {
	sessionConfig, err := GetSessionConfig()
	if err != nil {
		log.Errorf("Can not get session config: %v", err)
		return
	}

	// Generate new keypair
	log.Info("Generating new SSH keypair ...")
	privateKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		panic(err)
	}

	block := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	}

	// Generate a new random password and store this in the session config
	b := make([]byte, 32)
	_, err = rand.Read(b)
	if err != nil {
		panic(err)
	}
	randomPassword := fmt.Sprintf("%x", b)
	sessionConfig.PrivateKeyPassword = &randomPassword

	block, err = x509.EncryptPEMBlock(rand.Reader, block.Type, block.Bytes, []byte(randomPassword), x509.PEMCipherAES256)
	if err != nil {
		panic(err)
	}

	// Store private key in session config (which ends in the keyring)
	privateKeyASCII := string(pem.EncodeToMemory(block))
	sessionConfig.PrivateKey = &privateKeyASCII

	// Store public key in config
	log.Info("Storing public key in config ...")
	publicKey, err := ssh.NewPublicKey(&privateKey.PublicKey)
	if err != nil {
		panic(err)
	}
	sessionConfig.PublicKey = string(ssh.MarshalAuthorizedKey(publicKey))
	err = sessionConfig.Write()
	if err != nil {
		return
	}
}
