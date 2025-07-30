package server

import (
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
	"net/http"
)

type AddManualInvoiceForm struct {
	ClientId     string
	InvoiceTypes *[]shared.InvoiceType
	AppVars
}

type AddManualInvoiceHandler struct {
	router
}

func (h *AddManualInvoiceHandler) render(v AppVars, w http.ResponseWriter, r *http.Request) error {
	data := AddManualInvoiceForm{r.PathValue("clientId"), &shared.InvoiceTypes, v}

	return h.execute(w, r, data)
}
