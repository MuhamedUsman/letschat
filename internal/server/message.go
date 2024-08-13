package server

import (
	"github.com/M0hammadUsman/letschat/internal/common"
	"github.com/M0hammadUsman/letschat/internal/domain"
	"net/http"
)

func (s *Server) GetPagedMessageHandler(w http.ResponseWriter, r *http.Request) {
	u := common.ContextGetUser(r.Context())
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
	var query struct {
		Page     int `json:"page"`
		PageSize int `json:"pageSize"`
	}
	v := r.URL.Query()
	ev := domain.NewErrValidation()
	query.Page = s.readInt(v, "page", 1, ev)
	query.PageSize = s.readInt(v, "pageSize", 25, ev)
	if ev.HasErrors() {
		s.failedValidationResponse(w, r, ev.Errors)
		return nil, ev
	}
	filter := &domain.Filter{
		Page:     query.Page,
		PageSize: query.PageSize,
	}
	domain.ValidateFilters(ev, filter)
	if ev.HasErrors() {
		s.failedValidationResponse(w, r, ev.Errors)
		return nil, ev
	}
	return filter, nil
}
