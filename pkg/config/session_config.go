package config

import (
	"os"

	"github.com/charmbracelet/log"
	"gopkg.in/yaml.v3"
)

// SessionConfig holds session related credentials and other information
// used to connect to the API and some context information. The session
// is persistent, data are sored in various places (config file, keyring)
// and will be loaded on startup.
type SessionConfig struct {
	// SSH connection
	PublicKey          string  `yaml:"public_key"`
	PrivateKey         *string `yaml:"private_key,omitempty"`
	PrivateKeyPassword *string `yaml:"private_key_password,omitempty"`

	// Normal client usage (this thing is used by every user)
	ClientID     string `yaml:"client_id"`
	ClientSecret string `yaml:"client_secret,omitempty"`

	// For admin usage (admin endpoints)
	Username *string `yaml:"username,omitempty"`
	Password *string `yaml:"password,omitempty"`

	// Impersonate another user (admin only)
	ImpersonateAccessToken *string `yaml:"impersonate_access_token,omitempty"`
	ImpersonateExpiresIn   *int    `yaml:"impersonate_expires_in,omitempty"`

	// Session related stuff
	CurrentProject *string `yaml:"current_project"`

	// Log level for debug output
	LogLevel string `yaml:"log_level"`
}

// Read reads the config from the config file and stores it in the SessionConfig
// struct. Sensitive data will be loaded from the keyring. If username and password
// is set, these fields will also filled up and can be used for admin endpoints.
func (c *SessionConfig) Read() error {
	data, err := os.ReadFile(FilePath)
	if err != nil {
		return err
	}
	if err := yaml.Unmarshal(data, c); err != nil {
		log.Errorf("Error reading config: %s", err)
		return err
	}

	// Read sensitive data from the keyring
	return GetSecretsFromKeyring(c)
}

// Write writes the config to the config file and stores sensitive data in the
// keyring. If username and password is set, these fields will also stored.
func (c *SessionConfig) Write() error {
	// Write sensitive data to the keyring
	err := StoreSecretsInKeyring(c)
	if err != nil {
		return err
	}

	cleanedSessionConfig := c

	// Remove sensitive data from struct
	cleanedSessionConfig.PrivateKeyPassword = nil
	cleanedSessionConfig.Password = nil
	cleanedSessionConfig.ClientSecret = ""
	cleanedSessionConfig.ImpersonateAccessToken = nil

	// Convert to yaml
	data, err := yaml.Marshal(cleanedSessionConfig)
	if err != nil {
		return err
	}

	// Save config to disk
	return os.WriteFile(FilePath, data, 0600)
}
