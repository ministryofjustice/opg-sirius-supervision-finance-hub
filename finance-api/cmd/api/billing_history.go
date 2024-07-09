package api

import (
	"encoding/json"
	"net/http"
	"strconv"
)

func (s *Server) getBillingHistory(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	clientId, _ := strconv.Atoi(r.PathValue("clientId"))
	billingHistory, err := s.Service.GetBillingHistory(ctx, clientId)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	jsonData, err := json.Marshal(billingHistory)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(jsonData)
	if err != nil {
		return
	}
}
