package server

import (
	"github.com/M0hammadUsman/letschat/internal/api/utility"
	"github.com/M0hammadUsman/letschat/internal/domain"
	"net/http"
)

func (s *Server) GetPagedMessageHandler(w http.ResponseWriter, r *http.Request) {
	u := utility.ContextGetUser(r.Context())
	u, ok := s.Subscribers[u.ID]
	if !ok {
		s.noSubscriptionResponse(w, r)
		return
	}
	var ev *domain.ErrValidation
	filter, ev := s.populateFilter(w, r)
	if ev != nil {
		s.failedValidationResponse(w, r, ev.Errors)
	}
	metadata, err := s.Facade.WritePagedMessagesToWSConn(r.Context(), u.Messages, filter)
	if err != nil {
		s.serverErrorResponse(w, r, err)
		return
	}
	msg := "retrieved paged messages has been written to websocket connection"
	if err = s.writeJSON(w, envelop{"info": msg, "metadata": metadata}, http.StatusOK, nil); err != nil {
		s.serverErrorResponse(w, r, err)
	}
}

func (s *Server) populateFilter(w http.ResponseWriter, r *http.Request) (*domain.Filter, *domain.ErrValidation) {
	var filter domain.Filter
	v := r.URL.Query()
	ev := domain.NewErrValidation()
	filter.Page = s.readInt(v, "page", 1, ev)
	filter.PageSize = s.readInt(v, "size", 25, ev)
	if ev.HasErrors() {
		s.failedValidationResponse(w, r, ev.Errors)
		return nil, ev
	}
	domain.ValidateFilters(ev, &filter)
	if ev.HasErrors() {
		s.failedValidationResponse(w, r, ev.Errors)
		return nil, ev
	}
	return &filter, nil
}
