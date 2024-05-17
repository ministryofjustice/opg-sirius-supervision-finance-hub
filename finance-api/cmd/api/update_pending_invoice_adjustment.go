package api

import (
	"net/http"
	"strconv"
)

func (s *Server) updatePendingInvoiceAdjustment(w http.ResponseWriter, r *http.Request) {

	clientId, _ := strconv.Atoi(r.PathValue("id"))
	invoiceId, _ := strconv.Atoi(r.PathValue("invoiceId"))
	err := s.Service.UpdatePendingInvoiceAdjustment(clientId, invoiceId)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
}
