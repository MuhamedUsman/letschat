package client

import (
	"context"
	"errors"
	"fmt"
	"github.com/M0hammadUsman/letschat/internal/domain"
	"github.com/M0hammadUsman/letschat/internal/sync"
	"log/slog"
	"math"
	"time"
)

var (
	ErrMsgNotSent = errors.New("message not sent, 2 sec ctx timed out")
)

type sentMsgs struct {
	msgs chan<- *domain.Message
	// once we send a message, we read on this chan to ensure that the message is sent to the server
	done <-chan bool
}

type RecvMsgsBroadcaster = sync.Broadcaster[*domain.Message]

func newRecvMsgsBroadcaster() *RecvMsgsBroadcaster {
	return sync.NewBroadcaster[*domain.Message]()
}

func (c *Client) SendMessage(msg domain.Message) {
	c.sentMsgs.msgs <- &msg // this will send the msg
	// if it's not sent save it to db without the sentAt field
	// then, once we establish the conn back, we'll retry those
	if !<-c.sentMsgs.done {
		msg.SentAt = nil
		if err := c.repo.SaveMsg(&msg); err != nil {
			slog.Error(err.Error())
		}
		// before returning write the conversations with updated last msgs to chan, tui.ConversationModel will pick it
		convos, _ := c.repo.GetConversations()
		c.populateConvosWithLatestMsgs(convos)
		_ = c.repo.DeleteAllConversations()
		_ = c.repo.SaveConversations(convos...)
		c.Conversations.Write(convos)
	}
	// if sent save with sentAt field
	// and write it back to chan with sent state, so tui can update accordingly
	msg.Confirmation = domain.MsgDeliveredConfirmed
	if err := c.repo.SaveMsg(&msg); err != nil {
		slog.Error(err.Error())
	}
	msg.Operation = math.MinInt8 // so we don't have a redundant operation somewhere
	c.RecvMsgs.Write(&msg)
	convos, _ := c.repo.GetConversations()
	c.populateConvosWithLatestMsgs(convos)
	_ = c.repo.DeleteAllConversations()
	_ = c.repo.SaveConversations(convos...)
	c.Conversations.Write(convos)
}

func (c *Client) SendTypingStatus(msg domain.Message) {
	c.sentMsgs.msgs <- &msg
	if !<-c.sentMsgs.done {
		slog.Error(ErrMsgNotSent.Error())
	}
}

func (c *Client) GetMessagesAsPageAndMarkAsRead(senderID string, page int) ([]*domain.Message, *domain.Metadata, error) {
	f := domain.Filter{
		Page:     page,
		PageSize: 25,
	}
	msgs, metadata, err := c.repo.GetMsgsAsPage(senderID, f)
	if err != nil {
		return nil, nil, err
	}
	for _, msg := range msgs {
		if c.isValidReadUpdate(msg) {
			msg.ReadAt = ptr(time.Now())
		}
	}

	c.BT.Run(func(shtdwnCtx context.Context) {
		for _, msg := range msgs {
			if c.isValidReadUpdate(msg) {
				_ = c.SetMsgAsRead(msg) // Ignore & retry on reconnect
			}
		}
	})
	return msgs, metadata, nil
}

func (c *Client) handleReceivedMsgs(shtdwnCtx context.Context) {
	token, ch := c.RecvMsgs.Subscribe()
	defer c.RecvMsgs.Unsubscribe(token)
	for {
		select {
		case msg := <-ch:
			switch msg.Operation {

			case domain.CreateMsg:
				err := c.repo.SaveMsg(msg)
				if err != nil {
					slog.Error(err.Error())
				}
				if err = c.setMsgAsDelivered(msg.ID, msg.SenderID); err != nil {
					slog.Error(err.Error())
				}
				c.getPopulateSaveConvosAndWriteToChan()

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
				c.getPopulateSaveConvosAndWriteToChan()

			case domain.UserOnlineMsg:
				c.setUsrOnlineStatus(msg, true)

			case domain.UserOfflineMsg:
				c.setUsrOnlineStatus(msg, false)
			}

		case <-shtdwnCtx.Done():
			return
		}
	}
}

func (c *Client) SetMsgAsRead(msg *domain.Message) error {
	msgToSend := &domain.Message{
		ID:         msg.ID,
		SenderID:   c.CurrentUsr.ID,
		ReceiverID: msg.SenderID, // confirm that message is read
		ReadAt:     msg.ReadAt,   // if provided, use that
		Operation:  domain.UpdateMsg,
	}
	if msgToSend.ReadAt == nil {
		msgToSend.ReadAt = ptr(time.Now())
	}
	// this may block, in theory, depends on the connection
	c.sentMsgs.msgs <- msgToSend
	if !<-c.sentMsgs.done {
		// still update in local db with readAt, retry, once again when there is a connection established
		for i := range 5 {
			msgToUpdate, err := c.repo.GetMsgByID(msg.ID)
			if err != nil {
				return err
			}
			msgToUpdate.ReadAt = msg.ReadAt
			if err = c.repo.UpdateMsg(msgToUpdate); err != nil {
				if i == 4 {
					return err
				}
			} else {
				break
			}
		}
		return ErrMsgNotSent
	}
	// once ok update the local msg with MsgReadConfirmed
	for i := range 5 {
		msgToUpdate, err := c.repo.GetMsgByID(msg.ID)
		if err != nil {
			return err
		}
		msgToUpdate.ReadAt = msg.ReadAt
		msgToUpdate.Confirmation = domain.MsgReadConfirmed
		if err = c.repo.UpdateMsg(msgToUpdate); err != nil {
			if i == 4 {
				return err
			}
		} else {
			break
		}
	}
	return nil
}

func (c *Client) DeleteMsgForMe(msgId string) error {
	if err := c.repo.DeleteMsg(msgId); err != nil {
		return err
	}
	// update convos as the deleted msg may be the recent one
	c.getPopulateSaveConvosAndWriteToChan()
	return nil
}

func (c *Client) DeleteMsgForEveryone(msg *domain.Message) error {
	// this may block, in theory, depends on the connection
	c.sentMsgs.msgs <- msg
	if <-c.sentMsgs.done {
		if err := c.DeleteMsgForMe(msg.ID); err != nil {
			return err
		}
		c.getPopulateSaveConvosAndWriteToChan()
		return nil
	} else {
		return fmt.Errorf("ws conn closed due to error while deleting the message from the receiver")
	}
}

func (c *Client) DeleteForMeAllMsgsForConversation(senderId, receiverId string) error {
	err := c.repo.DeleteAllForSenderAndReceiver(senderId, receiverId)
	if err != nil {
		return err
	}
	c.getPopulateSaveConvosAndWriteToChan()
	return nil
}

// Helpers & Stuff -----------------------------------------------------------------------------------------------------

func (c *Client) setMsgAsDelivered(msgID, receiverID string) error {
	msg := &domain.Message{
		ID:          msgID,
		SenderID:    c.CurrentUsr.ID,
		ReceiverID:  receiverID,
		DeliveredAt: ptr(time.Now()),
		Operation:   domain.UpdateMsg,
	}
	c.sentMsgs.msgs <- msg
	// if msg is not sent
	if !<-c.sentMsgs.done {
		return ErrMsgNotSent
	}
	// update in local DB
	for i := range 5 { // this can yield domain.ErrEditConflict so, retry
		msgToUpdate, err := c.repo.GetMsgByID(msgID)
		if err != nil {
			return err
		}
		if msgToUpdate.ID == msg.ID {
			msgToUpdate.DeliveredAt = msg.DeliveredAt
		}
		msg.Confirmation = domain.MsgDeliveredConfirmed
		if err = c.repo.UpdateMsg(msgToUpdate); err != nil {
			if i == 4 {
				return err
			}
		} else {
			break
		}
	}
	return nil
}

func (c *Client) setUsrOnlineStatus(msg *domain.Message, online bool) {
	convos := c.Conversations.Get()
	lastOnline := msg.SentAt
	if online {
		lastOnline = nil
	}
	for i := range convos {
		// offline/online user is in the convos
		if convos[i].UserID == msg.SenderID {
			convos[i].LastOnline = lastOnline
			break
		}
	}
	c.Conversations.Write(convos)
}

func ptr[T any](v T) *T {
	return &v
}

func (c *Client) isValidReadUpdate(msg *domain.Message) bool {
	return msg.SenderID != c.CurrentUsr.ID && msg.DeliveredAt != nil && msg.Confirmation != domain.MsgReadConfirmed
}

// once there is a message we also update the conversations as the latest msg will also need update and save to db
func (c *Client) getPopulateSaveConvosAndWriteToChan() {
	convos := c.Conversations.Get()
	c.populateConvosWithLatestMsgs(convos)
	_ = c.repo.DeleteAllConversations()
	_ = c.repo.SaveConversations(convos...)
	c.Conversations.Write(convos)
}
