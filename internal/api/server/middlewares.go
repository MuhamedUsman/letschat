package server

import (
	"fmt"
	"github.com/M0hammadUsman/letschat/internal/api/utility"
	"github.com/M0hammadUsman/letschat/internal/domain"
	"net/http"
	"strings"
)

func (s *Server) authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Vary", "Authorization")
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			r = utility.ContextSetUser(r, domain.AnonymousUser)
			next.ServeHTTP(w, r)
			return
		}
		authHeaderParts := strings.Split(authHeader, " ")
		if len(authHeaderParts) != 2 || authHeaderParts[0] != "Bearer" {
			s.invalidCredentialResponse(w, r)
			return
		}
		token := authHeaderParts[1]
		usr, err := s.Facade.VerifyAuthToken(r.Context(), token)
		if err != nil {
			s.invalidAuthenticationTokenResponse(w, r)
			return
		}
		r = utility.ContextSetUser(r, usr)
		next.ServeHTTP(w, r)
	})
}

func (s *Server) requireAuthenticatedUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		usr := utility.ContextGetUser(r.Context())
		if usr.IsAnonymousUser() {
			s.authenticationRequiredResponse(w, r)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (s *Server) requireActivatedUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		usr := utility.ContextGetUser(r.Context())
		if !usr.Activated {
			s.inactiveAccountResponse(w, r)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (s *Server) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				w.Header().Set("Connection", "close")
				s.serverErrorResponse(w, r, fmt.Errorf("%v", err))
			}
		}()
		next.ServeHTTP(w, r)
	})
}
