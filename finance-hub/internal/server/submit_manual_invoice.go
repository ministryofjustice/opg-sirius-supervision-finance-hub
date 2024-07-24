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

type SubmitManualInvoiceHandler struct {
	router
}

func (h *SubmitManualInvoiceHandler) render(v AppVars, w http.ResponseWriter, r *http.Request) error {
	ctx := getContext(r)

	var (
		invoiceType      = r.PostFormValue("invoiceType")
		amount           = r.PostFormValue("amount")
		raisedDate       = r.PostFormValue("raisedDate")
		startDate        = r.PostFormValue("startDate")
		endDate          = r.PostFormValue("endDate")
		supervisionLevel = r.PostFormValue("supervisionLevel")
		raisedDateYear   = r.PostFormValue("raisedDateYear")
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

func invoiceData(invoiceType string, amount string, startDate string, raisedDate string, endDate string, raisedDateYear string, supervisionLevel string) (string, string, string, string, string) {
	switch invoiceType {
	case shared.InvoiceTypeAD.Key():
		amount = "100"
		startDate = raisedDate
		endDate = raisedDate
	case shared.InvoiceTypeS2.Key(), shared.InvoiceTypeB2.Key():
		if raisedDateYear != "" {
			raisedDate = raisedDateYear + "-03" + "-31"
		}
		endDate = raisedDate
		supervisionLevel = "GENERAL"
	case shared.InvoiceTypeS3.Key(), shared.InvoiceTypeB3.Key():
		if raisedDateYear != "" {
			raisedDate = raisedDateYear + "-03" + "-31"
		}
		endDate = raisedDate
		supervisionLevel = "MINIMAL"
	case "":
		amount = "1"
	}
	return amount, startDate, raisedDate, endDate, supervisionLevel
}
