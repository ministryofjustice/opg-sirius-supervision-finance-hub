package api

import (
	"encoding/json"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
	"net/http"
)

func (s *Server) updatePaymentMethod(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()

	var body shared.UpdatePaymentMethod

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return err
	}

	validationError := s.validator.ValidateStruct(body)

	if len(validationError.Errors) != 0 {
		return validationError
	}

	clientId, err := s.getPathID(r, "clientId")
	if err != nil {
		return err
	}

	err = s.service.UpdatePaymentMethod(ctx, clientId, body.PaymentMethod)

	if err != nil {
		return err
	}

	w.WriteHeader(http.StatusNoContent)
	return nil
}
