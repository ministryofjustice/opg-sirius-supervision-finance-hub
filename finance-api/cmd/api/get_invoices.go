package api

import (
	"encoding/json"
	"net/http"
)

func (s *Server) getInvoices(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()

	clientId, err := s.getPathID(r, "clientId")
	if err != nil {
		return err
	}

	invoices, err := s.service.GetInvoices(ctx, clientId)

	if err != nil {
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	return json.NewEncoder(w).Encode(invoices)
}
