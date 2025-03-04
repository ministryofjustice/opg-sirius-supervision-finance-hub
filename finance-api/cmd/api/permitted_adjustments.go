package api

import (
	"encoding/json"
	"errors"
	"github.com/jackc/pgx/v5"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/apierror"
	"net/http"
)

func (s *Server) getPermittedAdjustments(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()

	invoiceId, err := s.getPathID(r, "invoiceId")
	if err != nil {
		return err
	}

	types, err := s.service.GetPermittedAdjustments(ctx, invoiceId)

	if errors.Is(err, pgx.ErrNoRows) {
		return apierror.NotFoundError(err)
	} else if err != nil {
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	return json.NewEncoder(w).Encode(types)
}
