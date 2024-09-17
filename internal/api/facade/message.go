package facade

import (
	"context"
	"github.com/M0hammadUsman/letschat/internal/api/service"
	"github.com/M0hammadUsman/letschat/internal/common"
	"github.com/M0hammadUsman/letschat/internal/domain"
	"log/slog"
)

type MessageFacade struct {
	service   *service.Service
	txManager TXManager
	bgTask    *common.BackgroundTask
}

func NewMessageFacade(service *service.Service,
	txMan TXManager,
	bgTask *common.BackgroundTask) *MessageFacade {
	return &MessageFacade{
		service:   service,
		txManager: txMan,
		bgTask:    bgTask,
	}
}

func (f *MessageFacade) ProcessSentMessage(ctx context.Context,
	m domain.MessageSent,
	u *domain.User,
) (*domain.Message, error) {
	if ev := m.ValidateMessageSent(); ev != nil && ev.HasErrors() {
		return nil, ev
	}
	msg := f.service.PopulateMessage(m, u)
	if msg.Operation == domain.CreateMsg {
		convoExists, err := f.service.ConversationExists(ctx, msg.SenderID, m.ReceiverID)
		if err != nil {
			return nil, err
		}
		if !convoExists {
			if err = f.service.CreateConversation(ctx, msg.SenderID, m.ReceiverID); err != nil {
				return nil, err
			}
		}
	}
	f.processMessage(ctx, msg)
	return msg, nil
}

func (f *MessageFacade) WriteUnDeliveredMessagesToWSConn(ctx context.Context, c domain.MsgChan) error {
	return f.service.GetUnDeliveredMessages(ctx, c)
}

func (f *MessageFacade) WritePagedMessagesToWSConn(
	ctx context.Context,
	c domain.MsgChan,
	filter *domain.Filter,
) (*domain.Metadata, error) {
	return f.service.GetMessagesAsPage(ctx, c, filter)
}

// Helpers & Stuff ----------------------------------------------------------------------------------------------------

func (f *MessageFacade) processMessage(ctx context.Context, msg *domain.Message) {
	f.bgTask.Run(func(context.Context) {
		var err error
		for range 5 { // retries 5 times
			err = f.service.ProcessSentMessages(ctx, msg)
			if err == nil {
				break
			}
		}
		if err != nil {
			slog.Error(err.Error())
		}
	})
}
