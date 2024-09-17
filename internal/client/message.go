package client

import (
	"context"
	"errors"
	"github.com/M0hammadUsman/letschat/internal/client/sync"
	"github.com/M0hammadUsman/letschat/internal/domain"
	"github.com/google/uuid"
	"time"
)

var (
	ErrMsgNotSent = errors.New("message not sent, 2 sec ctx timed out")
)

type sentMsgs struct {
	msgs chan<- *domain.Message
	// once we send a message we read on this chan to ensure that the message is sent to the server
	done <-chan bool
}

type RecvMsgsMonitor = sync.StateMonitor[*domain.Message]

func newRecvMsgsStateMonitor() *RecvMsgsMonitor {
	return sync.NewStatus[*domain.Message](nil)
}

func (c *Client) SendMessage(msg string, recvrID string) (*domain.Message, error) {
	// first save in db
	m := new(domain.Message)
	m.ID = uuid.New().String()
	m.SenderID = c.CurrentUsr.ID
	m.ReceiverID = recvrID
	m.Body = msg
	m.SentAt = ptr(time.Now())
	m.Operation = domain.CreateMsg
	c.sentMsgs.msgs <- m // this will send the msg
	// if it's not sent save it to db without the sentAt field
	if !<-c.sentMsgs.done {
		m.SentAt = nil
		if err := c.repo.SaveMsg(m); err != nil {
			return nil, err
		}
		// before returning write the conversations with updated last msgs to chan, tui.ConversationModel will pick it
		c.writeUpdatedConvosToChan()
		return m, ErrMsgNotSent
	}
	// if sent save with sentAt field
	if err := c.repo.SaveMsg(m); err != nil {
		return nil, err
	}
	c.writeUpdatedConvosToChan()
	return m, nil
}

func (c *Client) handleReceivedMsgs(shtdwnCtx context.Context) {
	for {
		msg := c.RecvMsgs.WaitForStateChange() // once there is a new message we'll get that, util then it'll block
		select {
		case <-shtdwnCtx.Done():
			return
		default:
		}
		switch msg.Operation {

		case domain.CreateMsg:
			msg.DeliveredAt = ptr(time.Now())
			_ = c.repo.SaveMsg(msg)

		case domain.UpdateMsg:
			msgToUpdate, err := c.repo.GetMsgByID(msg.ID)
			if err != nil { // we've not found the msg in the user's local repo so there is noting to update
				continue
			}
			if msg.DeliveredAt != nil {
				msgToUpdate.DeliveredAt = msg.DeliveredAt
			}
			if msg.ReadAt != nil {
				msgToUpdate.ReadAt = msg.ReadAt
			}
			for range 5 { // retries for 5 times, in case there is domain.ErrEditConflict
				if err = c.repo.UpdateMsg(msgToUpdate); err == nil {
					break
				}
			}

		case domain.DeleteMsg:
			_ = c.repo.DeleteMsg(msg.ID)

		case domain.UserOnlineMsg:
			c.setUsrOnlineStatus(shtdwnCtx, msg, true)

		case domain.UserOfflineMsg:
			c.setUsrOnlineStatus(shtdwnCtx, msg, false)
		}
	}
}

// Helpers & Stuff -----------------------------------------------------------------------------------------------------

func (c *Client) setUsrOnlineStatus(shtdwnCtx context.Context, msg *domain.Message, online bool) {
	convos := c.Conversations.GetAndBlock()
	lastOnline := msg.SentAt
	if !online {
		lastOnline = nil
	}
	for i := range convos {
		// offline/online user is in the convos
		if convos[i].UserID == msg.SenderID {
			convos[i].LastOnline = lastOnline
			break
		}
	}
	c.Conversations.WriteToChan(convos)
	c.Conversations.Unblock()
	c.Conversations.Broadcast(shtdwnCtx)
}

func ptr[T any](v T) *T {
	return &v
}
