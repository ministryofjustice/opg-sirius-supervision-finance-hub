package server

import (
	"github.com/opg-sirius-finance-hub/shared"
	"net/http"
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

	data := AdjustInvoiceVars{&shared.AdjustmentTypes, r.PathValue("id"), r.PathValue("invoiceId"), v}

	return h.execute(w, r, data)
}
