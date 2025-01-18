package server

import (
	"context"
	"github.com/M0hammadUsman/letschat/internal/api/utility"
	"github.com/M0hammadUsman/letschat/internal/domain"
	"net/http"
	"time"
)

func (s *Server) GetConversationsHandler(w http.ResponseWriter, r *http.Request) {
	c, err := s.Facade.GetConversations(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
	if err = s.writeJSON(w, envelop{"conversations": c}, http.StatusOK, nil); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

// Once the receivers gets this broadcast, they will re-fetch the conversations, for synchronization
func (s *Server) syncConvos(ctx context.Context) error {
	convos, err := s.Facade.GetConversations(ctx)
	if err != nil {
		return err
	}
	u := utility.ContextGetUser(ctx)
	if u == nil {
		panic("no user was found in the context, Hint: missing Authentication middleware")
	}
	for _, convo := range convos {
		if convo.LastOnline != nil { // meaning the user is not online
			continue
		}
		t := time.Now()
		msg := domain.Message{
			SenderID:  u.ID,
			SentAt:    &t,
			Operation: domain.SyncConvosMsg,
		}
		if v, ok := s.Subscribers[convo.UserID]; ok {
			v.Messages <- &msg
		}
	}
	return nil
}
