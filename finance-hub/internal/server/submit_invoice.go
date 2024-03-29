package server

import (
	"errors"
	"fmt"
	"github.com/opg-sirius-finance-hub/finance-hub/internal/api"
	"github.com/opg-sirius-finance-hub/finance-hub/internal/util"
	"github.com/opg-sirius-finance-hub/shared"
	"net/http"
	"strconv"
)

type SubmitInvoiceHandler struct {
	router
}

func (h *SubmitInvoiceHandler) render(v AppVars, w http.ResponseWriter, r *http.Request) error {
	ctx := getContext(r)
	var (
		invoiceId, _ = strconv.Atoi(r.PathValue("id"))
		invoiceType  = r.PostFormValue("invoiceType")
		notes        = r.PostFormValue("notes")
		amount       = r.PostFormValue("amount")
	)

	err := h.Client().UpdateInvoice(ctx, ctx.ClientId, invoiceId, invoiceType, notes, amount)
	var verr api.ValidationError
	if errors.As(err, &verr) {
		data := UpdateInvoices{shared.InvoiceTypes, r.PathValue("id"), r.PathValue("invoiceId"), v}
		data.Errors = util.RenameErrors(verr.Errors)
		w.WriteHeader(http.StatusUnprocessableEntity)
		return h.execute(w, r, data)
	}
	if err != nil {
		return err
	}

	w.Header().Add("HX-Redirect", fmt.Sprintf("%s/clients/%d/invoices?success=%s", v.EnvironmentVars.Prefix, ctx.ClientId, invoiceType))
	return nil
}
