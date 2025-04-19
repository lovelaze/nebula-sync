package api

import (
	"net/http"
)

func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
	if s.healthy() {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func (s *Server) healthy() bool {
	return len(s.state.Stack) > 0 && s.state.Stack[0].Success
}
