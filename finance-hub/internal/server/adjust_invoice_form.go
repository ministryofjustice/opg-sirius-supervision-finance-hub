package server

import (
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
	"net/http"
	"strconv"
)

type AdjustInvoiceVars struct {
	AdjustmentTypes *[]shared.AdjustmentType
	ClientId        string
	InvoiceId       string
	AppVars
}

type AdjustInvoiceFormHandler struct {
	router
}

func (h *AdjustInvoiceFormHandler) render(v AppVars, w http.ResponseWriter, r *http.Request) error {
	ctx := getContext(r)

	clientId, _ := strconv.Atoi(r.PathValue("clientId"))
	invoiceId, _ := strconv.Atoi(r.PathValue("invoiceId"))
	allowedAdjustments, err := h.Client().GetPermittedAdjustments(ctx, clientId, invoiceId)
	if err != nil {
		return err
	}
	data := AdjustInvoiceVars{&allowedAdjustments, r.PathValue("clientId"), r.PathValue("invoiceId"), v}

	return h.execute(w, r, data)
}
