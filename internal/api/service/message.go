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
	} else if msg.Operation == domain.CreateMsg {
		msg.ID = uuid.New().String()
	} else {
		panic("msg.Operation != domain.CreateMsg, yet ID is nil, Hint: failing/bad validation")
	}
	if m.Body != nil {
		msg.Body = *m.Body
	}
	return msg
}

func (s *MessageService) ProcessSentMessages(ctx context.Context, m *domain.Message) error {
	// the Primary key constraint ensures, at any instant there only exists 1 msg for a specific msgId, so delete that
	// msg without specifying the operation
	switch m.Operation {

	case domain.CreateMsg:
		return s.messageRepo.InsertMessage(ctx, m)

	// these OPs cases will delete msgs with specified Ops, CreateMsg, DeliveredMsg, Any Op
	case domain.DeliveredMsg, domain.ReadMsg, domain.DeleteMsg:
		if m.Operation == domain.ReadMsg {
			msg, _ := s.messageRepo.GetByID(ctx, m.ID, domain.DeliveredMsg)
			if msg != nil {
				// if the sender is offline and the msg is delivered & read in that case, also persist deliveredAt field
				m.DeliveredAt = msg.DeliveredAt
			}
		}
		if err := s.messageRepo.DeleteMessage(ctx, m.ID); err != nil {
			return err
		}
		return s.messageRepo.InsertMessage(ctx, m)

	// these OPs are not for persistence, but merely a confirmation to ensure robustness
	// these OPs cases will delete msgs with specified Ops, DeliveredMsg, ReadMsg, DeleteMsg
	case domain.DeliveredConfirmMsg, domain.ReadConfirmMsg, domain.DeleteConfirmMsg:
		return s.messageRepo.DeleteMessage(ctx, m.ID)

	// these Ops will be processed directly if the appropriate party(sender/receiver) is online
	case domain.OnlineMsg, domain.OfflineMsg, domain.TypingMsg:
		return nil

	default:
		return fmt.Errorf("unknown operation %v", m.Operation)
	}
}

func (s *MessageService) GetUnDeliveredMessages(ctx context.Context, c domain.MsgChan) error {
	u := utility.ContextGetUser(ctx)
	// the order matters here
	ops := []domain.MsgOperation{domain.DeleteMsg, domain.DeliveredMsg, domain.ReadMsg, domain.CreateMsg}
	for _, op := range ops {
		// this directly writes to the msg chan
		if err := s.messageRepo.GetUnDeliveredMessages(ctx, u.ID, op, c); err != nil {
			return err
		}
	}
	return nil
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
