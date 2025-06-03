package server

import (
	"errors"
	"fmt"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/apierror"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-hub/internal/util"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
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
	ctx := r.Context()
	clientID := getClientID(r)

	var (
		invoiceId, _    = strconv.Atoi(r.PathValue("invoiceId"))
		adjustmentType  = r.PostFormValue("adjustmentType")
		notes           = r.PostFormValue("notes")
		amount          = r.PostFormValue("amount")
		managerOverride = r.PostFormValue("managerOverride")
	)

	err := h.Client().AddInvoiceAdjustment(ctx, clientID, v.EnvironmentVars.BillingTeamID, invoiceId, adjustmentType, notes, amount, managerOverride != "")

	if err == nil {
		w.Header().Add("HX-Redirect", fmt.Sprintf("%s/clients/%d/invoices?success=invoice-adjustment[%s]", v.EnvironmentVars.Prefix, clientID, adjustmentType))
		return nil
	}

	var (
		ve   apierror.ValidationError
		br   apierror.BadRequest
		data AppVars
	)
	switch {
	case errors.As(err, &ve):
		{
			data = AppVars{Errors: util.RenameErrors(ve.Errors)}
			w.WriteHeader(http.StatusUnprocessableEntity)
		}
	case errors.As(err, &br):
		{
			data = AppVars{Errors: apierror.ValidationErrors{br.Field: map[string]string{"tooHigh": br.Reason}}}
			w.WriteHeader(http.StatusUnprocessableEntity)
		}
	default:
		data = AppVars{Error: err.Error()}
		w.WriteHeader(http.StatusInternalServerError)
	}

	return h.execute(w, r, data)
}
