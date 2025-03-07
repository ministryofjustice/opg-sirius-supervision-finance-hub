package server

import (
	"errors"
	"fmt"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/apierror"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-hub/internal/api"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-hub/internal/util"
	"net/http"
	"net/url"
	"strings"
)

type SubmitManualInvoiceHandler struct {
	router
}

func getFieldPointer(postForm url.Values, key string) *string {
	var fieldPointer *string
	if postForm.Has(key) {
		fieldString := postForm.Get(key)
		fieldPointer = &fieldString
	}
	return fieldPointer
}

func (h *SubmitManualInvoiceHandler) render(v AppVars, w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	clientID := getClientID(r)

	invoiceType := r.PostFormValue("invoiceType")

	err := h.Client().AddManualInvoice(
		ctx,
		clientID,
		invoiceType,
		getFieldPointer(r.PostForm, "amount"),
		getFieldPointer(r.PostForm, "raisedDate"),
		getFieldPointer(r.PostForm, "raisedYear"),
		getFieldPointer(r.PostForm, "startDate"),
		getFieldPointer(r.PostForm, "endDate"),
		getFieldPointer(r.PostForm, "supervisionLevel"),
	)

	if err == nil {
		w.Header().Add("HX-Redirect", fmt.Sprintf("%s/clients/%d/invoices?success=invoice-type[%s]", v.EnvironmentVars.Prefix, clientID, strings.ToUpper(invoiceType)))
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
