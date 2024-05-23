package api

import (
	"encoding/json"
	"net/http"
	"strconv"
)

func (s *Server) updatePendingInvoiceAdjustment(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Status string `json:"status"`
	}

	ledgerId, _ := strconv.Atoi(r.PathValue("ledgerId"))

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return
	}

	err := s.Service.UpdatePendingInvoiceAdjustment(ledgerId, body.Status)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
