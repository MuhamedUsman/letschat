package client

import (
	"github.com/99designs/keyring"
)

var (
	appName     = "Letschat"
	serviceName = " Auth"
	tokenKey    = " Access Token"
)

type keyringManager struct {
	kr keyring.Keyring
}

func newKeyringManager() (*keyringManager, error) {
	cfg := keyring.Config{
		ServiceName:             serviceName,
		KeyCtlScope:             "user",
		LibSecretCollectionName: appName,
		WinCredPrefix:           appName,
	}
	kr, err := keyring.Open(cfg)
	if err != nil {
		return nil, err
	}
	return &keyringManager{kr: kr}, nil
}

func (k *keyringManager) setAuthTokenInKeyring(label, data string) error {
	item := keyring.Item{
		Key:         tokenKey,
		Data:        []byte(data),
		Description: "auth token to validate user after basic login",
	}
	item.Label = "user=" + label
	return k.kr.Set(item)
}

func (k *keyringManager) removeAuthTokenFromKeyring() error {
	return k.kr.Remove(tokenKey)
}

func (k *keyringManager) getAuthTokenFromKeyring() string {
	token, err := k.kr.Get(tokenKey)
	if err != nil {
		return ""
	}
	return string(token.Data)
}
