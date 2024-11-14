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
		SenderID:   sndr.ID,
		ReceiverID: m.ReceiverID,
		SentAt:     m.SentAt,
		Operation:  m.Operation,
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
	switch m.Operation {

	case domain.CreateMsg:
		return s.messageRepo.InsertMessage(ctx, m)

	case domain.DeliveredMsg:
		if err := s.messageRepo.DeleteMessageWithOperation(ctx, m.ID, domain.CreateMsg); err != nil {
			return err
		}
		return s.messageRepo.InsertMessage(ctx, m)

	case domain.DeliveredConfirmMsg:
		return s.messageRepo.DeleteMessageWithOperation(ctx, m.ID, domain.DeliveredMsg)

	case domain.ReadMsg:
		// if read, delete delivered one if exists
		if err := s.messageRepo.DeleteMessageWithOperation(ctx, m.ID, domain.DeliveredMsg); err != nil {
			return err
		}
		return s.messageRepo.InsertMessage(ctx, m)

	case domain.ReadConfirmMsg:
		return s.messageRepo.DeleteMessageWithOperation(ctx, m.ID, domain.ReadMsg)

	case domain.DeleteMsg:
		// delete all msgs with createMsg, deliveredMsg, readMsg OP's for specific id
		if err := s.messageRepo.DeleteMessage(ctx, m.ID); err != nil {
			return err
		}
		return s.messageRepo.InsertMessage(ctx, m)

	case domain.DeleteConfirmMsg:
		return s.messageRepo.DeleteMessageWithOperation(ctx, m.ID, domain.DeleteMsg)

	// these cases will be processed directly if the appropriate party(sender/receiver) is online
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
