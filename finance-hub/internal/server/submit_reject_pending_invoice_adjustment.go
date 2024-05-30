package server

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

type SubmitRejectPendingInvoiceAdjustmentHandler struct {
	router
}

func (h *SubmitRejectPendingInvoiceAdjustmentHandler) render(v AppVars, w http.ResponseWriter, r *http.Request) error {
	ctx := getContext(r)
	var (
		ledgerId, _ = strconv.Atoi(r.PathValue("ledgerId"))
		status      = "rejected"
	)

	err := h.Client().UpdatePendingInvoiceAdjustment(ctx, ctx.ClientId, ledgerId, status)
	if err != nil {
		return err
	}

	w.Header().Add("HX-Redirect", fmt.Sprintf("%s/clients/%d/pending-invoice-adjustments?success=reject-invoice-adjustment[%s]", v.EnvironmentVars.Prefix, ctx.ClientId, strings.ToUpper(r.PathValue("adjustmentType"))))
	return nil
}
