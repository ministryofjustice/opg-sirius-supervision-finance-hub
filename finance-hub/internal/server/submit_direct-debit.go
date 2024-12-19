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
	ctx := getContext(r)

	var (
		accountHolder = r.PostFormValue("accountHolder")
		accountName   = r.PostFormValue("accountName")
		sortCode      = r.PostFormValue("sortCode")
		accountNumber = r.PostFormValue("accountNumber")
	)

	err := h.Client().SubmitDirectDebit(accountHolder, accountName, sortCode, accountNumber)

	if err == nil {
		w.Header().Add("HX-Redirect", fmt.Sprintf("%s/clients/%d/invoices?success=direct-debit", v.EnvironmentVars.Prefix, ctx.ClientId))
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
			data := AppVars{Error: stErr.Error()}
			w.WriteHeader(stErr.Code)
			err = h.execute(w, r, data)
		}
	}

	return err
}
