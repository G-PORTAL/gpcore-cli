package config

import (
	"bufio"
	"errors"
	"github.com/99designs/keyring"
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

	ConfigFilePath = path.Join(dirname, ".config", consts.BinaryName, "config.yaml")
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
	pw := string(password.Data)
	config.Password = &pw

	// Private Key
	privateKey, err := ring.Get("private_key")
	if err == nil {
		pk := string(privateKey.Data)
		config.PrivateKey = &pk
	}

	// Password
	privateKeyPassword, err := ring.Get("private_key_password")
	if err == nil {
		pkpw := string(privateKeyPassword.Data)
		config.PrivateKeyPassword = &pkpw
	}

	return nil
}

func StoreSecretsInKeyring(config *SessionConfig) error {
	ring, err := secret.GetKeyring()
	if err != nil {
		return err
	}

	// Client secret
	if config.ClientSecret != "" {
		err = ring.Set(keyring.Item{
			Key:  "client_secret",
			Data: []byte(config.ClientSecret),
		})
		if err != nil {
			return err
		}
	}

	// Password
	if config.Password != nil {
		pw := []byte(*config.Password)
		err = ring.Set(keyring.Item{
			Key:  "password",
			Data: pw,
		})
		if err != nil {
			return err
		}
	}

	// Private Key
	if config.PrivateKey != nil {
		pk := []byte(*config.PrivateKey)
		err = ring.Set(keyring.Item{
			Key:  "private_key",
			Data: pk,
		})
		if err != nil {
			return err
		}
	}

	// Private Key Password
	if config.PrivateKeyPassword != nil {
		pwpk := []byte(*config.PrivateKeyPassword)
		err = ring.Set(keyring.Item{
			Key:  "private_key_password",
			Data: pwpk,
		})
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
