package server

import (
	"github.com/opg-sirius-finance-hub/shared"
	"net/http"
)

type UpdateInvoices struct {
	AdjustmentTypes []shared.AdjustmentType
	ClientId        string
	InvoiceId       string
	AppVars
}

type UpdateInvoiceHandler struct {
	router
}

func (h *UpdateInvoiceHandler) render(v AppVars, w http.ResponseWriter, r *http.Request) error {

	data := UpdateInvoices{shared.AdjustmentTypes, r.PathValue("id"), r.PathValue("invoiceId"), v}

	return h.execute(w, r, data)
}
