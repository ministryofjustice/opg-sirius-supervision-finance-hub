package server

import (
	"errors"
	"fmt"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/apierror"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-hub/internal/api"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-hub/internal/util"
	"net/http"
)

type SubmitDirectDebitHandler struct {
	router
}

func (h *SubmitDirectDebitHandler) render(v AppVars, w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	clientID := getClientID(r)

	var (
		accountHolder = r.PostFormValue("accountHolder")
		accountName   = r.PostFormValue("accountName")
		sortCode      = r.PostFormValue("sortCode")
		accountNumber = r.PostFormValue("accountNumber")
	)

	err := h.Client().AddDirectDebit(ctx, clientID, accountHolder, accountName, sortCode, accountNumber)

	if err == nil {
		w.Header().Add("HX-Redirect", fmt.Sprintf("%s/clients/%d/invoices?success=direct-debit", v.EnvironmentVars.Prefix, clientID))
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
