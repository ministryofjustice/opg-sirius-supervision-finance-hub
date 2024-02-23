package server

import (
	"errors"
	"github.com/opg-sirius-finance-hub/internal/model"
	"net/http"
	"strconv"
)

type InvoiceTab struct {
	AppVars
	Invoices model.InvoiceList
}

type InvoicesHandler struct {
	route
}

func (h *InvoicesHandler) render(v AppVars, w http.ResponseWriter, r *http.Request) error {
	var data InvoiceTab
	ctx := getContext(r)
	clientId, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		return errors.New("client id in string cannot be parsed to an integer")
	}

	invoices, err := (h.route.client).GetInvoices(ctx, clientId)
	if err != nil {
		return err
	}

	data.AppVars = v
	data.Invoices = invoices

	h.Data = data
	return h.execute(w, r)
}
