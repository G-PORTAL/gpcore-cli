package config

import (
	"errors"
	"github.com/G-PORTAL/gpcore-go/pkg/gpcore/client"
	"github.com/charmbracelet/log"
	"os"
)

// ConfigFilePath is the path to the config file used to store the session config. The
// default value is ~/.config/gpcore/config.yaml. This can be overwritten by setting the
// environment variable GPCORE_CONFIG or by passing the --config flag to the
// gpc command.
var FilePath = ""

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
// setting the environment variable GPCORE_ENDPOINT or by passing the --endpoint
// flag to the gpc command. This will "lock in" on the agent once set and can
// not be changes for subsequent client calls (because the connection is open).
var Endpoint = client.DefaultEndpoint

var sessionConfig *SessionConfig

func HasConfig() bool {
	if _, err := os.Stat(FilePath); err == nil {
		return true
	}
	return false
}

func HasAdminConfig() bool {
	if config, err := GetSessionConfig(); err == nil {
		return config.Username != nil && config.Password != nil
	}

	return false
}

func AskForCredentials() (string, string) {
	clientID, clientSecret := "", ""

	// Client ID
	err := errors.New("waiting for input")
	for err != nil {
		clientID, err = AskForInput("Client ID", false)
		if err != nil {
			log.Errorf("Error reading input: %s", err)
		}
	}

	// Client Secret
	err = errors.New("waiting for input")
	for err != nil {
		clientSecret, err = AskForInput("Client Secret", true)
		if err != nil {
			log.Errorf("Error reading input: %s", err)
		}
	}

	return clientID, clientSecret
}

func AskForAdminCredentials() (string, string) {
	username, password := "", ""

	// Username
	err := errors.New("waiting for input")
	for err != nil {
		username, err = AskForInput("Username", false)
		if err != nil {
			log.Errorf("Error reading input: %s", err)
		}
	}

	// Password
	err = errors.New("waiting for input")
	for err != nil {
		password, err = AskForInput("Password", true)
		if err != nil {
			log.Errorf("Error reading input: %s", err)
		}
	}

	return username, password
}

func GetSessionConfig() (*SessionConfig, error) {
	if sessionConfig != nil {
		return sessionConfig, nil
	}

	sessionConfig = &SessionConfig{}

	// No config file found?
	if !HasConfig() {
		log.Errorf("No config file found at %s", FilePath)
		log.Errorf("Create a new config file with \"gpcore agent setup\"")
		return nil, errors.New("no config file found")
	}

	// Read in config file
	err := sessionConfig.Read()
	if err != nil {
		log.Errorf("Error reading config: %s", err)
		return nil, err
	}

	return sessionConfig, nil
}

func init() {
	if os.Getenv("GPCORE_CONFIG") != "" {
		FilePath = os.Getenv("GPCORE_CONFIG")
	}
}
