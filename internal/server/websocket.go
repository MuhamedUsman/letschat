package server

import (
	"context"
	"errors"
	"github.com/M0hammadUsman/letschat/internal/common"
	"github.com/M0hammadUsman/letschat/internal/domain"
	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	"log/slog"
	"net/http"
	"sync"
	"time"
)

var (
	ErrAlreadySubscribed = errors.New("already subscribed")
)

func (s *Server) WebsocketSubscribeHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := s.subscribe(w, r)
	if err != nil {
		switch {
		case errors.Is(err, ErrAlreadySubscribed):
			s.redundantSubscription(w, r)
		default:
			slog.Error(err.Error())
		}
		return
	}
	defer conn.CloseNow()

	u := common.ContextGetUser(r.Context())
	s.addSubscriber(u)
	defer s.removeSubscriber(u)

	// retrieving unread messages if any
	if err = s.Facade.WriteUnreadMessagesToWSConn(r.Context(), u.Messages); err != nil {
		slog.Error(err.Error())
		return
	}

	errChan := make(chan error, 2)
	s.BackgroundTask.Run(func(shtdwnCtx context.Context) {
		errChan <- s.handleReceivedMessages(shtdwnCtx, r.Context(), conn)
	})
	s.BackgroundTask.Run(func(shtdwnCtx context.Context) {
		errChan <- s.handleSentMessages(shtdwnCtx, r.Context(), conn)
	})

	for range cap(errChan) {
		if err = <-errChan; err != nil {
			if websocket.CloseStatus(err) == websocket.StatusNormalClosure ||
				websocket.CloseStatus(err) == websocket.StatusGoingAway {
				return
			}
			slog.Error(err.Error())
		}
	}
}

func (s *Server) subscribe(w http.ResponseWriter, r *http.Request) (*websocket.Conn, error) {
	var mu sync.Mutex
	var conn *websocket.Conn

	u := common.ContextGetUser(r.Context()) // User will be authenticated and setup in the context using middleware
	if _, ok := s.Subscribers[u.ID]; ok {   // multiple online instances of the account are not allowed by design
		return nil, ErrAlreadySubscribed
	}
	u.Messages = make(chan *domain.Message, s.subscriberMessageBuffer)
	u.CloseSlow = func() {
		mu.Lock()
		defer mu.Unlock()
		if conn != nil {
			conn.Close(websocket.StatusPolicyViolation, "connection too slow to keep up with messages")
		}
	}
	r = common.ContextSetUser(r, u) // setting back updated user in context
	c, err := websocket.Accept(w, r, s.wsAcceptOpts)
	if err != nil {
		return nil, err
	}
	mu.Lock()
	conn = c
	mu.Unlock()
	return conn, nil
}

func (*Server) handleReceivedMessages(shutdownCtx, reqCtx context.Context, conn *websocket.Conn) error {
	u := common.ContextGetUser(reqCtx)
	// Listening for messages for this user
	for {
		select {
		case msg := <-u.Messages:
			if err := writeWithTimeout(conn, 5*time.Second, msg); err != nil {
				return err
			}
		case <-shutdownCtx.Done():
			return nil
		}
	}
}

func (s *Server) handleSentMessages(shutdownCtx, reqCtx context.Context, conn *websocket.Conn) error {
	u := common.ContextGetUser(reqCtx)
	for {
		var ms domain.MessageSent
		if err := wsjson.Read(reqCtx, conn, &ms); err != nil {
			s.wsInvalidJsonResponse(conn)
			return errors.Unwrap(err)
		}
		// ProcessSentMessage populate the domain.Message and also concurrently persist it to DB with 5 retries
		msg, ev := s.Facade.ProcessSentMessage(reqCtx, ms)
		if ev != nil {
			handleValidationError(conn, ev)
			continue
		}
		if relayTo, ok := s.Subscribers[ms.ReceiverID]; ok {
			select {
			case relayTo.Messages <- msg:
			case <-shutdownCtx.Done():
				return nil
			default:
				u.CloseSlow()
				return nil
			}
		}
	}
}

func (s *Server) addSubscriber(u *domain.User) {
	s.SubsMu.Lock()
	s.Subscribers[u.ID] = u
	s.SubsMu.Unlock()
}

func (s *Server) removeSubscriber(u *domain.User) {
	s.SubsMu.Lock()
	delete(s.Subscribers, u.ID)
	s.SubsMu.Unlock()
}

func writeWithTimeout(conn *websocket.Conn, t time.Duration, msg any) error {
	ctx, cancel := context.WithTimeout(context.Background(), t)
	defer cancel()
	return wsjson.Write(ctx, conn, msg)
}

func handleValidationError(conn *websocket.Conn, err error) {
	var ev *domain.ErrValidation
	if errors.As(err, &ev) {
		if err = writeWithTimeout(conn, 5*time.Second, ev.Errors); err != nil {
			slog.Error(err.Error())
			return
		}
	}
}
