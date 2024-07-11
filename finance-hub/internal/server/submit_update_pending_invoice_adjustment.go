package server

import (
	"errors"
	"fmt"
	"github.com/opg-sirius-finance-hub/finance-hub/internal/api"
	"net/http"
	"strconv"
	"strings"
)

type SubmitUpdatePendingInvoiceAdjustmentHandler struct {
	router
}

func (h *SubmitUpdatePendingInvoiceAdjustmentHandler) render(v AppVars, w http.ResponseWriter, r *http.Request) error {
	ctx := getContext(r)
	var (
		ledgerId, _ = strconv.Atoi(r.PathValue("ledgerId"))
		status      = strings.ToUpper(r.PathValue("status"))
	)

	err := h.Client().UpdatePendingInvoiceAdjustment(ctx, ctx.ClientId, ledgerId, status)

	if err == nil {
		w.Header().Add("HX-Redirect", fmt.Sprintf("%s/clients/%d/pending-invoice-adjustments?success=%s-invoice-adjustment[%s]", v.EnvironmentVars.Prefix, ctx.ClientId, strings.ToLower(status), strings.ToUpper(r.PathValue("adjustmentType"))))
	} else {
		var (
			stErr api.StatusError
		)
		if errors.As(err, &stErr) {
			data := AppVars{Error: stErr.Error()}
			w.WriteHeader(stErr.Code)
			err = h.execute(w, r, data)
		}
	}

	return err
}
