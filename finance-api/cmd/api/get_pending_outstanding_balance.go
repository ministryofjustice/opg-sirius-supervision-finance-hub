package api

import (
	"encoding/json"
	"errors"
	"github.com/jackc/pgx/v5"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/apierror"
	"net/http"
)

// getPendingOutstandingBalance returns the outstanding debt including any future confirmed ledgers for direct debit instructions
func (s *Server) getPendingOutstandingBalance(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()

	clientId, err := s.getPathID(r, "clientId")
	if err != nil {
		return err
	}

	outstandingBalance, err := s.service.GetPendingOutstandingBalance(ctx, clientId)

	if errors.Is(err, pgx.ErrNoRows) {
		return apierror.NotFoundError(err)
	} else if err != nil {
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	return json.NewEncoder(w).Encode(outstandingBalance)
}
