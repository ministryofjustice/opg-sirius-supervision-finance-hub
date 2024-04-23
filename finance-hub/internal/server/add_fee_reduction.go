package server

import (
	"errors"
	"fmt"
	"github.com/opg-sirius-finance-hub/finance-hub/internal/api"
	"github.com/opg-sirius-finance-hub/finance-hub/internal/util"
	"github.com/opg-sirius-finance-hub/shared"
	"net/http"
)

type SubmitFeeReductionsHandler struct {
	router
}

func (h *SubmitFeeReductionsHandler) render(v AppVars, w http.ResponseWriter, r *http.Request) error {
	ctx := getContext(r)

	var (
		feeType           = r.PostFormValue("feeType")
		startYear         = r.PostFormValue("startDateYear")
		lengthOfAward     = r.PostFormValue("lengthOfAward")
		dateReceived      = r.PostFormValue("dateReceived")
		feeReductionNotes = r.PostFormValue("feeReductionNotes")
	)

	err := h.Client().AddFeeReduction(ctx, ctx.ClientId, feeType, startYear, lengthOfAward, dateReceived, feeReductionNotes)
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

	w.Header().Add("HX-Redirect", fmt.Sprintf("%s/clients/%d/fee-reductions?success=%s", v.EnvironmentVars.Prefix, ctx.ClientId, feeType))
	return nil
}
