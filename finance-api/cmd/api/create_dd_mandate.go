package api

import (
	"encoding/json"
	"net/http"

	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
)

func (s *Server) createDirectDebitMandate(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	logger := s.Logger(ctx)

	var createMandate shared.CreateMandate
	defer unchecked(r.Body.Close)

	if err := json.NewDecoder(r.Body).Decode(&createMandate); err != nil {
		return err
	}

	validationError := s.validator.ValidateStruct(createMandate)

	if len(validationError.Errors) != 0 {
		return validationError
	}

	clientId, err := s.getPathID(r, "clientId")
	if err != nil {
		return err
	}

	if err := s.service.CreateDirectDebitMandate(ctx, clientId, createMandate); err != nil {
		logger.Error("creating mandate in createDirectDebitMandate failed", "err", err)
		return err
	}

	if err := s.service.CreateDirectDebitSchedule(ctx, clientId, shared.CreateSchedule{AllPayCustomer: createMandate.AllPayCustomer}); err != nil {
		logger.Error("creating schedule in createDirectDebitMandate failed", "err", err)
		// Confirmed with business they do not want an error message returned even if the schedule fails
	}

	w.WriteHeader(http.StatusCreated)
	return nil
}
