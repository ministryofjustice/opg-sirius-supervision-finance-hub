package server

import (
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
	"net/http"
	"strconv"
)

type AddInvoiceAdjustmentForm struct {
	AdjustmentTypes *[]shared.AdjustmentType
	ClientId        string
	InvoiceId       string
	AppVars
}

type AddInvoiceAdjustmentFormHandler struct {
	router
}

func (h *AddInvoiceAdjustmentFormHandler) render(v AppVars, w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	clientID := getClientID(r)

	invoiceId, _ := strconv.Atoi(r.PathValue("invoiceId"))
	allowedAdjustments, err := h.Client().GetPermittedAdjustments(ctx, clientID, invoiceId)
	if err != nil {
		return err
	}
	data := AddInvoiceAdjustmentForm{&allowedAdjustments, r.PathValue("clientId"), r.PathValue("invoiceId"), v}

	return h.execute(w, r, data)
}
