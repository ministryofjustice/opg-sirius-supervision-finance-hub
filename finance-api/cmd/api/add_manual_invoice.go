package api

import (
	"encoding/json"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
	"net/http"
)

func (s *Server) addManualInvoice(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()

	var addManualInvoice shared.AddManualInvoice
	defer unchecked(r.Body.Close)

	if err := json.NewDecoder(r.Body).Decode(&addManualInvoice); err != nil {
		return err
	}

	validationError := s.validator.ValidateStruct(addManualInvoice)

	if len(validationError.Errors) != 0 {
		return validationError
	}

	clientId, err := s.getPathID(r, "clientId")
	if err != nil {
		return err
	}

	err = s.service.AddManualInvoice(ctx, clientId, addManualInvoice)

	if err != nil {
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	return nil
}
