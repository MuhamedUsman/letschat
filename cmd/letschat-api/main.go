package main

import (
	"fmt"
	"github.com/MuhamedUsman/letschat/internal/api/facade"
	"github.com/MuhamedUsman/letschat/internal/api/mailer"
	"github.com/MuhamedUsman/letschat/internal/api/repository"
	"github.com/MuhamedUsman/letschat/internal/api/server"
	"github.com/MuhamedUsman/letschat/internal/api/service"
	"github.com/MuhamedUsman/letschat/internal/api/utility"
	"github.com/MuhamedUsman/letschat/internal/common"
	"log/slog"
	"os"
)

func main() {
	utility.ConfigureSlog(os.Stderr)
	cfg := utility.ParseFlags()
	// Base
	db := repository.OpenDB(cfg)
	bgTask := common.NewBackgroundTask()
	mailr := mailer.New(cfg)
	// Repositories
	userRepo := repository.NewUserRepository(db)
	tokenRepo := repository.NewTokenRepository(db)
	messageRepo := repository.NewMessageRepository(db)
	conversationRepo := repository.NewConversationRepository(db)
	// Services
	userService := service.NewUserService(userRepo)
	tokenService := service.NewTokenService(tokenRepo)
	messageService := service.NewMessageService(messageRepo)
	conversationService := service.NewConversationService(conversationRepo)
	// Service Group
	srv := service.New(userService, tokenService, messageService, conversationService)
	// Facades
	userFacade := facade.NewUserFacade(srv, db, mailr, bgTask)
	tokenFacade := facade.NewTokenFacade(srv, db, mailr, bgTask)
	messageFacade := facade.NewMessageFacade(srv, db, bgTask)
	conversationFacade := facade.NewConversationFacade(srv)
	// Facade Group
	fac := facade.New(userFacade, tokenFacade, messageFacade, conversationFacade)
	// Server
	s := server.NewServer(cfg, bgTask, fac)
	// printing banner
	fmt.Println("    __         __            __          __ \n   / /   ___  / /___________/ /_  ____ _/ /_\n  / /   / _ \\/ __/ ___/ ___/ __ \\/ __ `/ __/\n / /___/  __/ /_(__  ) /__/ / / / /_/ / /_  \n/_____/\\___/\\__/____/\\___/_/ /_/\\__,_/\\__/  \n                                            ")
	// Starting Server and setting up cleanup processes
	s.ShutdownCleanup() // will run once the server shutdown initiates
	if err := s.Serve(); err != nil {
		slog.Error(err.Error())
	}
}
