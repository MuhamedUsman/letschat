package facade

import (
	"context"
	"github.com/MuhamedUsman/letschat/internal/api/service"
	"github.com/MuhamedUsman/letschat/internal/domain"
)

type ConversationFacade struct {
	service *service.Service
}

func NewConversationFacade(srv *service.Service) *ConversationFacade {
	return &ConversationFacade{srv}
}

func (f *ConversationFacade) GetConversations(ctx context.Context) ([]*domain.Conversation, error) {
	return f.service.GetConversations(ctx)
}
