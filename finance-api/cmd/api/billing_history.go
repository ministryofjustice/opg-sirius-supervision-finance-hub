package api

import (
	"encoding/json"
	"net/http"
	"strconv"
)

func (s *Server) getBillingHistory(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()

	clientId, _ := strconv.Atoi(r.PathValue("clientId"))
	billingHistory, err := s.Service.GetBillingHistory(ctx, clientId)

	if err != nil {
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	return json.NewEncoder(w).Encode(billingHistory)
}
