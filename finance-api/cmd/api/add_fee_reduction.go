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

	err := s.Service.AddFeeReduction(addFeeReduction)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err != nil {
		return
	}
}
