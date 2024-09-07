package client

import (
	"github.com/M0hammadUsman/letschat/internal/api/common"
	"github.com/M0hammadUsman/letschat/internal/domain"
	"sync"
)

var (
	once   sync.Once
	client *Client
)

type Conversations map[string][]*domain.Message // key -> userID, val -> list of messages

type Client struct {
	AuthToken string // if zero valued -> requires login
	krm       *keyringManager
	// initiated by tui.LetschatModel so we can communicate proper error in TUI
	messages      domain.MsgChan
	conversations Conversations
	bt            *common.BackgroundTask
}

func Init() error {
	var c Client
	var err error
	once.Do(func() {
		c.krm, err = newKeyringManager()
		// ignoring the error, we'll determine if the item is not found using zero value of Client.AuthToken
		c.AuthToken, _ = c.krm.getAuthTokenFromKeyring()
		c.messages = make(domain.MsgChan, 16)
		c.bt = common.NewBackgroundTask()
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

func (c *Client) ListenForMessages() domain.MsgChan {
	return c.messages
}
