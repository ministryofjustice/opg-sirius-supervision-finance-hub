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

type SubmitCancelFeeReductionsHandler struct {
	router
}

func (h *SubmitCancelFeeReductionsHandler) render(v AppVars, w http.ResponseWriter, r *http.Request) error {
	ctx := getContext(r)

	var (
		notes             = r.PostFormValue("cancellationReason")
		feeReductionId, _ = strconv.Atoi(r.PathValue("feeReductionId"))
	)
	err := h.Client().CancelFeeReduction(ctx, ctx.ClientId, feeReductionId, notes)

	if err == nil {
		w.Header().Add("HX-Redirect", fmt.Sprintf("%s/clients/%d/fee-reductions?success=fee-reduction[CANCELLED]", v.EnvironmentVars.Prefix, ctx.ClientId))
	} else {
		var (
			valErr shared.ValidationError
			stErr  api.StatusError
		)
		if errors.As(err, &valErr) {
			data := AppVars{Errors: util.RenameErrors(valErr.Errors)}
			w.WriteHeader(http.StatusUnprocessableEntity)
			err = h.execute(w, r, data)
		} else if errors.As(err, &stErr) {
			data := AppVars{Error: stErr.Error()}
			w.WriteHeader(stErr.Code)
			err = h.execute(w, r, data)
		}
	}

	return err
}
