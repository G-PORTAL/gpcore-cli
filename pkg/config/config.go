package config

import (
	"bufio"
	"gopkg.in/yaml.v3"
	"os"
	"strings"
)

// Path is the path to the config file used to store the session config. The
// default value is ~/.gpc.yaml. This can be overwritten by setting the
// environment variable GPCLOUD_CONFIG or by passing the --config flag to the
// gpc command.
var Path = ""

// JSONOutput is a global flag that can be used to output the result of a command
// as JSON. This can be enabled by passing the --json flag to the gpc command.
var JSONOutput = false

// CSVOutput is a global flag that can be used to output the result of a command
// as CSV. This can be enabled by passing the --csv flag to the gpc command.
var CSVOutput = false
var sessionConfig *SessionConfig

type SessionConfig struct {
	ClientID       string  `yaml:"client_id"`
	ClientSecret   string  `yaml:"client_secret"`
	CurrentProject *string `yaml:"current_project"`
}

func GetSessionConfig() (*SessionConfig, error) {
	if sessionConfig != nil {
		return sessionConfig, nil
	}
	if os.Getenv("GPCLOUD_CONFIG") != "" {
		Path = os.Getenv("GPCLOUD_CONFIG")
	}
	if _, err := os.Stat(Path); err != nil {
		reader := bufio.NewReader(os.Stdin)
		println("Please enter your Client ID:")
		clientID, _ := reader.ReadString('\n')
		println("Please enter your Client Secret:")
		clientSecret, _ := reader.ReadString('\n')
		sessionConfig = &SessionConfig{
			ClientID:     strings.TrimSpace(clientID),
			ClientSecret: strings.TrimSpace(clientSecret),
		}
		if err := sessionConfig.Write(); err != nil {
			return nil, err
		}
		return sessionConfig, nil
	}
	data, err := os.ReadFile(Path)
	if err != nil {
		return nil, err
	}
	sessionConfig = &SessionConfig{}
	if err := yaml.Unmarshal(data, sessionConfig); err != nil {
		return nil, err
	}
	return sessionConfig, nil
}

func (c *SessionConfig) Write() error {
	data, err := yaml.Marshal(c)
	if err != nil {
		return err
	}
	err = os.WriteFile(Path, data, 0600)
	if err != nil {
		return err
	}
	return nil
}
