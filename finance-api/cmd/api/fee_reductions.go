package api

import (
	"encoding/json"
	"net/http"
	"strconv"
)

func (s *Server) getFeeReductions(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()

	clientId, _ := strconv.Atoi(r.PathValue("clientId"))
	accountInfo, err := s.service.GetFeeReductions(ctx, clientId)
	if err != nil {
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	return json.NewEncoder(w).Encode(accountInfo)
}
