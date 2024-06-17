package api

import (
	"encoding/json"
	"net/http"
	"strconv"
)

func (s *Server) getPermittedAdjustments(w http.ResponseWriter, r *http.Request) {
	invoiceId, _ := strconv.Atoi(r.PathValue("invoiceId"))
	types, err := s.Service.GetPermittedAdjustments(invoiceId)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	jsonData, err := json.Marshal(types)
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
