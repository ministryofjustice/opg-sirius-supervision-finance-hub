package api

import (
	"encoding/json"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
	"net/http"
	"strconv"
)

func (s *Server) addFeeReduction(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()

	var addFeeReduction shared.AddFeeReduction
	defer r.Body.Close()

	if err := json.NewDecoder(r.Body).Decode(&addFeeReduction); err != nil {
		return err
	}

	validationError := s.Validator.ValidateStruct(addFeeReduction)

	if len(validationError.Errors) != 0 {
		return validationError
	}

	clientId, _ := strconv.Atoi(r.PathValue("clientId"))
	err := s.Service.AddFeeReduction(ctx, clientId, addFeeReduction)

	if err != nil {
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	return nil
}
