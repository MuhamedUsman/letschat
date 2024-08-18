package main

import (
	common2 "github.com/M0hammadUsman/letschat/internal/api/common"
	"github.com/M0hammadUsman/letschat/internal/api/facade"
	"github.com/M0hammadUsman/letschat/internal/api/mailer"
	"github.com/M0hammadUsman/letschat/internal/api/repository"
	"github.com/M0hammadUsman/letschat/internal/api/server"
	"github.com/M0hammadUsman/letschat/internal/api/service"
)

func main() {
	common2.ConfigureSlog()
	cfg := common2.ParseFlags()
	// Base
	db := repository.OpenDB(cfg)
	bgTask := common2.NewBackgroundTask()
	mailr := mailer.New(cfg)
	// Repositories
	userRepo := repository.NewUserRepository(db)
	tokenRepo := repository.NewTokenRepository(db)
	messageRepo := repository.NewMessageRepository(db)
	// Services
	userService := service.NewUserService(userRepo)
	tokenService := service.NewTokenService(tokenRepo)
	messageService := service.NewMessageService(messageRepo)
	// Service Group
	srv := service.New(userService, tokenService, messageService)
	// Facades
	userFacade := facade.NewUserFacade(srv, db, mailr, bgTask)
	tokenFacade := facade.NewTokenFacade(srv, db, mailr, bgTask)
	messageFacade := facade.NewMessageFacade(srv, db, bgTask)
	// Facade Group
	fac := facade.New(userFacade, tokenFacade, messageFacade)
	// Server
	s := server.NewServer(cfg, bgTask, fac)
	s.Serve()
}
