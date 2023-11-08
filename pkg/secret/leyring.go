package secret

import (
	"github.com/99designs/keyring"
	"github.com/G-PORTAL/gpcloud-cli/pkg/consts"
)

func GetKeyring() (keyring.Keyring, error) {
	return keyring.Open(keyring.Config{
		ServiceName:                    consts.BinaryName,
		KeychainName:                   consts.BinaryName,
		KeychainTrustApplication:       true,
		KeychainSynchronizable:         true,
		KeychainAccessibleWhenUnlocked: true,
	})
}
