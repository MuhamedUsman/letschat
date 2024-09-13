package client

import (
	"encoding/json"
	"github.com/M0hammadUsman/letschat/internal/domain"
	"io"
	"log/slog"
	"net/http"
)

func (c *Client) GetConversations() ([]*domain.Conversation, error, int) {
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
