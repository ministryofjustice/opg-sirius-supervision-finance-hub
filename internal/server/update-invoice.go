package server

import (
	"github.com/opg-sirius-finance-hub/internal/model"
	"net/http"
)

type UpdateInvoices struct {
	InvoiceTypes []model.InvoiceType
	ClientId     string
	InvoiceId    string
	AppVars
}

type UpdateInvoiceHandler struct {
	router
}

func (h *UpdateInvoiceHandler) render(v AppVars, w http.ResponseWriter, r *http.Request) error {

	data := UpdateInvoices{model.InvoiceTypes, r.PathValue("id"), r.PathValue("invoiceId"), v}

	return h.execute(w, r, data)
}
