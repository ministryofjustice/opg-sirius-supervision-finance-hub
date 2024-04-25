package api

import (
	"encoding/json"
	"net/http"
	"strconv"
)

func (s *Server) getInvoiceAdjustments(w http.ResponseWriter, r *http.Request) {
	clientId, _ := strconv.Atoi(r.PathValue("id"))
	data, err := s.Service.GetInvoiceAdjustments(clientId)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	jsonData, err := json.Marshal(data)
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
