package server

import (
	"github.com/M0hammadUsman/letschat/internal/api/common"
	"github.com/M0hammadUsman/letschat/internal/domain"
	"net/http"
	"strings"
)

func (s *Server) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Vary", "Authorization")
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			r = common.ContextSetUser(r, domain.AnonymousUser)
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
		r = common.ContextSetUser(r, usr)
		next.ServeHTTP(w, r)
	})
}

func (s *Server) requireAuthenticatedUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		usr := common.ContextGetUser(r.Context())
		if usr.IsAnonymousUser() {
			s.authenticationRequiredResponse(w, r)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (s *Server) requireActivatedUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		usr := common.ContextGetUser(r.Context())
		if !usr.Activated {
			s.inactiveAccountResponse(w, r)
			return
		}
		next.ServeHTTP(w, r)
	})
}
