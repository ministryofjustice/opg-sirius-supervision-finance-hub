package server

import (
	"errors"
	"fmt"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-hub/internal/api"
	"net/http"
	"strconv"
)

type SubmitRefundDecisionHandler struct {
	router
}

func (h *SubmitRefundDecisionHandler) render(v AppVars, w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	clientID := getClientID(r)

	// Limit request body size to 10MB to prevent memory exhaustion
	r.Body = http.MaxBytesReader(w, r.Body, 10<<20)

	var (
		refundID, _ = strconv.Atoi(r.PathValue("refundId"))
		decision    = r.PostFormValue("decision")
	)

	err := h.Client().UpdateRefundDecision(ctx, clientID, refundID, decision)

	if err == nil {
		w.Header().Add("HX-Redirect", fmt.Sprintf("%s/clients/%d/refunds?success=refunds[%s]", v.EnvironmentVars.Prefix, clientID, decision))
	} else {
		var (
			stErr api.StatusError
		)
		if errors.As(err, &stErr) {
			v.Error = stErr.Error()
			v.Code = stErr.Code
			data := v
			w.WriteHeader(stErr.Code)
			err = h.execute(w, r, data)
		}
	}

	return err
}
