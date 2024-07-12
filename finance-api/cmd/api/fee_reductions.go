package api

import (
	"encoding/json"
	"net/http"
	"strconv"
)

func (s *Server) getFeeReductions(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	clientId, _ := strconv.Atoi(r.PathValue("clientId"))
	accountInfo, err := s.Service.GetFeeReductions(ctx, clientId)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	jsonData, err := json.Marshal(accountInfo)
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
