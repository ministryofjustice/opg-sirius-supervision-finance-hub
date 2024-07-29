package server

import (
	"errors"
	"fmt"
	"github.com/opg-sirius-finance-hub/finance-hub/internal/api"
	"github.com/opg-sirius-finance-hub/finance-hub/internal/util"
	"github.com/opg-sirius-finance-hub/shared"
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
	ctx := getContext(r)

	var (
		invoiceType      = r.PostFormValue("invoiceType")
		raisedDateYear   = r.PostFormValue("raisedDateYear")
		supervisionLevel = r.PostFormValue("supervisionLevel")
		amount           = getFieldPointer(r.PostForm, "amount")
		startDate        = getFieldPointer(r.PostForm, "startDate")
		raisedDate       = getFieldPointer(r.PostForm, "raisedDate")
		endDate          = getFieldPointer(r.PostForm, "endDate")
	)

	amount, startDate, raisedDate, endDate, supervisionLevel = invoiceData(invoiceType, amount, startDate, raisedDate, endDate, raisedDateYear, supervisionLevel)

	err := h.Client().AddManualInvoice(ctx, ctx.ClientId, invoiceType, amount, raisedDate, startDate, endDate, supervisionLevel)

	if err == nil {
		w.Header().Add("HX-Redirect", fmt.Sprintf("%s/clients/%d/invoices?success=invoice-type[%s]", v.EnvironmentVars.Prefix, ctx.ClientId, strings.ToUpper(invoiceType)))
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

func invoiceData(invoiceType string, amount *string, startDate *string, raisedDate *string, endDate *string, raisedDateYear string, supervisionLevel string) (*string, *string, *string, *string, string) {
	switch invoiceType {
	case shared.InvoiceTypeAD.Key():
		amountString := "100"
		amount = &amountString
		startDate = raisedDate
		endDate = raisedDate
	case shared.InvoiceTypeS2.Key(), shared.InvoiceTypeB2.Key():
		if raisedDateYear != "" {
			raisedDateString := raisedDateYear + "-03" + "-31"
			raisedDate = &raisedDateString
		}
		endDate = raisedDate
		supervisionLevel = "GENERAL"
	case shared.InvoiceTypeS3.Key(), shared.InvoiceTypeB3.Key():
		if raisedDateYear != "" {
			raisedDateString := raisedDateYear + "-03" + "-31"
			raisedDate = &raisedDateString
		}
		endDate = raisedDate
		supervisionLevel = "MINIMAL"
	}
	return amount, startDate, raisedDate, endDate, supervisionLevel
}
