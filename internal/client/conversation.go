package client

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/M0hammadUsman/letschat/internal/domain"
	"github.com/M0hammadUsman/letschat/internal/sync"
	"io"
	"log"
	"log/slog"
	"net/http"
	"slices"
)

type Convos []*domain.Conversation

type ConvosBroadcaster = sync.Broadcaster[Convos]

func newConvosBroadcaster() *ConvosBroadcaster {
	return sync.NewBroadcaster[Convos]()
}

// populateConversationsAccordingToWsConnState gets the conversations once there is a read from WsConnStateChan
// and populates the conversations in Client, once the connection state is Connected it fetches from the server,
// in case of Disconnected it retrieves from the local db
func (c *Client) populateConversationsAccordingToWsConnState(shtdwnCtx context.Context) {
	token, ch := c.WsConnState.Subscribe()
	defer c.WsConnState.Unsubscribe(token)
	for {
		select {
		case s := <-ch:
			switch s {
			case Connected:
				convos, code, err := c.getConversations()
				if err != nil { // fetch from db
					convos, err = c.repo.GetConversations()
					if err != nil {
						log.Fatal(err)
					}
				}
				if code == http.StatusUnauthorized {
					c.LoginState.Write(false) // user will be redirected to log-in by tui
				} else {
					c.populateConvosAndWriteToChan(convos)
					_ = c.repo.DeleteAllConversations()
					_ = c.repo.SaveConversations(convos...) // ignore the error
				}
			case Disconnected:
				convos, err := c.repo.GetConversations()
				if err != nil {
					log.Fatal(err)
				}
				c.populateConvosAndWriteToChan(convos)
			default:
			}
		case <-shtdwnCtx.Done():
			return
		}
	}
}

func (c *Client) getConversations() ([]*domain.Conversation, int, error) {
	r, err := http.NewRequest(http.MethodGet, getConversations, nil)
	if err != nil {
		slog.Error(err.Error())
		return nil, 0, ErrApplication
	}
	r.Header.Set("Authorization", "Bearer "+c.AuthToken)
	resp, err := http.DefaultClient.Do(r)
	if err != nil {
		slog.Error(err.Error())
		return nil, http.StatusServiceUnavailable, getMostNestedError(err)
	}
	defer resp.Body.Close()
	readBody, _ := io.ReadAll(resp.Body)
	var res struct {
		Conversations []*domain.Conversation `json:"conversations"`
	}
	if err = json.Unmarshal(readBody, &res); err != nil {
		slog.Error(err.Error())
		return nil, 0, ErrApplication
	}
	return res.Conversations, resp.StatusCode, nil
}

func (c *Client) populateConvosAndWriteToChan(convos []*domain.Conversation) {
	cui := make([]string, len(convos))
	for i, convo := range convos {
		cui[i] = convo.UserID
	}
	latestMsgs, _ := c.repo.GetLatestMsgBodyForConvos(cui...)
	for i, convo := range convos {
		if msg, ok := latestMsgs[convo.UserID]; ok {
			convos[i].LatestMsg = msg.Body
			convos[i].LatestMsgSentAt = msg.SentAt
		}
	}
	// sort in descending order latest msgs conversations first
	slices.SortFunc(convos, func(a, b *domain.Conversation) int {
		if b.LatestMsgSentAt == nil && a.LatestMsgSentAt == nil {
			return 0
		}
		if b.LatestMsgSentAt == nil {
			return -1
		}
		if a.LatestMsgSentAt == nil {
			return 1
		}
		return b.LatestMsgSentAt.Compare(*a.LatestMsgSentAt)
	})
	c.Conversations.Write(convos)
}

func (c *Client) CreateConvoIfNotExist(convo *domain.Conversation) error {
	if _, err := c.repo.GetConversationByUserID(convo.UserID); err != nil {
		if errors.Is(err, domain.ErrRecordNotFound) {
			// we need to create a new convo
			return c.repo.SaveConversations(convo)
		}
		return err
	}
	return nil
}
