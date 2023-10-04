package client

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"golang.org/x/crypto/ssh"
	"gpcloud-cli/pkg/config"
	"log"
	"os"
	"path"
)

const SSHKeyFile = "gpc"
const SSHKeyPassword = "G^cSH@aGHz8*T74KC^!8mKj&#5iH6j%zvQwH" // Randomly generated

func Setup() {
	filePath := GetPrivateKeyFilepath()
	// Check if we already created the keypair
	if _, err := os.Stat(filePath); err == nil {
		return
	}

	// Generate keypair
	log.Printf("Generating new SSH keypair ...")
	privateKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		panic(err)
	}

	block := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	}

	// TODO: Ask for the password for more security
	block, err = x509.EncryptPEMBlock(rand.Reader, block.Type, block.Bytes, []byte(SSHKeyPassword), x509.PEMCipherAES256)
	if err != nil {
		panic(err)
	}

	// Write private key to disk
	log.Printf("Store private key to disk (%s) ...", filePath)

	err = os.WriteFile(filePath, pem.EncodeToMemory(block), 0600)
	if err != nil {
		panic(err)
	}

	// Store public key in config
	log.Printf("Storing public key in config ...")
	cfg, err := config.GetSessionConfig()
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

func GetPrivateKeyFilepath() string {
	homedir, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}
	return path.Join(homedir, ".ssh", SSHKeyFile)
}