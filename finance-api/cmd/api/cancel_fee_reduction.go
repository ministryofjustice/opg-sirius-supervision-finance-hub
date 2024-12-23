package api

import (
	"encoding/json"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
	"net/http"
	"strconv"
)

func (s *Server) cancelFeeReduction(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()

	var cancelFeeReduction shared.CancelFeeReduction
	defer r.Body.Close()

	if err := json.NewDecoder(r.Body).Decode(&cancelFeeReduction); err != nil {
		return err
	}

	validationError := s.validator.ValidateStruct(cancelFeeReduction)

	if len(validationError.Errors) != 0 {
		return validationError
	}

	feeReductionId, _ := strconv.Atoi(r.PathValue("feeReductionId"))
	err := s.service.CancelFeeReduction(ctx, feeReductionId, cancelFeeReduction)

	if err != nil {
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	return nil
}
