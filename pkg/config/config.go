package config

import (
	"bufio"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"github.com/99designs/keyring"
	"github.com/G-PORTAL/gpcloud-go/pkg/gpcloud/client"
	"github.com/G-PORTAL/gpcore-cli/pkg/consts"
	"github.com/G-PORTAL/gpcore-cli/pkg/secret"
	"github.com/charmbracelet/log"
	"golang.org/x/crypto/ssh"
	"gopkg.in/yaml.v3"
	"os"
	"path"
	"strings"
)

// ConfigFilePath is the path to the config file used to store the session config. The
// default value is ~/.config/gpcloud/config.yaml. This can be overwritten by setting the
// environment variable GPCLOUD_CONFIG or by passing the --config flag to the
// gpc command.
var ConfigFilePath = ""

var SSHKeyFilePath = ""

// JSONOutput is a global flag that can be used to output the result of a command
// as JSON. This can be enabled by passing the --json flag to the gpc command.
var JSONOutput = false

// CSVOutput is a global flag that can be used to output the result of a command
// as CSV. This can be enabled by passing the --csv flag to the gpc command.
var CSVOutput = false

// Verbose enable verbose mode. This can be enabled by passing the --verbose flag
// to the gpc command.
var Verbose = false

// Endpoint is the API endpoint used by the client. This can be overwritten by
// setting the environment variable GPCLOUD_ENDPOINT or by passing the --endpoint
// flag to the gpc command. This will "lock in" on the agent once set and can
// not be changes for subsequent client calls (because the connection is open).
var Endpoint = client.DefaultEndpoint

var sessionConfig *SessionConfig

type SessionConfig struct {
	ClientID       string  `yaml:"client_id"`
	CurrentProject *string `yaml:"current_project"`
	PublicKey      string  `yaml:"public_key"`
	Username       string  `yaml:"username"`

	ClientSecret string `yaml:"client_secret,omitempty"`
	Password     string `yaml:"password,omitempty"`
}

func init() {
	dirname, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}

	ConfigFilePath = path.Join(dirname, ".config", consts.BinaryName, "config.yaml")
	SSHKeyFilePath = path.Join(dirname, ".config", consts.BinaryName, "id_rsa")
}

func ReadConfigFile(config *SessionConfig) error {
	data, err := os.ReadFile(ConfigFilePath)
	if err != nil {
		return err
	}
	if err := yaml.Unmarshal(data, config); err != nil {
		log.Errorf("Error reading config: %s", err)
		return err
	}

	return nil
}

func GetSecretsFromKeyring(config *SessionConfig) error {
	ring, err := secret.GetKeyring()
	if err != nil {
		return err
	}

	// Client secret
	clientSecret, err := ring.Get("client_secret")
	if err != nil {
		return err
	}
	config.ClientSecret = string(clientSecret.Data)

	// Password
	password, err := ring.Get("password")
	if err != nil {
		return err
	}
	config.Password = string(password.Data)

	return nil
}

func StoreSecretsInKeyring(config *SessionConfig) error {
	ring, err := secret.GetKeyring()
	if err != nil {
		return err
	}

	// Client secret
	err = ring.Set(keyring.Item{
		Key:  "client_secret",
		Data: []byte(config.ClientSecret),
	})
	if err != nil {
		return err
	}

	// Password
	err = ring.Set(keyring.Item{
		Key:  "password",
		Data: []byte(config.Password),
	})
	if err != nil {
		return err
	}

	return nil
}

func AskForInput(name string, isSecret bool) (string, error) {
	reader := bufio.NewReader(os.Stdin)

	passNotice := ""
	if isSecret {
		passNotice = " (will be stored in keyring)"
	}
	println("Please enter your " + name + passNotice + ":")
	input, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}

	input = strings.TrimSpace(input)

	if input == "" {
		return "", errors.New("input is empty")
	}

	return input, nil
}

func HasConfig() bool {
	if _, err := os.Stat(ConfigFilePath); err == nil {
		return true
	}
	return false
}

func ResetConfig() error {
	log.Info("Reset config (new config file and keyring) ...")
	sessionConfig = nil

	// Remove config file
	if _, err := os.Stat(ConfigFilePath); err == nil {
		err = os.Remove(ConfigFilePath)
		if err != nil {
			return err
		}
	}

	// Remove secrets from keyring
	ring, err := secret.GetKeyring()
	if err != nil {
		return err
	}

	if err = ring.Remove("client_secret"); err != nil && !errors.Is(err, keyring.ErrKeyNotFound) {
		return err
	}
	if err = ring.Remove("password"); err != nil && !errors.Is(err, keyring.ErrKeyNotFound) {
		return err
	}

	// Create new session config (which will ask for credentials)
	_, err = GetSessionConfig()
	if err != nil {
		return err
	}

	// Create new ssh keypair
	log.Infof("Delete old SSH keypair ...")
	if _, err := os.Stat(SSHKeyFilePath); err == nil {
		err = os.Remove(SSHKeyFilePath)
		if err != nil {
			return err
		}
	}
	log.Infof("Setup SSH keypair ...")
	SetupSSHKey()

	// Set new configuration
	err = ReadConfigFile(sessionConfig)
	if err != nil {
		return err
	}

	err = GetSecretsFromKeyring(sessionConfig)
	if err != nil {
		return err
	}

	return nil
}

func GetSessionConfig() (*SessionConfig, error) {
	if sessionConfig != nil {
		return sessionConfig, nil
	}

	if os.Getenv("GPCORE_CONFIG") != "" {
		ConfigFilePath = os.Getenv("GPCORE_CONFIG")
	}

	sessionConfig = &SessionConfig{}

	// No config file found?
	if _, err := os.Stat(ConfigFilePath); err != nil {
		log.Infof("No config file found at %s, creating a new one ...", ConfigFilePath)

		// Client ID
		err := errors.New("waiting for input")
		for err != nil {
			sessionConfig.ClientID, err = AskForInput("Client ID", false)
			if err != nil {
				log.Errorf("Error reading input: %s", err)
			}
		}

		// Client Secret
		err = errors.New("waiting for input")
		for err != nil {
			sessionConfig.ClientSecret, err = AskForInput("Client Secret", true)
			if err != nil {
				log.Errorf("Error reading input: %s", err)
			}
		}

		// Username
		err = errors.New("waiting for input")
		for err != nil {
			sessionConfig.Username, err = AskForInput("Username", false)
			if err != nil {
				log.Errorf("Error reading input: %s", err)
			}
		}

		// Password
		err = errors.New("waiting for input")
		for err != nil {
			sessionConfig.Password, err = AskForInput("Password", true)
			if err != nil {
				log.Errorf("Error reading input: %s", err)
			}
		}

		// Store secrets in keyring
		if err := StoreSecretsInKeyring(sessionConfig); err != nil {
			log.Errorf("Error storing secrets in keyring: %s", err)
			return nil, err
		}

		// Write config to disk
		if err := sessionConfig.Write(); err != nil {
			log.Errorf("Error writing config to disk: %s", err)
			return nil, err
		}
	}

	// Read in config file
	err := ReadConfigFile(sessionConfig)
	if err != nil {
		log.Errorf("Error reading config: %s", err)
		return nil, err
	}

	// Get secrets from keyring
	if err := GetSecretsFromKeyring(sessionConfig); err != nil {
		log.Errorf("Error reading secrets from keyring: %s", err)
		return nil, err
	}

	return sessionConfig, nil
}

func (c *SessionConfig) CreateConfigDirectory() error {
	// Check if directory exists, if not, create it recursively
	directory := path.Dir(ConfigFilePath)
	if _, err := os.Stat(directory); os.IsNotExist(err) {
		log.Debugf("Creating directory %s", directory)
		if err := os.MkdirAll(directory, 0700); err != nil {
			return err
		}
	}

	return nil
}

func (c *SessionConfig) Write() error {
	err := c.CreateConfigDirectory()
	if err != nil {
		return err
	}

	// Remove sensitive information before storing to disk
	c.ClientSecret = ""
	c.Password = ""

	data, err := yaml.Marshal(c)
	if err != nil {
		return err
	}

	err = os.WriteFile(ConfigFilePath, data, 0600)
	if err != nil {
		return err
	}

	return nil
}

// SetupSSHKey generates a new SSH keypair if it does not exist yet. The private key
// is stored in the users home directory and the public key is stored in the
// config. The private key is secured with the SSHKeyPassword password.
func SetupSSHKey() {
	cfg, err := GetSessionConfig()
	if err != nil {
		log.Errorf("Can not get session config: %v", err)
		return
	}

	err = cfg.CreateConfigDirectory()
	if err != nil {
		log.Errorf("Can not create config directory: %v", err)
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

	// We use the user password to also encrypt the private key
	sessionConfig, err := GetSessionConfig()
	if err != nil {
		panic(err)
	}

	block, err = x509.EncryptPEMBlock(rand.Reader, block.Type, block.Bytes, []byte(sessionConfig.Password), x509.PEMCipherAES256)
	if err != nil {
		panic(err)
	}

	// Write private key to disk
	log.Infof("Store private key to disk (%s) ...", SSHKeyFilePath)

	err = os.WriteFile(SSHKeyFilePath, pem.EncodeToMemory(block), 0600)
	if err != nil {
		panic(err)
	}

	// Store public key in config
	log.Info("Storing public key in config ...")
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
