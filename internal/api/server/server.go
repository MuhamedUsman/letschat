package server

import (
	"fmt"
	"github.com/M0hammadUsman/letschat/internal/api/facade"
	"github.com/M0hammadUsman/letschat/internal/api/utility"
	"github.com/M0hammadUsman/letschat/internal/common"
	"github.com/M0hammadUsman/letschat/internal/domain"
	"github.com/coder/websocket"
	"golang.org/x/time/rate"

	"log"
	"log/slog"
	"net/http"
	"sync"
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

func (s *Server) Serve() {
	srv := &http.Server{
		Addr:         fmt.Sprint(":", s.Config.Port),
		Handler:      s.routes(),
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 6 * time.Second,
		IdleTimeout:  time.Minute,
	}
	slog.Info("starting server", "addr", srv.Addr)
	log.Fatal(srv.ListenAndServe())
}
