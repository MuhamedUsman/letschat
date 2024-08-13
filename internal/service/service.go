package service

import (
	"github.com/M0hammadUsman/letschat/internal/domain"
)

type Service struct {
	domain.UserService
	domain.TokenService
	domain.MessageService
}

func New(us domain.UserService, ts domain.TokenService, ms domain.MessageService) *Service {
	return &Service{
		UserService:    us,
		TokenService:   ts,
		MessageService: ms,
	}
}
