package api

import (
	"encoding/json"
	"github.com/opg-sirius-finance-hub/shared"
	"net/http"
)

func (s *Server) addFeeReduction(w http.ResponseWriter, r *http.Request) {
	var addFeeReduction shared.AddFeeReduction
	defer r.Body.Close()

	if err := json.NewDecoder(r.Body).Decode(&addFeeReduction); err != nil {
		return
	}

	validationError, err := s.Service.AddFeeReduction(addFeeReduction)

	if len(validationError.Errors) != 0 {
		errorData, _ := json.Marshal(validationError)
		w.Header().Set("Content-Type", "application/json")
		// Write the JSON response body
		_, _ = w.Write(errorData)
		http.Error(w, "", http.StatusInternalServerError)
		return
	}

	if err != nil {
		//errorData, _ := json.Marshal(ValidationError{Message: "string here", Errors: ValidationErrors{"dateReportReceived": {"invalid-lte": "This must be on or before 25/04/2024"}, "dateReportLodged": {"invalid-lte": "This must be on or before 25/04/2024"}}})
		//w.Header().Set("Content-Type", "application/json")
		//// Write the JSON response body
		//_, _ = w.Write(errorData)
		//http.Error(w, "", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err != nil {
		return
	}
}
