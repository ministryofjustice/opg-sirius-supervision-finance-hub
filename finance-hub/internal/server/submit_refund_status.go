package server

import (
	"errors"
	"fmt"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-hub/internal/api"
	"net/http"
	"strconv"
)

type SubmitRefundStatusHandler struct {
	router
}

func (h *SubmitRefundStatusHandler) render(v AppVars, w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	clientID := getClientID(r)
	var (
		refundID, _ = strconv.Atoi(r.PathValue("refundId"))
		status      = r.PostFormValue("status")
	)

	err := h.Client().UpdateRefundStatus(ctx, clientID, refundID, status)

	if err == nil {
		w.Header().Add("HX-Redirect", fmt.Sprintf("%s/clients/%d/refunds?success=refunds[%s]", v.EnvironmentVars.Prefix, clientID, status))
	} else {
		var (
			stErr api.StatusError
		)
		if errors.As(err, &stErr) {
			data := AppVars{Error: stErr.Error(), Code: stErr.Code}
			w.WriteHeader(stErr.Code)
			err = h.execute(w, r, data)
		}
	}

	return err
}
