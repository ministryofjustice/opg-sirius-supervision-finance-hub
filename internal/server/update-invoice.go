package server

import (
	"github.com/opg-sirius-finance-hub/internal/model"
	"net/http"
)

type UpdateInvoices struct {
	InvoiceTypes []model.InvoiceType
	InvoiceType  string
	Notes        string
	ClientId     string
	InvoiceId    string
	AppVars
}

type UpdateInvoiceHandler struct {
	router
}

func (h *UpdateInvoiceHandler) render(v AppVars, w http.ResponseWriter, r *http.Request) error {
	invoiceTypes := []model.InvoiceType{
		{Handle: "writeOff", Description: "Write off"},
		{Handle: "addCredit", Description: "Add credit"},
		{Handle: "addDebit", Description: "Add debit"},
		{Handle: "unapply", Description: "Unapply"},
		{Handle: "reapply", Description: "Reapply"},
	}
	data := UpdateInvoices{invoiceTypes, "", "", r.PathValue("id"), r.PathValue("invoiceId"), v}

	return h.execute(w, r, data)
}
