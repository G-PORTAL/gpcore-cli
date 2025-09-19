package config

import (
	"bufio"
	"errors"
	"os"
	"path"
	"strings"

	"github.com/G-PORTAL/gpcore-cli/pkg/consts"
	"github.com/G-PORTAL/gpcore-cli/pkg/secret"
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

	// Impersonate access token
	impersonateAccessToken, err := ring.Get("impersonate_access_token")
	if err == nil {
		config.ImpersonateAccessToken = &impersonateAccessToken
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

	// Impersonate access token
	if config.ImpersonateAccessToken != nil {
		err := ring.Set("impersonate_access_token", *config.ImpersonateAccessToken)
		if err != nil {
			return err
		}
	} else {
		// If we do not have set the access token, we also need to remove it
		// from the keyring if still set there.
		if _, err := ring.Get("impersonate_access_token"); err == nil {
			err := ring.Remove("impersonate_access_token")
			if err != nil {
				return err
			}
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
