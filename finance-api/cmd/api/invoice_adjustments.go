package api

import (
	"encoding/json"
	"net/http"
	"strconv"
)

func (s *Server) getInvoiceAdjustments(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()

	clientId, _ := strconv.Atoi(r.PathValue("clientId"))
	data, err := s.Service.GetInvoiceAdjustments(ctx, clientId)

	if err != nil {
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	return json.NewEncoder(w).Encode(data)
}
