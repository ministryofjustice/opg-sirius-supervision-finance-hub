package api

import (
	"encoding/json"
	"errors"
	"github.com/opg-sirius-finance-hub/finance-api/internal/service"
	"github.com/opg-sirius-finance-hub/shared"
	"net/http"
	"strconv"
)

func (s *Server) addFeeReduction(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var addFeeReduction shared.AddFeeReduction
	defer r.Body.Close()

	if err := json.NewDecoder(r.Body).Decode(&addFeeReduction); err != nil {
		return
	}

	validationError := s.Validator.ValidateStruct(addFeeReduction)

	if len(validationError.Errors) != 0 {
		errorData, _ := json.Marshal(validationError)
		w.Header().Set("Content-Type", "application/json")
		http.Error(w, "", http.StatusUnprocessableEntity)
		_, _ = w.Write(errorData)

		return
	}

	clientId, _ := strconv.Atoi(r.PathValue("clientId"))
	err := s.Service.AddFeeReduction(ctx, clientId, addFeeReduction)

	if err != nil {
		var e service.BadRequest
		ok := errors.As(err, &e)
		if ok {
			http.Error(w, e.Reason, http.StatusBadRequest)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
}
