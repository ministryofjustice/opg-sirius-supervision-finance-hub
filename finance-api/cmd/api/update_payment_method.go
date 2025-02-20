package api

import (
	"encoding/json"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
	"net/http"
	"strconv"
)

func (s *Server) updatePaymentMethod(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()

	var body shared.UpdatePaymentMethod

	clientId, _ := strconv.Atoi(r.PathValue("clientId"))

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return err
	}

	validationError := s.validator.ValidateStruct(body)

	if len(validationError.Errors) != 0 {
		return validationError
	}

	err := s.service.UpdatePaymentMethod(ctx, clientId, body.PaymentMethod)

	if err != nil {
		return err
	}

	w.WriteHeader(http.StatusNoContent)
	return nil
}
