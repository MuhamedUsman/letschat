package server

import (
	"github.com/justinas/alice"
	"net/http"
)

func (s *Server) routes() http.Handler {
	mux := http.NewServeMux()
	// Middlewares
	base := alice.New(s.Authenticate)
	authenticated := base.Append(s.requireAuthenticatedUser)
	activated := authenticated.Append(s.requireActivatedUser)
	// User Routes
	mux.HandleFunc("POST /v1/users", s.RegisterUserHandler)
	mux.Handle("GET /v1/users/{field}", authenticated.ThenFunc(s.GetByUniqueFieldHandler))
	mux.Handle("PUT /v1/users", activated.ThenFunc(s.UpdateUserHandler))
	mux.HandleFunc("POST /v1/users/activate", s.ActivateUserHandler)
	// Token Routes
	mux.HandleFunc("POST /v1/tokens/otp", s.GenerateOTPHandler)
	mux.HandleFunc("POST /v1/tokens/auth", s.GenerateAuthTokenHandler)
	// Messages Routes
	mux.Handle("GET /v1/messages", activated.ThenFunc(s.GetPagedMessageHandler))
	// Websocket Routes
	mux.Handle("/sub", activated.ThenFunc(s.WebsocketSubscribeHandler))

	return base.Then(mux)
}
