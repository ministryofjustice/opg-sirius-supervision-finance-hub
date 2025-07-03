package api

import (
	"encoding/json"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
	"net/http"
)

func (s *Server) addFeeReduction(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()

	var addFeeReduction shared.AddFeeReduction
	defer unchecked(r.Body.Close)

	if err := json.NewDecoder(r.Body).Decode(&addFeeReduction); err != nil {
		return err
	}

	validationError := s.validator.ValidateStruct(addFeeReduction)

	if len(validationError.Errors) != 0 {
		return validationError
	}

	clientId, err := s.getPathID(r, "clientId")
	if err != nil {
		return err
	}

	err = s.service.AddFeeReduction(ctx, clientId, addFeeReduction)

	if err != nil {
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	return nil
}
