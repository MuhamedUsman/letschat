package client

import (
	"context"
	"errors"
	"github.com/M0hammadUsman/letschat/internal/client/sync"
	"github.com/M0hammadUsman/letschat/internal/domain"
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
	WaitingForReconnection
	Reconnecting
	Connected
)

type WsConnMonitor = sync.StateMonitor[WsConnState]

func newWsConnMonitor() *WsConnMonitor {
	return sync.NewStatus[WsConnState](Disconnected)
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
	if err != nil {
		if r != nil && r.StatusCode == http.StatusUnauthorized {
			c.LoginState.WriteToChan(false)
		}
		c.WsConnState.WriteToChan(Disconnected)
		return
	}
	defer conn.CloseNow()
	if r.StatusCode != http.StatusSwitchingProtocols {
		c.WsConnState.WriteToChan(Disconnected)
		return
	}
	c.WsConnState.WriteToChan(Connected)
	errChan := make(chan error)
	go func() { errChan <- c.handleSentMessages(conn, shtdwnCtx) }()
	go func() { errChan <- c.handleReceiveMessages(conn, shtdwnCtx) }()
	if err = <-errChan; err != nil {
		if shtdwnCtx.Err() == nil { // In case the shtdwnCtx is canceled we do not signal a Disconnect
			c.WsConnState.WriteToChan(Disconnected)
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
		slog.Error(err.Error())
	}
	conn.Close(websocket.StatusNormalClosure, "client exited letschat")
}

func (c *Client) handleReceiveMessages(conn *websocket.Conn, shtdwnCtx context.Context) error {
	for {
		var msg domain.Message
		if err := wsjson.Read(shtdwnCtx, conn, &msg); err != nil {
			return err
		}
		c.RecvMsgs.WriteToChan(&msg)
		c.RecvMsgs.Broadcast(shtdwnCtx)
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
// Why not reconnect from WsConnectAndListenForMessages method, the reason is,
// we'll only get disconnect error once we try to write to this conn,
// and also we use this chan to read and communicate to the user in TUI what is happening in the background
func (c *Client) attemptWsReconnectOnDisconnect(shtdwnCtx context.Context) {
	attempt := 1
	maxAttempts := 5
	maxDelay := 30 * time.Second
	for {
		s := c.WsConnState.WaitForStateChange()
		select {
		case <-shtdwnCtx.Done():
			return
		default:
		}
		// we switch on the s so we can attempt a reconnect while skipping the backoff time if need be
		switch s {
		case Disconnected:
			c.WsConnState.WriteToChan(WaitingForReconnection)
		case WaitingForReconnection:
			// After 5th retry
			if attempt == maxAttempts {
				c.WsConnState.WriteToChan(Disconnected)
				return
			}
			// reconnecting after backoff time
			expbackoff := exponentialBackoff(attempt, maxDelay)
			t := time.NewTimer(expbackoff)
			select {
			case <-t.C:
				attempt++
				c.WsConnState.WriteToChan(Reconnecting)
			case <-shtdwnCtx.Done():
				// stop on timer is not necessary after go 1.23
				return
			}
		case Reconnecting:
			go c.wsConnectAndListenForMessages(shtdwnCtx)
			attempt++
		case Connected:
			attempt = 0
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
