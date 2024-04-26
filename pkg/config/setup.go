package config

import (
	"errors"
	"github.com/G-PORTAL/gpcore-cli/pkg/secret"
	"github.com/charmbracelet/log"
	"gopkg.in/op/go-logging.v1"
	"os"
	"path"
)

func SetupConfig() error {
	log.Info("Reset config (new config file and keyring) ...")
	sessionConfig = &SessionConfig{}
	err := CleanupConfig()
	if err != nil {
		return err
	}

	// Check if directory exists, if not, create it recursively
	directory := path.Dir(ConfigFilePath)
	if _, err := os.Stat(directory); os.IsNotExist(err) {
		log.Debugf("Creating directory %s", directory)
		if err := os.MkdirAll(directory, 0700); err != nil {
			return err
		}
	}

	// Ask for credentials (clientID, clientSecret)
	clientID, clientSecret := AskForCredentials()
	sessionConfig.ClientID = clientID
	sessionConfig.ClientSecret = clientSecret

	// Create new ssh keypair
	log.Infof("Setup SSH keypair ...")
	SetupSSHKey()

	// Set debug level to warning
	sessionConfig.LogLevel = "INFO"
	logging.SetLevel(logging.INFO, "")

	// Save the config
	return sessionConfig.Write()
}

func CleanupConfig() error {
	// Remove config file
	if _, err := os.Stat(ConfigFilePath); err == nil {
		err = os.Remove(ConfigFilePath)
		if err != nil {
			return err
		}
	}

	// Remove secrets from keyring
	ring := secret.GetKeyring()

	if err := ring.Remove("client_secret"); err != nil && !errors.Is(err, secret.ErrKeyNotFound) {
		return err
	}
	if err := ring.Remove("private_key"); err != nil && !errors.Is(err, secret.ErrKeyNotFound) {
		return err
	}
	if err := ring.Remove("private_key_password"); err != nil && !errors.Is(err, secret.ErrKeyNotFound) {
		return err
	}

	return nil
}

func SetupAdminConfig() error {
	log.Info("Reset admin config (username/password) ...")
	err := CleanupAdminConfig()
	if err != nil {
		return err
	}

	username, password := AskForAdminCredentials()
	sessionConfig.Username = &username
	sessionConfig.Password = &password

	return sessionConfig.Write()
}

func CleanupAdminConfig() error {
	// Remove secrets from keyring
	ring := secret.GetKeyring()
	if err := ring.Remove("password"); err != nil && !errors.Is(err, secret.ErrKeyNotFound) {
		return err
	}

	return nil
}
