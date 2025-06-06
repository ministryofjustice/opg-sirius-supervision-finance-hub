package server

import (
	"errors"
	"fmt"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/apierror"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-hub/internal/api"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-hub/internal/util"
	"net/http"
)

type SubmitRefundHandler struct {
	router
}

func (h *SubmitRefundHandler) render(v AppVars, w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	clientID := getClientID(r)

	var (
		accountName   = r.PostFormValue("accountName")
		accountNumber = r.PostFormValue("accountNumber")
		sortCode      = r.PostFormValue("sortCode")
		notes         = r.PostFormValue("notes")
	)

	err := h.Client().AddRefund(ctx, clientID, accountName, accountNumber, sortCode, notes)

	if err == nil {
		w.Header().Add("HX-Redirect", fmt.Sprintf("%s/clients/%d/refund?success=refund", v.EnvironmentVars.Prefix, clientID))
	} else {
		var (
			valErr apierror.ValidationError
			stErr  api.StatusError
		)
		if errors.As(err, &valErr) {
			data := AppVars{Errors: util.RenameErrors(valErr.Errors)}
			w.WriteHeader(http.StatusUnprocessableEntity)
			err = h.execute(w, r, data)
		} else if errors.As(err, &stErr) {
			data := AppVars{Error: stErr.Error(), Code: stErr.Code}
			w.WriteHeader(stErr.Code)
			err = h.execute(w, r, data)
		}
	}

	return err
}
