package client

import (
	"context"
	"errors"
	"github.com/MuhamedUsman/letschat/internal/domain"
	"github.com/MuhamedUsman/letschat/internal/sync"
	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	"log/slog"
	"math"
	"math/rand/v2"
	"net/http"
	"time"
)

// WsConnState will be switch cased by tui.TabContainerModel, so making it unique
type WsConnState int

const (
	Disconnected WsConnState = iota - 1
	Idle
	WaitingForConnection
	Connecting
	Connected
)

type WsConnBroadcaster = sync.Broadcaster[WsConnState]

func newWsConnBroadcaster() *WsConnBroadcaster {
	return sync.NewBroadcaster[WsConnState]()
}

// WsConnectAndListenForMessages connects to ws and listen for recvMsgs,
// writes Disconnected and Connected wsConnStatus to WsConnStateChan
// we read on WsConnStateChan for reconnection and stuff
func (c *Client) wsConnectAndListenForMessages(shtdwnCtx context.Context) {
	h := make(http.Header)
	h.Set("Authorization", "Bearer "+c.AuthToken)
	opts := &websocket.DialOptions{
		CompressionMode: websocket.CompressionContextTakeover,
		HTTPHeader:      h,
	}
	conn, r, err := websocket.Dial(context.Background(), subscribeTo, opts)
	c.wsConn = conn
	if err != nil {
		if r != nil && r.StatusCode == http.StatusUnauthorized {
			c.LoginState.Write(false)
		}
		if c.LoginState.Get() {
			c.WsConnState.Write(Disconnected)
		}
		return
	}
	defer conn.CloseNow()
	if r.StatusCode != http.StatusSwitchingProtocols {
		if c.LoginState.Get() {
			c.WsConnState.Write(Disconnected)
		}
		return
	}
	c.WsConnState.Write(Connected)
	errChan := make(chan error)
	go func() { errChan <- c.handleSentMessages(conn, shtdwnCtx) }()
	go func() { errChan <- c.handleReceiveMessages(conn, shtdwnCtx) }()
	if err = <-errChan; err != nil {
		if shtdwnCtx.Err() == nil && c.LoginState.Get() { // In case the shtdwnCtx is canceled we do not signal a Disconnect
			c.WsConnState.Write(Disconnected)
		}
		if websocket.CloseStatus(err) == websocket.StatusNormalClosure ||
			websocket.CloseStatus(err) == websocket.StatusGoingAway ||
			websocket.CloseStatus(err) == websocket.StatusAbnormalClosure ||
			errors.Is(err, context.Canceled) {
			if err = conn.Close(websocket.StatusNormalClosure, "client exited letschat"); err != nil {
				slog.Error(err.Error())
			}
			return
		}
	}
	conn.Close(websocket.StatusNormalClosure, "client exited letschat")
}

func (c *Client) handleReceiveMessages(conn *websocket.Conn, shtdwnCtx context.Context) error {
	for {
		var msg domain.Message
		if err := wsjson.Read(shtdwnCtx, conn, &msg); err != nil {
			return err
		}
		c.RecvMsgs.Write(&msg)
	}
}

func (c *Client) handleSentMessages(conn *websocket.Conn, shtdwnCtx context.Context) error {
	msgChan := make(chan *domain.Message)
	doneChan := make(chan bool)
	// ensuring no misuse, making it <- unidirectional
	c.sentMsgs.msgs = msgChan
	c.sentMsgs.done = doneChan
	for {
		select {
		case msg := <-msgChan:
			if msg.Operation == domain.DeliveredMsg {
			}
			if err := writeWithTimeout(conn, 2*time.Second, msg); err != nil {
				doneChan <- false
				return err
			}
			doneChan <- true
		case <-shtdwnCtx.Done():
			return shtdwnCtx.Err()
		}
	}
}

// AttemptWsReconnectOnDisconnect must be run in a separate go routine, principal -> finite state machine
func (c *Client) attemptWsReconnectOnDisconnect(shtdwnCtx context.Context) {
	token, ch := c.WsConnState.Subscribe()
	defer c.WsConnState.Unsubscribe(token)
	attempt := 1
	maxAttempts := 5
	maxDelay := 40 * time.Second
	for {
		select {
		case s := <-ch:
			// we switch on the s, so we can attempt a reconnect while skipping the backoff time if need be
			switch s {
			case Disconnected:
				c.WsConnState.Write(WaitingForConnection)
			case Idle:
				if c.wsConn != nil {
					c.wsConn.CloseNow()
				}
				attempt = 0
				// do nothing, will be the case when user is logging in or signing up
			case WaitingForConnection:
				// After 5th retry
				if attempt == maxAttempts {
					c.WsConnState.Write(Disconnected)
					return
				}
				// reconnecting after backoff time
				expbackoff := exponentialBackoff(attempt, maxDelay)
				t := time.NewTimer(expbackoff)
				select {
				case <-t.C:
					c.WsConnState.Write(Connecting)
				case <-shtdwnCtx.Done():
					// stop on timer is not necessary after go 1.23
					return
				}
			case Connecting:
				go c.wsConnectAndListenForMessages(shtdwnCtx)
				attempt++
			case Connected:
				attempt = 0
			}
		case <-shtdwnCtx.Done():
			return
		}
	}
}

// Helpers & Stuff -----------------------------------------------------------------------------------------------------

func exponentialBackoff(attempt int, maxDelay time.Duration) time.Duration {
	delay := time.Duration(math.Pow(2, float64(attempt))) * time.Second
	jitter := time.Duration(rand.IntN(int(time.Second)))
	delay += jitter
	if delay > maxDelay {
		delay = maxDelay
	}
	return delay
}

func writeWithTimeout(conn *websocket.Conn, t time.Duration, msg *domain.Message) error {
	ctx, cancel := context.WithTimeout(context.Background(), t)
	defer cancel()
	return wsjson.Write(ctx, conn, msg)
}
