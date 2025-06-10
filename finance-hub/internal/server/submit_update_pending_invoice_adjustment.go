package server

import (
	"errors"
	"fmt"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-hub/internal/api"
	"net/http"
	"strconv"
	"strings"
)

type SubmitUpdatePendingInvoiceAdjustmentHandler struct {
	router
}

func (h *SubmitUpdatePendingInvoiceAdjustmentHandler) render(v AppVars, w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	clientID := getClientID(r)
	var (
		adjustmentId, _ = strconv.Atoi(r.PathValue("adjustmentId"))
		status          = strings.ToUpper(r.PathValue("status"))
	)

	err := h.Client().UpdatePendingInvoiceAdjustment(ctx, clientID, adjustmentId, status)

	if err == nil {
		w.Header().Add("HX-Redirect", fmt.Sprintf("%s/clients/%d/invoice-adjustments?success=%s-invoice-adjustment[%s]", v.EnvironmentVars.Prefix, clientID, strings.ToLower(status), strings.ToUpper(r.PathValue("adjustmentType"))))
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
