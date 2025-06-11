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
		w.Header().Add("HX-Redirect", fmt.Sprintf("%s/clients/%d/refunds?success=refund-added", v.EnvironmentVars.Prefix, clientID))
		return nil
	}

	var (
		ve    apierror.ValidationError
		stErr api.StatusError
		data  AppVars
	)
	switch {
	case errors.As(err, &ve):
		{
			data = AppVars{Errors: util.RenameErrors(ve.Errors)}
			w.WriteHeader(http.StatusUnprocessableEntity)
		}
	case errors.As(err, &stErr):
		{
			data = AppVars{Error: stErr.Error(), Code: stErr.Code}
			w.WriteHeader(http.StatusUnprocessableEntity)
		}
	default:
		data = AppVars{Error: err.Error()}
		w.WriteHeader(http.StatusInternalServerError)
	}

	return h.execute(w, r, data)
}
