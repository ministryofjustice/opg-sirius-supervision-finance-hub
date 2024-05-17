package server

import (
	"fmt"
	"net/http"
	"strconv"
)

type SubmitPendingInvoiceAdjustmentHandler struct {
	router
}

func (h *SubmitPendingInvoiceAdjustmentHandler) render(v AppVars, w http.ResponseWriter, r *http.Request) error {
	ctx := getContext(r)
	var (
		invoiceId, _ = strconv.Atoi(r.PathValue("id"))
	)

	err := h.Client().SubmitPendingInvoiceAdjustment(ctx, ctx.ClientId, invoiceId)
	//var verr shared.ValidationError
	//if errors.As(err, &verr) {
	//	data := UpdateInvoices{shared.AdjustmentTypes, r.PathValue("id"), r.PathValue("invoiceId"), v}
	//	data.Errors = util.RenameErrors(verr.Errors)
	//	w.WriteHeader(http.StatusUnprocessableEntity)
	//	return h.execute(w, r, data)
	//}
	if err != nil {
		return err
	}

	w.Header().Add("HX-Redirect", fmt.Sprintf("%s/clients/%d/invoices?success=%s", v.EnvironmentVars.Prefix, ctx.ClientId, adjustmentType))
	return nil
}
