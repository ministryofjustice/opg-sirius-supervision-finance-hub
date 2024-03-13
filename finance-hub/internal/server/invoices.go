package server

import (
	"github.com/opg-sirius-finance-hub/shared"
	"net/http"
)

type InvoiceTab struct {
	Invoices shared.Invoices
	AppVars
}

type InvoicesHandler struct {
	router
}

func (h *InvoicesHandler) render(v AppVars, w http.ResponseWriter, r *http.Request) error {
	ctx := getContext(r)

	invoices, err := h.Client().GetInvoices(ctx, ctx.ClientId)
	if err != nil {
		return err
	}

	data := &InvoiceTab{invoices, v}
	data.selectTab("invoices")
	return h.execute(w, r, data)
}
