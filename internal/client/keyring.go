package client

import (
	"github.com/99designs/keyring"
)

var (
	serviceName = "auth-token"
	tokenKey    = "letschat-auth-token"
)

type keyringManager struct {
	kr keyring.Keyring
}

func newKeyringManager() (*keyringManager, error) {
	kr, err := keyring.Open(keyring.Config{ServiceName: serviceName})
	if err != nil {
		return nil, err
	}
	return &keyringManager{kr: kr}, nil
}

func (k *keyringManager) setAuthTokenInKeyring(data string) error {
	item := keyring.Item{
		Key:         tokenKey,
		Data:        []byte(data),
		Label:       serviceName,
		Description: "auth token to validate user after basic login",
	}
	return k.kr.Set(item)
}

func (k *keyringManager) getAuthTokenFromKeyring() (string, error) {
	token, err := k.kr.Get(tokenKey)
	if err != nil {
		return "", err
	}
	return string(token.Data), nil
}
