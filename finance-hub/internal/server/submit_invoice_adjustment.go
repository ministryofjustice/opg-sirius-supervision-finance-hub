package server

import (
	"errors"
	"fmt"
	"github.com/opg-sirius-finance-hub/finance-hub/internal/util"
	"github.com/opg-sirius-finance-hub/shared"
	"net/http"
	"strconv"
)

type SubmitInvoiceAdjustmentHandler struct {
	router
}

type InvoiceAdjustmentForm struct {
	AdjustmentType shared.AdjustmentType
	Notes          string
	Amount         string
}

func (h *SubmitInvoiceAdjustmentHandler) render(v AppVars, w http.ResponseWriter, r *http.Request) error {
	ctx := getContext(r)
	var (
		invoiceId, _   = strconv.Atoi(r.PathValue("invoiceId"))
		adjustmentType = r.PostFormValue("adjustmentType")
		notes          = r.PostFormValue("notes")
		amount         = r.PostFormValue("amount")
	)

	err := h.Client().AdjustInvoice(ctx, ctx.ClientId, v.EnvironmentVars.SupervisionBillingTeam, invoiceId, adjustmentType, notes, amount)

	var verr shared.ValidationError
	if errors.As(err, &verr) {
		data := AppVars{Errors: util.RenameErrors(verr.Errors)}
		w.WriteHeader(http.StatusUnprocessableEntity)
		return h.execute(w, r, data)
	}

	var be shared.BadRequest
	if errors.As(err, &be) {
		data := AppVars{Errors: shared.ValidationErrors{be.Field: map[string]string{"tooHigh": be.Reason}}}
		w.WriteHeader(http.StatusUnprocessableEntity)
		return h.execute(w, r, data)
	}

	if err != nil {
		return err
	}

	w.Header().Add("HX-Redirect", fmt.Sprintf("%s/clients/%d/invoices?success=%s", v.EnvironmentVars.Prefix, ctx.ClientId, adjustmentType))
	return nil
}
