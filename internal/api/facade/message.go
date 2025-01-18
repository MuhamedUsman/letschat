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
) (*domain.Message, bool, error) {
	if ev := m.ValidateMessageSent(); ev != nil && ev.HasErrors() {
		return nil, false, ev
	}
	msg := f.service.PopulateMessage(m, u)
	convoCreated := false
	if msg.Operation == domain.CreateMsg {
		convoExists, err := f.service.ConversationExists(ctx, msg.SenderID, m.ReceiverID)
		if err != nil {
			return nil, false, err
		}
		if !convoExists {
			convoCreated, err = f.service.CreateConversation(ctx, msg.SenderID, m.ReceiverID)
			if err != nil {
				return nil, convoCreated, err
			}
		}
	}
	f.processMessage(ctx, msg)
	return msg, convoCreated, nil
}

func (f *MessageFacade) WriteUnDeliveredMessagesToWSConn(ctx context.Context, c domain.MsgChan) error {
	return f.service.GetUnDeliveredMessages(ctx, c)
}

// Helpers & Stuff ----------------------------------------------------------------------------------------------------

func (f *MessageFacade) processMessage(ctx context.Context, msg *domain.Message) {
	f.bgTask.Run(func(context.Context) {
		if err := f.txManager.RunInTX(ctx, func(ctx context.Context) error {
			return f.service.ProcessSentMessages(ctx, msg)
		}); err != nil {
			slog.Error(err.Error())
		}

	})
}
