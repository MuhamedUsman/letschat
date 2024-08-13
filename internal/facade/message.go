package facade

import (
	"context"
	"github.com/M0hammadUsman/letschat/internal/common"
	"github.com/M0hammadUsman/letschat/internal/domain"
	"github.com/M0hammadUsman/letschat/internal/service"
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
	m domain.MessageSent) (*domain.Message, *domain.ErrValidation) {
	if ev := m.ValidateMessageSent(); ev != nil && ev.HasErrors() {
		return nil, ev
	}
	msg := f.service.PopulateMessage(ctx, m)
	f.bgTask.Run(func(context.Context) {
		var err error
		for range 5 { // retries 5 times
			err = f.service.ProcessSentMessages(ctx, msg)
		}
		if err != nil {
			slog.Debug(err.Error())
		}
	})
	return msg, nil
}

func (f *MessageFacade) WriteUnreadMessagesToWSConn(ctx context.Context, c domain.MsgChan) error {
	return f.service.GetUnreadMessages(ctx, c)
}

func (f *MessageFacade) WritePagedMessagesToWSConn(
	ctx context.Context,
	c domain.MsgChan,
	filter *domain.Filter,
) (*domain.Metadata, error) {
	return f.service.GetMessagesAsPage(ctx, c, filter)
}
