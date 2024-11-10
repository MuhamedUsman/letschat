package service

import (
	"context"
	"fmt"
	"github.com/M0hammadUsman/letschat/internal/api/utility"
	"github.com/M0hammadUsman/letschat/internal/domain"
	"github.com/google/uuid"
)

type MessageService struct {
	messageRepo domain.MessageRepository
}

func NewMessageService(messageRepo domain.MessageRepository) *MessageService {
	return &MessageService{messageRepo}
}

func (*MessageService) PopulateMessage(m domain.MessageSent, sndr *domain.User) *domain.Message {
	msg := &domain.Message{
		SenderID:    sndr.ID,
		ReceiverID:  m.ReceiverID,
		SentAt:      m.SentAt,
		DeliveredAt: m.DeliveredAt,
		ReadAt:      m.ReadAt,
		Operation:   m.Operation,
	}
	if m.ID != nil {
		msg.ID = *m.ID
	} else {
		msg.ID = uuid.New().String()
	}
	if m.Body != nil {
		msg.Body = *m.Body
	}
	return msg
}

func (s *MessageService) ProcessSentMessages(ctx context.Context, m *domain.Message) error {
	switch m.Operation {
	case domain.CreateMsg:
		return s.messageRepo.InsertMessage(ctx, m)
	case domain.UpdateMsg:
		var msg *domain.Message
		var err error
		for i := range 5 { // the msg may not be there, due to delete and updates
			msg, err = s.messageRepo.GetByID(ctx, m.ID)
			if i == 4 && err != nil {
				return fmt.Errorf("get by id, while updating msg, id=\"%v\" %v", m.ID, err)
			}
			break
		}
		if m.DeliveredAt != nil {
			msg.DeliveredAt = m.DeliveredAt
		}
		if m.ReadAt != nil {
			msg.ReadAt = m.ReadAt
		}
		return s.messageRepo.UpdateMessage(ctx, msg)
	case domain.DeleteMsg:
		return s.messageRepo.DeleteMessage(ctx, m.ID)
	case domain.UserOnlineMsg:
		return nil
	case domain.UserOfflineMsg:
		return nil
	case domain.UserTypingMsg:
		return nil
	default:
		return fmt.Errorf("unknown operation %v", m.Operation)
	}
}

func (s *MessageService) GetUnDeliveredMessages(ctx context.Context, c domain.MsgChan) error {
	u := utility.ContextGetUser(ctx)
	return s.messageRepo.GetUnDeliveredMessages(ctx, u.ID, c)
}

func (s *MessageService) GetMessagesAsPage(
	ctx context.Context,
	c domain.MsgChan,
	filter *domain.Filter,
) (*domain.Metadata, error) {
	u := utility.ContextGetUser(ctx)
	return s.messageRepo.GetMessagesAsPage(ctx, u.ID, c, filter)
}

func (s *MessageService) SaveMessage(ctx context.Context, m *domain.Message) error {
	return s.messageRepo.InsertMessage(ctx, m)
}

func (s *MessageService) UpdateMessage(ctx context.Context, m *domain.Message) error {
	msg, err := s.messageRepo.GetByID(ctx, m.ID)
	if err != nil {
		return err
	}
	m.Version = msg.Version
	return s.messageRepo.UpdateMessage(ctx, m)
}

func (s *MessageService) DeleteMessage(ctx context.Context, mID string) error {
	return s.messageRepo.DeleteMessage(ctx, mID)
}
