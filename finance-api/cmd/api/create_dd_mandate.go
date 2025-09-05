package api

import (
	"encoding/json"
	"net/http"

	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
)

func (s *Server) createDirectDebitMandate(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()

	var createMandate shared.CreateMandate
	defer unchecked(r.Body.Close)

	if err := json.NewDecoder(r.Body).Decode(&createMandate); err != nil {
		return err
	}

	validationError := s.validator.ValidateStruct(createMandate)

	if len(validationError.Errors) != 0 {
		return validationError
	}
	//TODO:
	// modulus check
	// return 400
	// create mandate
	// create schedule

	clientId, err := s.getPathID(r, "clientId")
	if err != nil {
		return err
	}

	if err := s.service.CreateDirectDebitMandate(ctx, clientId, createMandate); err != nil {
		return err
	}

	w.WriteHeader(http.StatusCreated)
	return nil
}
