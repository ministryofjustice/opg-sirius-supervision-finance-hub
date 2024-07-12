package server

import (
	"errors"
	"fmt"
	"github.com/opg-sirius-finance-hub/finance-hub/internal/api"
	"github.com/opg-sirius-finance-hub/finance-hub/internal/util"
	"github.com/opg-sirius-finance-hub/shared"
	"net/http"
	"strings"
)

type SubmitFeeReductionsHandler struct {
	router
}

func (h *SubmitFeeReductionsHandler) render(v AppVars, w http.ResponseWriter, r *http.Request) error {
	ctx := getContext(r)

	var (
		feeType       = r.PostFormValue("feeType")
		startYear     = r.PostFormValue("startYear")
		lengthOfAward = r.PostFormValue("lengthOfAward")
		dateReceived  = r.PostFormValue("dateReceived")
		notes         = r.PostFormValue("notes")
	)

	err := h.Client().AddFeeReduction(ctx, ctx.ClientId, feeType, startYear, lengthOfAward, dateReceived, notes)

	if err == nil {
		w.Header().Add("HX-Redirect", fmt.Sprintf("%s/clients/%d/fee-reductions?success=fee-reduction[%s]", v.EnvironmentVars.Prefix, ctx.ClientId, strings.ToUpper(feeType)))
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
