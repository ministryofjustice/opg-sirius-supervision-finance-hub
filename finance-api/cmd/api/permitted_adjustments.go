package api

import (
	"encoding/json"
	"errors"
	"github.com/jackc/pgx/v5"
	"github.com/opg-sirius-finance-hub/apierror"
	"net/http"
	"strconv"
)

func (s *Server) getPermittedAdjustments(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()

	invoiceId, _ := strconv.Atoi(r.PathValue("invoiceId"))
	types, err := s.Service.GetPermittedAdjustments(ctx, invoiceId)

	if errors.Is(err, pgx.ErrNoRows) {
		return apierror.NotFoundError(err)
	} else if err != nil {
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(types)
	return err
}
