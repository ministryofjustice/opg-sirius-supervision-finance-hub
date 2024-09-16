package api

import (
	"encoding/json"
	"errors"
	"github.com/opg-sirius-finance-hub/apierror"
	"github.com/opg-sirius-finance-hub/shared"
	"net/http"
	"strconv"
)

func (s *Server) PostLedgerEntry(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()

	clientId, _ := strconv.Atoi(r.PathValue("clientId"))
	invoiceId, _ := strconv.Atoi(r.PathValue("invoiceId"))

	var ledgerEntry shared.AddInvoiceAdjustmentRequest
	err := json.NewDecoder(r.Body).Decode(&ledgerEntry)
	if err != nil {
		return err
	}

	validationError := s.Validator.ValidateStruct(ledgerEntry)

	if len(validationError.Errors) != 0 {
		w.Header().Set("Content-Type", "application/json")
		http.Error(w, "", http.StatusUnprocessableEntity)
		err = json.NewEncoder(w).Encode(validationError)
		return err
	}

	invoiceReference, err := s.Service.AddInvoiceAdjustment(ctx, clientId, invoiceId, &ledgerEntry)

	if err != nil {
		var e apierror.BadRequests
		ok := errors.As(err, &e)
		if ok {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			err = json.NewEncoder(w).Encode(e)
			if err != nil {
				return err
			}
			return err
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(invoiceReference)
	return err
}
