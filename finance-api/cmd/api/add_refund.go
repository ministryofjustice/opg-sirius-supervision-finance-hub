package api

import (
	"encoding/json"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
	"net/http"
)

func (s *Server) addRefund(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()

	var refund shared.AddRefund
	defer r.Body.Close()

	if err := json.NewDecoder(r.Body).Decode(&refund); err != nil {
		return err
	}

	validationError := s.validator.ValidateStruct(refund)

	if len(validationError.Errors) != 0 {
		return validationError
	}

	clientId, err := s.getPathID(r, "clientId")
	if err != nil {
		return err
	}

	err = s.service.AddRefund(ctx, clientId, refund)

	if err != nil {
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	return nil
}
