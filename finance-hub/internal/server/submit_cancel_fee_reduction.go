package server

import (
	"errors"
	"fmt"
	"github.com/opg-sirius-finance-hub/finance-hub/internal/util"
	"github.com/opg-sirius-finance-hub/shared"
	"net/http"
	"strconv"
)

type CancelFeeReductionFormValues struct {
	Notes string
}

type SubmitCancelFeeReductionsHandler struct {
	router
}

func (h *SubmitCancelFeeReductionsHandler) render(v AppVars, w http.ResponseWriter, r *http.Request) error {
	ctx := getContext(r)

	var (
		notes             = r.PostFormValue("notes")
		feeReductionId, _ = strconv.Atoi(r.PathValue("feeReductionId"))
	)
	err := h.Client().CancelFeeReduction(ctx, ctx.ClientId, feeReductionId, notes)
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

	w.Header().Add("HX-Redirect", fmt.Sprintf("%s/clients/%d/fee-reductions?success=fee-reduction[CANCELLED]", v.EnvironmentVars.Prefix, ctx.ClientId))
	return nil
}
