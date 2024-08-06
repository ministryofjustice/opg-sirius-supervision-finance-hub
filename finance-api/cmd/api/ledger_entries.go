package api

import (
	"encoding/json"
	"errors"
	"github.com/opg-sirius-finance-hub/shared"
	"net/http"
	"strconv"
)

func (s *Server) PostLedgerEntry(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	clientId, _ := strconv.Atoi(r.PathValue("clientId"))
	invoiceId, _ := strconv.Atoi(r.PathValue("invoiceId"))

	var ledgerEntry shared.AddInvoiceAdjustmentRequest
	err := json.NewDecoder(r.Body).Decode(&ledgerEntry)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	validationError := s.Validator.ValidateStruct(ledgerEntry)

	if len(validationError.Errors) != 0 {
		errorData, _ := json.Marshal(validationError)
		w.Header().Set("Content-Type", "application/json")
		http.Error(w, "", http.StatusUnprocessableEntity)
		_, _ = w.Write(errorData)

		return
	}

	invoiceReference, err := s.Service.AddInvoiceAdjustment(ctx, clientId, invoiceId, &ledgerEntry)

	if err != nil {
		var e shared.BadRequest
		ok := errors.As(err, &e)
		if ok {
			errorData, _ := json.Marshal(e)
			w.Header().Set("Content-Type", "application/json")
			http.Error(w, "", http.StatusBadRequest)
			_, _ = w.Write(errorData)

			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	jsonData, err := json.Marshal(invoiceReference)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_, _ = w.Write(jsonData)
}
