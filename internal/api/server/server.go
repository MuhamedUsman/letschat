package server

import (
	"context"
	"errors"
	"fmt"
	"github.com/M0hammadUsman/letschat/internal/api/facade"
	"github.com/M0hammadUsman/letschat/internal/api/utility"
	"github.com/M0hammadUsman/letschat/internal/common"
	"github.com/M0hammadUsman/letschat/internal/domain"
	"github.com/coder/websocket"
	"golang.org/x/time/rate"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

type Server struct {
	Config                  *utility.Config
	BackgroundTask          *common.BackgroundTask
	Facade                  *facade.Facade
	wsAcceptOpts            *websocket.AcceptOptions
	subscriberMessageBuffer int
	publishLimiter          *rate.Limiter

	SubsMu      sync.Mutex
	Subscribers map[string]*domain.User
}

func NewServer(cfg *utility.Config, bt *common.BackgroundTask, facade *facade.Facade) *Server {
	return &Server{
		Config:         cfg,
		BackgroundTask: bt,
		Facade:         facade,
		wsAcceptOpts: &websocket.AcceptOptions{
			CompressionMode:    websocket.CompressionContextTakeover,
			InsecureSkipVerify: true,
		},
		subscriberMessageBuffer: 16,
		publishLimiter:          rate.NewLimiter(rate.Limit(100*time.Millisecond), 10),
		Subscribers:             make(map[string]*domain.User), // keys are userID
	}
}

func (s *Server) Serve() error {
	srv := &http.Server{
		Addr:         fmt.Sprint(":", s.Config.Port),
		Handler:      s.routes(),
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 6 * time.Second,
		IdleTimeout:  time.Minute,
	}
	shutdownErr := make(chan error)
	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
		sig := <-quit
		slog.Info("shutting down server", "signal", sig.String())
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := srv.Shutdown(ctx); err != nil {
			slog.Error(err.Error())
			shutdownErr <- err
		} else {
			shutdownErr <- nil
		}
	}()
	slog.Info("starting server", "addr", srv.Addr)
	err := srv.ListenAndServe()
	if !errors.Is(err, http.ErrServerClosed) {
		return err
	} else {
		slog.Info("waiting for ongoing http requests", "max wait", "5 sec...")
	}
	if err = <-shutdownErr; err != nil {
		return err
	}
	slog.Info("server down, waiting for background tasks to gracefully shutdown",
		"tasks", s.BackgroundTask.Tasks,
		"max wait", "5 sec...")
	if err = s.BackgroundTask.Shutdown(6 * time.Second); err != nil {
		slog.Warn(err.Error())
	}
	slog.Info("stopped server", "addr", srv.Addr)
	return nil
}

func (s *Server) ShutdownCleanup() {
	s.BackgroundTask.Run(func(shtdwnCtx context.Context) {
		<-shtdwnCtx.Done()
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		for range 5 { // 5 reties if something gets wrong
			if err := s.Facade.SetOnlineUsersLastSeen(ctx); err == nil {
				break
			}
		}
	})
}
