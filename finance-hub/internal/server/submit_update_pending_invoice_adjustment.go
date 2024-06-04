package server

import (
	"fmt"
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
	if err != nil {
		return err
	}

	var successAction string
	if status == "APPROVED" {
		successAction = "approve"
	} else if status == "REJECTED" {
		successAction = "reject"
	}

	w.Header().Add("HX-Redirect", fmt.Sprintf("%s/clients/%d/pending-invoice-adjustments?success=%s-invoice-adjustment[%s]", v.EnvironmentVars.Prefix, ctx.ClientId, successAction, strings.ToUpper(r.PathValue("adjustmentType"))))
	return nil
}
