package service

import (
	"github.com/MuhamedUsman/letschat/internal/domain"
)

type Service struct {
	domain.UserService
	domain.TokenService
	domain.MessageService
	domain.ConversationService
}

func New(us domain.UserService,
	ts domain.TokenService,
	ms domain.MessageService,
	cs domain.ConversationService) *Service {
	return &Service{
		UserService:         us,
		TokenService:        ts,
		MessageService:      ms,
		ConversationService: cs,
	}
}
