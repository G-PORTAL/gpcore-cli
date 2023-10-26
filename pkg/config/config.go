package config

import (
	"bufio"
	"errors"
	"github.com/99designs/keyring"
	"github.com/G-PORTAL/gpcloud-cli/pkg/consts"
	"github.com/G-PORTAL/gpcloud-go/pkg/gpcloud/client"
	"github.com/charmbracelet/log"
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

	ClientSecret string
	Password     string
}

func init() {
	log.Infof("Initializing config ...")
	dirname, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}

	ConfigFilePath = path.Join(dirname, ".config", consts.BinaryName, "config.yaml")
	SSHKeyFilePath = path.Join(dirname, ".config", consts.BinaryName, "id_rsa")
}

func GetSecretsFromKeyring(config *SessionConfig) error {
	ring, err := keyring.Open(keyring.Config{
		ServiceName: "gpc",
	})
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
	ring, err := keyring.Open(keyring.Config{
		ServiceName: "gpc",
	})
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

func ResetSessionConfig() error {
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
	ring, err := keyring.Open(keyring.Config{
		ServiceName: "gpc",
	})
	if err != nil {
		return err
	}
	if err = ring.Remove("client_secret"); err != nil {
		return err
	}
	if err = ring.Remove("password"); err != nil {
		return err
	}

	// Create new session config (which will ask for credentials)
	_, err = GetSessionConfig()
	if err != nil {
		return err
	}

	return nil
}

func GetSessionConfig() (*SessionConfig, error) {
	if sessionConfig != nil {
		return sessionConfig, nil
	}

	if os.Getenv("GPCLOUD_CONFIG") != "" {
		ConfigFilePath = os.Getenv("GPCLOUD_CONFIG")
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
	data, err := os.ReadFile(ConfigFilePath)
	if err != nil {
		return nil, err
	}
	if err := yaml.Unmarshal(data, sessionConfig); err != nil {
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
	// check if directory exists, if not create it recursively
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
