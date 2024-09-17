package client

import (
	"context"
	"encoding/json"
	"github.com/M0hammadUsman/letschat/internal/client/sync"
	"github.com/M0hammadUsman/letschat/internal/domain"
	"io"
	"log"
	"log/slog"
	"net/http"
)

type Conversations []*domain.Conversation
type ConversationsMonitor = sync.StateMonitor[Conversations]

func newConversationsMonitor() *ConversationsMonitor {
	return sync.NewStatus[Conversations](nil)
}

// populateConversationsAccordingToWsConnState gets the conversations once there is a read from WsConnStateChan
// and populates the conversations in Client, once the connection state is Connected it fetches from the server,
// in case of Disconnected it retrieves from the local db
func (c *Client) populateConversationsAccordingToWsConnState(shtdwnCtx context.Context) {
	for {
		s := c.WsConnState.WaitForStateChange()
		select {
		case <-shtdwnCtx.Done():
			return
		default:
		}
		switch s {
		case Connected:
			convos, err, code := c.getConversations()
			if err != nil { // fetch from db
				convos, err = c.repo.GetConversations()
				if err != nil {
					log.Fatal(err)
				}
			}
			if code == http.StatusUnauthorized {
				c.LoginState.WriteToChan(false) // user will be redirected to log-in by tui
			} else {
				c.populateConvosWithLatestMsgs(convos)
				c.Conversations.WriteToChan(convos)
				_ = c.repo.SaveConversations(convos...) // ignore the error
			}
		case Disconnected:
			convos, err := c.repo.GetConversations()
			c.populateConvosWithLatestMsgs(convos)
			if err != nil {
				log.Fatal(err)
			}
			c.Conversations.WriteToChan(convos)
		default:
		}
	}
}

func (c *Client) getConversations() ([]*domain.Conversation, error, int) {
	r, err := http.NewRequest(http.MethodGet, getConversations, nil)
	if err != nil {
		slog.Error(err.Error())
		return nil, ErrApplication, 0
	}
	r.Header.Set("Authorization", "Bearer "+c.AuthToken)
	resp, err := http.DefaultClient.Do(r)
	if err != nil {
		slog.Error(err.Error())
		return nil, getMostNestedError(err), http.StatusServiceUnavailable
	}
	defer resp.Body.Close()
	readBody, _ := io.ReadAll(resp.Body)
	var res struct {
		Conversations []*domain.Conversation `json:"conversations"`
	}
	if err = json.Unmarshal(readBody, &res); err != nil {
		slog.Error(err.Error())
		return nil, ErrApplication, 0
	}
	return res.Conversations, nil, resp.StatusCode
}

func (c *Client) populateConvosWithLatestMsgs(convos []*domain.Conversation) {
	cui := make([]string, len(convos))
	for i, convo := range convos {
		cui[i] = convo.UserID
	}
	latestMsgs, _ := c.repo.GetLatestMsgBodyForConvos(cui...)
	for _, convo := range convos {
		convo.LatestMsg = latestMsgs[convo.UserID]
	}
}

func (c *Client) writeUpdatedConvosToChan() {
	convos, _ := c.repo.GetConversations()
	c.Conversations.WriteToChan(convos)
}
