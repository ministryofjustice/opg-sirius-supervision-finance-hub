package api

import (
	"encoding/json"
	"github.com/opg-sirius-finance-hub/shared"
	"net/http"
	"strconv"
)

func (s *Server) AddInvoiceAdjustment(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()

	clientId, _ := strconv.Atoi(r.PathValue("clientId"))
	invoiceId, _ := strconv.Atoi(r.PathValue("invoiceId"))

	var ledgerEntry shared.AddInvoiceAdjustmentRequest
	if err := json.NewDecoder(r.Body).Decode(&ledgerEntry); err != nil {
		return err
	}

	validationError := s.Validator.ValidateStruct(ledgerEntry)

	if len(validationError.Errors) != 0 {
		return validationError
	}

	invoiceReference, err := s.Service.AddInvoiceAdjustment(ctx, clientId, invoiceId, &ledgerEntry)

	if err != nil {
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	return json.NewEncoder(w).Encode(invoiceReference)
}
