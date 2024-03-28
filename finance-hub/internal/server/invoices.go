package server

import (
	"github.com/opg-sirius-finance-hub/finance-hub/internal/util"
	"github.com/opg-sirius-finance-hub/shared"
	"net/http"
)

type Invoices []Invoice

type Invoice struct {
	Id                 int
	Ref                string
	Status             string
	Amount             string
	RaisedDate         string
	Received           string
	OutstandingBalance string
	Ledgers            []shared.Ledger
	SupervisionLevels  []shared.SupervisionLevel
}

type InvoiceTab struct {
	Invoices Invoices
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

	data := &InvoiceTab{h.transform(invoices), v}
	data.selectTab("invoices")
	return h.execute(w, r, data)
}

func (h *InvoicesHandler) transform(in shared.Invoices) Invoices {
	var out Invoices
	for _, f := range in {
		out = append(out, Invoice{
			Id:                 f.Id,
			Ref:                f.Ref,
			Status:             f.Status,
			Amount:             util.IntToDecimalString(f.Amount),
			RaisedDate:         f.RaisedDate.String(),
			Received:           util.IntToDecimalString(f.Received),
			OutstandingBalance: util.IntToDecimalString(f.OutstandingBalance),
			Ledgers:            f.Ledgers,
			SupervisionLevels:  f.SupervisionLevels,
		})
	}
	return out
}
