package server

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

type SubmitApprovePendingInvoiceAdjustmentHandler struct {
	router
}

func (h *SubmitApprovePendingInvoiceAdjustmentHandler) render(v AppVars, w http.ResponseWriter, r *http.Request) error {
	ctx := getContext(r)
	var (
		ledgerId, _ = strconv.Atoi(r.PathValue("ledgerId"))
		status      = "approved"
	)

	err := h.Client().UpdatePendingInvoiceAdjustment(ctx, ctx.ClientId, ledgerId, status)
	if err != nil {
		return err
	}

	w.Header().Add("HX-Redirect", fmt.Sprintf("%s/clients/%d/pending-invoice-adjustments?success=approve-invoice-adjustment[%s]", v.EnvironmentVars.Prefix, ctx.ClientId, strings.ToUpper(r.PathValue("adjustmentType"))))
	return nil
}
