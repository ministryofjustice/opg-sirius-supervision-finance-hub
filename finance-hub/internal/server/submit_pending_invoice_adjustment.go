package server

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

type SubmitPendingInvoiceAdjustmentHandler struct {
	router
}

func (h *SubmitPendingInvoiceAdjustmentHandler) render(v AppVars, w http.ResponseWriter, r *http.Request) error {
	ctx := getContext(r)
	var (
		ledgerId, _ = strconv.Atoi(r.PathValue("ledgerId"))
	)

	err := h.Client().UpdatePendingInvoiceAdjustment(ctx, ctx.ClientId, ledgerId)
	if err != nil {
		return err
	}

	w.Header().Add("HX-Redirect", fmt.Sprintf("%s/clients/%d/pending-invoice-adjustments?success=%s", v.EnvironmentVars.Prefix, ctx.ClientId, strings.ToLower(r.PathValue("adjustmentType"))))
	return nil
}
