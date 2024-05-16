package api

import (
	"encoding/json"
	"github.com/opg-sirius-finance-hub/shared"
	"net/http"
	"strconv"
)

func (s *Server) cancelFeeReduction(w http.ResponseWriter, r *http.Request) {
	var cancelFeeReduction shared.CancelFeeReduction
	defer r.Body.Close()

	if err := json.NewDecoder(r.Body).Decode(&cancelFeeReduction); err != nil {
		return
	}

	validationError := s.Validator.ValidateStruct(cancelFeeReduction, "CancelFeeReductionNotes")

	if len(validationError.Errors) != 0 {
		errorData, _ := json.Marshal(validationError)
		w.Header().Set("Content-Type", "application/json")
		http.Error(w, "", http.StatusUnprocessableEntity)
		_, _ = w.Write(errorData)

		return
	}

	feeReductionId, _ := strconv.Atoi(r.PathValue("feeReductionId"))
	err := s.Service.CancelFeeReduction(feeReductionId)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
}
