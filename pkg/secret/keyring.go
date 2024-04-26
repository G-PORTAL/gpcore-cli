package secret

import (
	"errors"
	"github.com/G-PORTAL/gpcore-cli/pkg/consts"
	"github.com/charmbracelet/log"
	zalandoKeyring "github.com/zalando/go-keyring"
)

var ErrKeyNotFound = errors.New("the key does not exist in the keyring")
var ErrInvalidData = errors.New("they keyring data is too large to set")

func GetKeyring() Keyring {
	return &ZalandoKeyring{}
}

type Keyring interface {
	Get(key string) (string, error)
	Set(key, value string) error
	Remove(key string) error
}

type ZalandoKeyring struct{}

func (z *ZalandoKeyring) Get(key string) (string, error) {
	log.Debugf("Getting key %q from keyring", key)
	value, err := zalandoKeyring.Get(consts.BinaryName, key)
	if errors.Is(err, zalandoKeyring.ErrNotFound) {
		return "", ErrKeyNotFound
	}

	return value, err
}

func (z *ZalandoKeyring) Set(key, value string) error {
	log.Debugf("Setting key %q in keyring", key)
	err := zalandoKeyring.Set(consts.BinaryName, key, value)
	if errors.Is(err, zalandoKeyring.ErrNotFound) {
		return ErrKeyNotFound
	}

	if errors.Is(err, zalandoKeyring.ErrSetDataTooBig) {
		return ErrInvalidData
	}

	return err
}

func (z *ZalandoKeyring) Remove(key string) error {
	log.Debugf("Removing key %q from keyring", key)
	err := zalandoKeyring.Delete(consts.BinaryName, key)
	if errors.Is(err, zalandoKeyring.ErrNotFound) {
		return ErrKeyNotFound
	}

	return err
}
