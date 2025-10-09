package api

import (
	"encoding/json"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
	"net/http"
)

func (s *Server) updatePendingInvoiceAdjustment(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()

	var body shared.UpdateInvoiceAdjustment

	clientId, err := s.getPathID(r, "clientId")
	if err != nil {
		return err
	}

	adjustmentId, err := s.getPathID(r, "adjustmentId")
	if err != nil {
		return err
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return err
	}

	validationError := s.validator.ValidateStruct(body)

	if len(validationError.Errors) != 0 {
		return validationError
	}

	err = s.service.UpdatePendingInvoiceAdjustment(ctx, clientId, adjustmentId, body.Status)

	if err != nil {
		return err
	}

	w.WriteHeader(http.StatusNoContent)
	return nil
}
