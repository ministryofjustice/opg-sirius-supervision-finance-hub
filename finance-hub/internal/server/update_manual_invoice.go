package server

import (
	"github.com/opg-sirius-finance-hub/shared"
	"net/http"
)

type UpdateManualInvoice struct {
	ClientId     string
	InvoiceTypes *[]shared.InvoiceType
	AppVars
}

type UpdateManualInvoiceHandler struct {
	router
}

func (h *UpdateManualInvoiceHandler) render(v AppVars, w http.ResponseWriter, r *http.Request) error {
	data := UpdateManualInvoice{r.PathValue("clientId"), &shared.InvoiceTypes, v}

	return h.execute(w, r, data)
}
