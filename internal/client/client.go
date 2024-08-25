package client

import (
	"sync"
)

var (
	once   sync.Once
	client *Client
)

type Client struct {
	AuthToken string // if zero requires login
	krm       *keyringManager
}

func Init() error {
	var c Client
	var err error
	once.Do(func() {
		c.krm, err = newKeyringManager()
		// ignoring the error, we'll determine if the item is not found using zero value of Client.AuthToken
		c.AuthToken, _ = c.krm.getAuthTokenFromKeyring()
	})
	if err != nil {
		return err
	}
	client = &c
	return nil
}

func Get() *Client {
	return client
}
