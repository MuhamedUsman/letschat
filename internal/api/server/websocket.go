package server

import (
	"context"
	"errors"
	"github.com/M0hammadUsman/letschat/internal/api/utility"
	"github.com/M0hammadUsman/letschat/internal/domain"
	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	"io"
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
	u := utility.ContextGetUser(r.Context())

	s.addSubscriber(u)
	if err = s.Facade.UpdateUserOnlineStatus(r.Context(), u, true); err != nil {
		s.editConflictResponse(w, r)
		return
	}
	if err = s.broadcastUserOnlineStatus(r.Context(), u, true); err != nil {
		s.serverErrorResponse(w, r, err)
		return
	}

	// buffered because if there is any error we'll return so we don't want the other writes to block
	errChan := make(chan error, 1) // if there is a single err we log and return
	reqCtx, cancel := context.WithCancel(r.Context())
	defer cancel()
	s.BackgroundTask.Run(func(shtdwnCtx context.Context) {
		errChan <- s.handleReceivedMessages(shtdwnCtx, reqCtx, conn)
	})
	s.BackgroundTask.Run(func(shtdwnCtx context.Context) {
		errChan <- s.handleSentMessages(shtdwnCtx, reqCtx, conn)
	})

	if err = s.Facade.WriteUnDeliveredMessagesToWSConn(r.Context(), u.Messages); err != nil {
		slog.Error(err.Error())
		return
	}
	defer s.WebsocketSubscribeHandlerDeferFunc(r.Context(), conn)

	if err = <-errChan; err != nil {
		// Once there is an error from one of the background tasks,
		// means the Ws connection is closed so we cancel the reqCtx
		// so the other background task can exit listening on it
		cancel()
		if websocket.CloseStatus(err) == websocket.StatusNormalClosure ||
			websocket.CloseStatus(err) == websocket.StatusAbnormalClosure ||
			websocket.CloseStatus(err) == websocket.StatusGoingAway ||
			errors.Is(err, io.EOF) ||
			errors.Is(err, context.Canceled) {
			return
		}
		slog.Error(err.Error())
	}
}

// WebsocketSubscribeHandlerDeferFunc sets the user's LastOnline to time.Now
func (s *Server) WebsocketSubscribeHandlerDeferFunc(reqCtx context.Context, conn *websocket.Conn) {
	u := utility.ContextGetUser(reqCtx)
	s.broadcastUserOnlineStatus(reqCtx, u, false)
	s.removeSubscriber(u)
	conn.CloseNow()
	for range 5 { // Very unlikely to fail
		if err := s.Facade.UpdateUserOnlineStatus(reqCtx, u, false); err == nil { // successful case
			break
		}
	}
}

func (s *Server) subscribe(w http.ResponseWriter, r *http.Request) (*websocket.Conn, error) {
	var mu sync.Mutex
	var conn *websocket.Conn

	u := utility.ContextGetUser(r.Context()) // User will be authenticated and setup in the context using middleware
	if _, ok := s.Subscribers[u.ID]; ok {    // multiple online instances of the account are not allowed by design
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
	r = utility.ContextSetUser(r, u) // setting back updated user in context
	c, err := websocket.Accept(w, r, s.wsAcceptOpts)
	if err != nil {
		return nil, err
	}
	mu.Lock()
	conn = c
	mu.Unlock()
	return conn, nil
}

func (s *Server) handleReceivedMessages(shutdownCtx, reqCtx context.Context, conn *websocket.Conn) error {
	u := utility.ContextGetUser(reqCtx)
	// Listening for messages for this user
	for {
		select {
		case msg := <-u.Messages:
			if s.publishLimiter.Allow() {
				if err := writeWithTimeout(conn, 2*time.Second, msg); err != nil {
					slog.Error(err.Error())
					return err
				}
			}
		case <-reqCtx.Done():
			return nil
		case <-shutdownCtx.Done():
			return nil
		}
	}
}

func (s *Server) handleSentMessages(shutdownCtx, reqCtx context.Context, conn *websocket.Conn) error {
	u := utility.ContextGetUser(reqCtx)
	for {
		var ms domain.MessageSent
		// read will immediately errors out once the client shuts the Ws connection
		if err := wsjson.Read(shutdownCtx, conn, &ms); err != nil {
			return err
		}
		// ProcessSentMessage populate the domain.Message and also concurrently persist it to DB with 5 retries
		msg, err := s.Facade.ProcessSentMessage(reqCtx, ms, u)
		if err != nil {
			if errors.Is(err, domain.ErrValidation{}) {
				handleValidationError(conn, err)
			} else {
				return err
			}
			continue
		}
		if relayTo, ok := s.Subscribers[ms.ReceiverID]; ok {
			// we do not want to send msg, these Ops are only for ack to server
			if msg.Operation == domain.DeliveredConfirmMsg ||
				msg.Operation == domain.ReadConfirmMsg ||
				msg.Operation == domain.DeleteConfirmMsg {
				continue
			}
			select {
			case relayTo.Messages <- msg:
			case <-shutdownCtx.Done():
				return nil
			case <-reqCtx.Done():
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

func (s *Server) broadcastUserOnlineStatus(ctx context.Context, u *domain.User, online bool) error {
	convos, err := s.Facade.GetConversations(ctx)
	if err != nil {
		return err
	}
	for _, convo := range convos {
		if convo.LastOnline != nil { // meaning the user is not online
			continue
		}
		t := time.Now()
		op := domain.OfflineMsg
		if online {
			op = domain.OnlineMsg
		}
		msg := domain.Message{
			SenderID:  u.ID,
			SentAt:    &t,
			Operation: op,
		}
		if v, ok := s.Subscribers[convo.UserID]; ok {
			v.Messages <- &msg
		}
	}
	return nil
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
