package api

import (
	"net/http"
	"strconv"
)

func (s *Server) updatePendingInvoiceAdjustment(w http.ResponseWriter, r *http.Request) {

	ledgerId, _ := strconv.Atoi(r.PathValue("ledgerId"))
	err := s.Service.UpdatePendingInvoiceAdjustment(ledgerId)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
