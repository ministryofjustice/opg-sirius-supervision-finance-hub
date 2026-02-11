package api

import (
	"encoding/json"
	"net/http"
)

func (s *Server) getAnnualBillingLettersCount(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()

	accountInfo, err := s.service.GetAnnualBillingInfo(ctx)

	if err != nil {
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	return json.NewEncoder(w).Encode(accountInfo)
}
