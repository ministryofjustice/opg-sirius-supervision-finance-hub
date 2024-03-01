package server

import (
	"errors"
	"fmt"
	"github.com/opg-sirius-finance-hub/internal/sirius"
	"github.com/opg-sirius-finance-hub/internal/util"
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
	)

	err := h.Client().UpdateInvoice(ctx, ctx.ClientId, invoiceId, invoiceType, notes)
	var verr sirius.ValidationError
	if errors.As(err, &verr) {
		w.WriteHeader(http.StatusUnprocessableEntity)
		return h.execute(w, r, util.RenameErrors(verr.Errors))
	}
	if err != nil {
		return err
	}

	w.Header().Add("HX-Redirect", fmt.Sprintf("%s/clients/%d/invoices?success=%s", v.EnvironmentVars.Prefix, ctx.ClientId, invoiceType))
	return nil
}
