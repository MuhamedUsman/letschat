package server

import (
	"github.com/justinas/alice"
	"net/http"
)

func (s *Server) routes() http.Handler {
	mux := http.NewServeMux()
	// Middlewares
	base := alice.New(s.recoverPanic, s.authenticate)
	authenticated := alice.New(s.requireAuthenticatedUser)
	protected := authenticated.Append(s.requireActivatedUser)
	// User Routes
	mux.HandleFunc("POST /v1/users", s.RegisterUserHandler)
	mux.Handle("GET /v1/users/{field}", authenticated.ThenFunc(s.GetByUniqueFieldHandler))
	mux.Handle("GET /v1/users", authenticated.ThenFunc(s.SearchUserHandler))
	mux.Handle("GET /v1/users/current", protected.ThenFunc(s.GetCurrentActiveUserHandler))
	mux.Handle("PUT /v1/users", protected.ThenFunc(s.UpdateUserHandler))
	mux.HandleFunc("POST /v1/users/activate", s.ActivateUserHandler)
	// Token Routes
	mux.HandleFunc("POST /v1/tokens/otp", s.GenerateOTPHandler)
	mux.HandleFunc("POST /v1/tokens/auth", s.GenerateAuthTokenHandler)
	// Conversation Routes
	mux.Handle("GET /v1/conversations", protected.ThenFunc(s.GetConversationsHandler))
	// Websocket Routes
	mux.Handle("/sub", protected.ThenFunc(s.WebsocketSubscribeHandler))

	return base.Then(mux)
}
