package api

import (
	"net/http"
)

func (s *Server) cancelDirectDebit(w http.ResponseWriter, r *http.Request) error {
	//call all pay here

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	return nil
}
