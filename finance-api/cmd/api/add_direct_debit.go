package api

import (
	"encoding/json"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
	"net/http"
)

func (s *Server) addDirectDebit(w http.ResponseWriter, r *http.Request) error {
	var directDebit shared.AddDirectDebit
	defer unchecked(r.Body.Close)

	if err := json.NewDecoder(r.Body).Decode(&directDebit); err != nil {
		return err
	}

	validationError := s.validator.ValidateStruct(directDebit)

	if len(validationError.Errors) != 0 {
		return validationError
	}

	//call all pay here

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	return nil
}
