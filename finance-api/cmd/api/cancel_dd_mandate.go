package api

import (
	"encoding/json"
	"net/http"

	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
)

func (s *Server) cancelDirectDebitMandate(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()

	var cancelMandate shared.CancelMandate
	defer unchecked(r.Body.Close)

	if err := json.NewDecoder(r.Body).Decode(&cancelMandate); err != nil {
		return err
	}

	validationError := s.validator.ValidateStruct(cancelMandate)

	if len(validationError.Errors) != 0 {
		return validationError
	}

	clientId, err := s.getPathID(r, "clientId")
	if err != nil {
		return err
	}

	if err := s.service.CancelDirectDebitMandate(ctx, clientId, cancelMandate); err != nil {
		return err
	}

	w.WriteHeader(http.StatusNoContent)
	return nil
}
