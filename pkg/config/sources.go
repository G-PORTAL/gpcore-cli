package config

import (
	"bufio"
	"errors"
	"github.com/G-PORTAL/gpcore-cli/pkg/consts"
	"github.com/G-PORTAL/gpcore-cli/pkg/secret"
	"os"
	"path"
	"strings"
)

func init() {
	dirname, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}

	FilePath = path.Join(dirname, ".config", consts.BinaryName, "config.yaml")
}

func GetSecretsFromKeyring(config *SessionConfig) error {
	ring := secret.GetKeyring()

	// Client secret
	clientSecret, err := ring.Get("client_secret")
	if err == nil {
		config.ClientSecret = clientSecret
	}

	// Private Key Password
	privateKeyPassword, err := ring.Get("private_key_password")
	if err == nil {
		config.PrivateKeyPassword = &privateKeyPassword
	}

	// Password (admin)
	password, err := ring.Get("password")
	if err == nil {
		config.Password = &password
	}

	return nil
}

func StoreSecretsInKeyring(config *SessionConfig) error {
	ring := secret.GetKeyring()
	// Client secret
	if config.ClientSecret != "" {
		err := ring.Set("client_secret", config.ClientSecret)
		if err != nil {
			return err
		}
	}

	// Password
	if config.Password != nil {
		err := ring.Set("password", *config.Password)
		if err != nil {
			return err
		}
	}

	// Private Key Password
	if config.PrivateKeyPassword != nil {
		err := ring.Set("private_key_password", *config.PrivateKeyPassword)
		if err != nil {
			return err
		}
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
