package api

import (
	"encoding/json"
	"errors"
	"github.com/opg-sirius-finance-hub/shared"
	"net/http"
	"strconv"
)

func (s *Server) addManualInvoice(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()

	var addManualInvoice shared.AddManualInvoice
	defer r.Body.Close()

	if err := json.NewDecoder(r.Body).Decode(&addManualInvoice); err != nil {
		return err
	}

	validationError := s.Validator.ValidateStruct(addManualInvoice)

	if len(validationError.Errors) != 0 {
		w.Header().Set("Content-Type", "application/json")
		http.Error(w, "", http.StatusUnprocessableEntity)
		err := json.NewEncoder(w).Encode(validationError)
		return err
	}

	clientId, _ := strconv.Atoi(r.PathValue("clientId"))
	err := s.Service.AddManualInvoice(ctx, clientId, addManualInvoice)

	if err != nil {
		var e shared.BadRequests
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
	return nil
}
