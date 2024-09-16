package api

import (
	"encoding/json"
	"github.com/opg-sirius-finance-hub/shared"
	"net/http"
	"strconv"
)

func (s *Server) updatePendingInvoiceAdjustment(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var body shared.UpdateInvoiceAdjustment

	clientId, _ := strconv.Atoi(r.PathValue("clientId"))
	adjustmentId, _ := strconv.Atoi(r.PathValue("adjustmentId"))

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return
	}

	validationError := s.Validator.ValidateStruct(body)

	if len(validationError.Errors) != 0 {
		errorData, _ := json.Marshal(validationError)
		w.Header().Set("Content-Type", "application/json")
		http.Error(w, "", http.StatusBadRequest)
		_, _ = w.Write(errorData)

		return
	}

	err := s.Service.UpdatePendingInvoiceAdjustment(ctx, clientId, adjustmentId, body.Status)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
