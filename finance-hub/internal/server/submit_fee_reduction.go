package server

import (
	"errors"
	"fmt"
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
		feeType       = r.PostFormValue("feeType")
		startYear     = r.PostFormValue("startDateYear")
		lengthOfAward = r.PostFormValue("lengthOfAward")
		dateReceived  = r.PostFormValue("dateReceived")
		notes         = r.PostFormValue("notes")
	)

	err := h.Client().AddFeeReduction(ctx, ctx.ClientId, feeType, startYear, lengthOfAward, dateReceived, notes)
	var verr shared.ValidationError
	if errors.As(err, &verr) {
		data := AppVars{Errors: util.RenameErrors(verr.Errors)}
		data.Errors = util.RenameErrors(verr.Errors)
		w.WriteHeader(http.StatusUnprocessableEntity)
		return h.execute(w, r, data)
	}

	if err != nil {
		return err
	}

	w.Header().Add("HX-Redirect", fmt.Sprintf("%s/clients/%d/fee-reductions?success=fee-reduction[%s]", v.EnvironmentVars.Prefix, ctx.ClientId, feeType))
	return nil
}
