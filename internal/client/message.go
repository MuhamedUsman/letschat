package client

import (
	"context"
	"errors"
	"github.com/M0hammadUsman/letschat/internal/domain"
	"github.com/M0hammadUsman/letschat/internal/sync"
	"github.com/google/uuid"
	"log/slog"
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

func (c *Client) SendMessage(msg string, recvrID string) error {
	// first save in db
	m := new(domain.Message)
	m.ID = uuid.New().String()
	m.SenderID = c.CurrentUsr.ID
	m.ReceiverID = recvrID
	m.Body = msg
	m.SentAt = ptr(time.Now().UTC())
	m.Operation = domain.CreateMsg
	c.sentMsgs.msgs <- m // this will send the msg
	// if it's not sent save it to db without the sentAt field
	if !<-c.sentMsgs.done {
		m.SentAt = nil
		if err := c.repo.SaveMsg(m); err != nil {
			return err
		}
		// before returning write the conversations with updated last msgs to chan, tui.ConversationModel will pick it
		convos, _ := c.repo.GetConversations()
		c.populateConvosAndWriteToChan(convos)
		return ErrMsgNotSent
	}
	// if sent save with sentAt field
	if err := c.repo.SaveMsg(m); err != nil {
		return err
	}
	convos, _ := c.repo.GetConversations()
	c.populateConvosAndWriteToChan(convos)
	return nil
}

func (c *Client) GetMessagesAsPage(senderID string, page int) ([]*domain.Message, *domain.Metadata, error) {
	f := domain.Filter{
		Page:     page,
		PageSize: 25,
	}
	msgs, metadata, err := c.repo.GetMsgsAsPage(senderID, f)
	if err != nil {
		return nil, nil, err
	}
	return msgs, metadata, nil
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
			err := c.repo.SaveMsg(msg)
			if err != nil {
				slog.Error(err.Error())
			}
			// TODO: send back message as delivered

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
		// once there is a message we also update the conversations as the latest msg will also need update
		convos := c.Conversations.GetAndBlock()
		c.Conversations.Unblock()
		c.populateConvosAndWriteToChan(convos)
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
