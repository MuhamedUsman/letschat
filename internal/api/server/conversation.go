package server

import "net/http"

func (s *Server) GetConversationsHandler(w http.ResponseWriter, r *http.Request) {
	c, err := s.Facade.GetConversations(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
	if err = s.writeJSON(w, envelop{"conversations": c}, http.StatusOK, nil); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}
