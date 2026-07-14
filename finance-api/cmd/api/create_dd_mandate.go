package api

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/apierror"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/allpay"
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

	pendingCollection, err := s.service.CreateDirectDebitMandate(ctx, clientId, createMandate)
	if err != nil {
		var modulusErr allpay.ErrorModulusCheckFailed
		if errors.As(err, &modulusErr) {
			return modulusCheckFailedValidationError(modulusErr)
		}
		logger.Error("creating mandate in createDirectDebitMandate failed", "err", err)
		return err
	}

	if err := s.service.SendDirectDebitCollectionEvent(ctx, clientId, pendingCollection); err != nil {
		logger.Error("Sending direct-debit-collection event in createDirectDebitMandate failed", "err", err)
		return err
	}

	w.WriteHeader(http.StatusCreated)
	return nil
}

func modulusCheckFailedValidationError(err allpay.ErrorModulusCheckFailed) apierror.ValidationError {
	return apierror.ValidationError{
		Errors: apierror.ValidationErrors{
			"AccountDetails": map[string]string{
				"invalid": err.Error(),
			},
		},
	}
}
