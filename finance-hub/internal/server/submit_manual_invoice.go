package server

import (
	"errors"
	"fmt"
	"github.com/opg-sirius-finance-hub/finance-hub/internal/util"
	"github.com/opg-sirius-finance-hub/shared"
	"net/http"
	"strings"
	"time"
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
		raisedDateDay    = r.PostFormValue("raised-date-day")
		raisedDateMonth  = r.PostFormValue("raised-date-month")
		raisedDateYear   = r.PostFormValue("raised-date-year")
	)

	switch invoiceType {
	case shared.InvoiceTypeAD.Key():
		amount = "100"
		startDate = raisedDate
		endDate = calculateOneYearMinusOneDay(startDate)
	case shared.InvoiceTypeS2.Key(), shared.InvoiceTypeB2.Key():
		raisedDate = raisedDateYear + "-0" + raisedDateMonth + "-" + raisedDateDay
		endDate = raisedDate
		supervisionLevel = "GENERAL"
	case shared.InvoiceTypeS3.Key(), shared.InvoiceTypeB3.Key():
		raisedDate = raisedDateYear + "-0" + raisedDateMonth + "-" + raisedDateDay
		endDate = raisedDate
		supervisionLevel = "MINIMAL"
	}

	err := h.Client().AddManualInvoice(ctx, ctx.ClientId, invoiceType, amount, raisedDate, startDate, endDate, supervisionLevel)
	var verr shared.ValidationError
	if errors.As(err, &verr) {
		data := AppVars{Errors: util.RenameErrors(verr.Errors)}
		data.Errors = util.RenameErrors(verr.Errors)
		w.WriteHeader(http.StatusUnprocessableEntity)
		return h.execute(w, r, data)
	}

	if err != nil {
		return err
	}

	w.Header().Add("HX-Redirect", fmt.Sprintf("%s/clients/%d/manual-invoice?success=invoice-type[%s]", v.EnvironmentVars.Prefix, ctx.ClientId, strings.ToUpper(invoiceType)))
	return nil
}

func calculateOneYearMinusOneDay(startDate string) string {
	if startDate == "" {
		return ""
	}
	inputDate, _ := time.Parse("2006-01-02", startDate)
	newDate := inputDate.AddDate(1, 0, 0)

	oneDay := 24 * time.Hour
	newDate = newDate.Add(-oneDay)

	return newDate.Format("2006-01-02")
}
