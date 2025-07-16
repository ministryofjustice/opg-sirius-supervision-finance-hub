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
	fmt.Println("in handler")

	var (
		accountHolder = r.PostFormValue("accountHolder")
		accountName   = r.PostFormValue("accountName")
		sortCode      = r.PostFormValue("sortCode")
		accountNumber = r.PostFormValue("accountNumber")
	)

	fmt.Println(fmt.Sprintf("submit direct debit handler %s, %s, %s, %s", accountHolder, accountName, sortCode, accountNumber))
	err := h.Client().AddDirectDebit(ctx, clientID, accountHolder, accountName, sortCode, accountNumber)

	fmt.Println(fmt.Sprintf("submit direct debit handler ERR %s", err))

	if err == nil {
		w.Header().Add("HX-Redirect", fmt.Sprintf("%s/clients/%d/invoices?success=direct-debit", v.EnvironmentVars.Prefix, clientID))
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
