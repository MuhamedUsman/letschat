package service

import (
	"context"
	"github.com/M0hammadUsman/letschat/internal/api/utility"
	"github.com/M0hammadUsman/letschat/internal/domain"
)

// also include delete in here

type ConversationService struct {
	conversationRepository domain.ConversationRepository
}

func NewConversationService(cr domain.ConversationRepository) *ConversationService {
	return &ConversationService{conversationRepository: cr}
}

func (s *ConversationService) CreateConversation(ctx context.Context, senderID, receiverID string) (bool, error) {
	return s.conversationRepository.CreateConversation(ctx, senderID, receiverID)
}

func (s *ConversationService) GetConversations(ctx context.Context) ([]*domain.Conversation, error) {
	usr := utility.ContextGetUser(ctx)
	if usr == nil {
		panic("no user was found in the context, Hint: missing Authentication middleware")
	}
	return s.conversationRepository.GetConversations(ctx, usr.ID)
}

func (s *ConversationService) ConversationExists(ctx context.Context, senderID, receiverID string) (bool, error) {
	return s.conversationRepository.ConversationExists(ctx, senderID, receiverID)
}
