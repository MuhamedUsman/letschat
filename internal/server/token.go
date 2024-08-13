package server

import (
	"errors"
	"github.com/M0hammadUsman/letschat/internal/domain"
	"net/http"
)

func (s *Server) GenerateOTPHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Email string `json:"email"`
	}
	if err := s.readJSON(w, r, &input); err != nil {
		s.badRequestResponse(w, r, err)
		return
	}
	if err := s.Facade.GenerateOTP(r.Context(), input.Email); err != nil {
		var ev *domain.ErrValidation
		switch {
		case errors.As(err, &ev):
			s.failedValidationResponse(w, r, ev.Errors)
		case errors.Is(err, domain.ErrAlreadyActive):
			s.alreadyActivatedResponse(w, r)
		default:
			s.serverErrorResponse(w, r, err)
		}
		return
	}
	w.WriteHeader(http.StatusAccepted)
}

func (s *Server) GenerateAuthTokenHandler(w http.ResponseWriter, r *http.Request) {
	var usr domain.UserAuth
	if err := s.readJSON(w, r, &usr); err != nil {
		s.badRequestResponse(w, r, err)
		return
	}
	token, err := s.Facade.GenerateAuthToken(r.Context(), &usr)
	if err != nil {
		var ev *domain.ErrValidation
		switch {
		case errors.As(err, &ev):
			s.failedValidationResponse(w, r, ev.Errors)
		default:
			s.serverErrorResponse(w, r, err)
		}
		return
	}
	if err = s.writeJSON(w, envelop{"token": token}, http.StatusOK, nil); err != nil {
		s.serverErrorResponse(w, r, err)
	}
}
