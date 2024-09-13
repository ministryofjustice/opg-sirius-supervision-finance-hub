package api

import (
	"encoding/json"
	"github.com/opg-sirius-finance-hub/shared"
	"net/http"
	"strconv"
)

func (s *Server) updatePendingInvoiceAdjustment(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()

	var body shared.UpdateInvoiceAdjustment

	clientId, _ := strconv.Atoi(r.PathValue("clientId"))
	adjustmentId, _ := strconv.Atoi(r.PathValue("adjustmentId"))

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return err
	}

	validationError := s.Validator.ValidateStruct(body)

	if len(validationError.Errors) != 0 {
		w.Header().Set("Content-Type", "application/json")
		http.Error(w, "", http.StatusUnprocessableEntity)
		err := json.NewEncoder(w).Encode(validationError)
		return err
	}

	err := s.Service.UpdatePendingInvoiceAdjustment(ctx, clientId, adjustmentId, body.Status)

	if err != nil {
		return err
	}

	w.WriteHeader(http.StatusNoContent)
	return nil
}
