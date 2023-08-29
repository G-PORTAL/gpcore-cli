package config

import (
	"bufio"
	"gopkg.in/yaml.v3"
	"os"
	"strings"
)

var configPath = ""
var sessionConfig *SessionConfig

func init() {
	dirname, _ := os.UserHomeDir()
	configPath = dirname + "/.gpc.yaml"
}

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
		configPath = os.Getenv("GPCLOUD_CONFIG")
	}
	if _, err := os.Stat(configPath); err != nil {
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
	data, err := os.ReadFile(configPath)
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
	err = os.WriteFile(configPath, data, 0600)
	if err != nil {
		return err
	}
	return nil
}
