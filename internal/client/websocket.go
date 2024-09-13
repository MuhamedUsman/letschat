package client

import (
	"context"
	"errors"
	"github.com/M0hammadUsman/letschat/internal/domain"
	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	"log/slog"
	"math"
	"math/rand/v2"
	"net/http"
	"time"
)

// WsConnectAndListenForMessages connects to ws and listen for messages,
// writes Disconnected and Connected wsConnStatus to WsConnStateChan
// we read on WsConnStateChan for reconnection and stuff
func (c *Client) WsConnectAndListenForMessages() {
	h := make(http.Header)
	h.Set("Authorization", "Bearer "+c.AuthToken)
	opts := &websocket.DialOptions{
		CompressionMode: websocket.CompressionContextTakeover,
		HTTPHeader:      h,
	}
	conn, r, err := websocket.Dial(context.Background(), subscribeTo, opts)
	if err != nil {
		if r != nil && r.StatusCode == http.StatusUnauthorized {
			c.UsrLoggedIn <- false
		}
		c.wsConnStatus <- Disconnected
		return
	}
	defer conn.CloseNow()
	if r.StatusCode != http.StatusSwitchingProtocols {
		c.wsConnStatus <- Disconnected
		return
	}
	c.wsConnStatus <- Connected
	errChan := make(chan error)
	c.BT.Run(func(shtdwnCtx context.Context) {
		errChan <- c.writeMessagesToClientMsgChannel(conn, shtdwnCtx)
	})
	/*c.BT.Run(func(shtdwnCtx context.Context) {
		errChan <- c.pingServer(conn, shtdwnCtx)
	})*/
	if err = <-errChan; err != nil {
		c.wsConnStatus <- Disconnected // disconnect
		if websocket.CloseStatus(err) == websocket.StatusNormalClosure ||
			websocket.CloseStatus(err) == websocket.StatusGoingAway ||
			websocket.CloseStatus(err) == websocket.StatusAbnormalClosure ||
			errors.Is(err, context.Canceled) {
			return
		}
		slog.Error(err.Error())
	}
	conn.Close(websocket.StatusNormalClosure, "client exited letschat")
}

/*func (c *Client) pingServer(conn *websocket.Conn, shtdwnCtx context.Context) error {
	t := time.NewTimer(5 * time.Second)
	for {
		select {
		case <-t.C:
			if err := conn.Ping(shtdwnCtx); err != nil {
				c.wsConnStatus <- Disconnected
				return err
			}
		case <-shtdwnCtx.Done():
			return nil
		}
		t.Reset(5 * time.Second)
	}
}*/

func (c *Client) writeMessagesToClientMsgChannel(conn *websocket.Conn, shtdwnCtx context.Context) error {
	for {
		var msg domain.Message
		if err := wsjson.Read(shtdwnCtx, conn, &msg); err != nil {
			return err
		}
		c.messages <- &msg
	}
}

// AttemptWsReconnectOnDisconnect must be run in a separate go routine, principal -> finite state machine
// Why not reconnect from WsConnectAndListenForMessages method, the reason is,
// we'll only get disconnect error once we try to write to this conn,
// and also we use this chan to read and communicate to the user in TUI what is happening in the background
func (c *Client) AttemptWsReconnectOnDisconnect() {
	attempt := 1
	maxAttempts := 5
	maxDelay := 30 * time.Second
	c.BT.Run(func(shtdwnCtx context.Context) {
		for {
			select {
			case state := <-c.wsConnStatus:
				// we switch on the state so we can attempt a reconnect while skipping the backoff time if need be
				switch state {
				case Disconnected:
					c.ConnState = Disconnected
					// writing to same chan like this if it is unbuffered, write will block because there is no
					// read as we are writing in the same select block that reads on this chan
					// I found out the hard way
					c.wsConnStatus <- WaitingForReconnection
				case WaitingForReconnection:
					// After 5th retry
					if attempt == maxAttempts {
						return
					}
					c.ConnState = WaitingForReconnection
					// reconnecting after backoff time
					expbackoff := exponentialBackoff(attempt, maxDelay)
					t := time.NewTimer(expbackoff)
					select {
					case <-t.C:
						attempt++
						c.wsConnStatus <- Reconnecting
					case <-shtdwnCtx.Done():
						// stop on timer is not necessary after go 1.23
						return
					}
				case Reconnecting:
					c.ConnState = Reconnecting
					go c.WsConnectAndListenForMessages()
					attempt++
				case Connected:
					c.ConnState = Connected
					attempt = 0
				}
			case <-shtdwnCtx.Done():
				return
			}
		}
	})
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

func writeWithTimeout(conn *websocket.Conn, t time.Duration, msg any) error {
	ctx, cancel := context.WithTimeout(context.Background(), t)
	defer cancel()
	return wsjson.Write(ctx, conn, msg)
}
