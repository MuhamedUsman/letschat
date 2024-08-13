package main

import (
	"github.com/M0hammadUsman/letschat/internal/common"
	"github.com/M0hammadUsman/letschat/internal/facade"
	"github.com/M0hammadUsman/letschat/internal/mailer"
	"github.com/M0hammadUsman/letschat/internal/repository"
	"github.com/M0hammadUsman/letschat/internal/server"
	"github.com/M0hammadUsman/letschat/internal/service"
)

func main() {
	common.ConfigureSlog()
	cfg := common.ParseFlags()
	// Base
	db := repository.OpenDB(cfg)
	bgTask := common.NewBackgroundTask()
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
