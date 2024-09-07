package client

import (
	"context"
	"github.com/M0hammadUsman/letschat/internal/domain"
	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	"log"
	"net/http"
)

func (c *Client) WsConnect() (error, int) {
	opts := &websocket.DialOptions{CompressionMode: websocket.CompressionContextTakeover}
	conn, resp, err := websocket.Dial(context.Background(), subscribeTo, opts)
	if err != nil {
		log.Println(err.Error())
		return getMostNestedError(err), http.StatusServiceUnavailable
	}
	defer resp.Body.Close()
	defer conn.CloseNow()
	if resp.StatusCode == http.StatusOK {
		err, code := c.writeMessagesToClientMsgChannel(conn)
		if err != nil {
			log.Println(err.Error())
			return getMostNestedError(err), code
		}
	}
	return nil, 0
}

func (c *Client) writeMessagesToClientMsgChannel(conn *websocket.Conn) (error, int) {
	var err error
	c.bt.Run(func(shtdwnCtx context.Context) {
		for {
			select {
			case <-shtdwnCtx.Done():
				conn.Close(websocket.StatusNormalClosure, "client exited letschat")
				return
			default:
				var msg domain.Message
				if er := wsjson.Read(context.Background(), conn, &msg); err != nil {
					if websocket.CloseStatus(err) == websocket.StatusNormalClosure {
						return
					}
					err = er
					return
				}
				c.messages <- &msg
			}
		}
	})
	return err, int(websocket.CloseStatus(err))
}
