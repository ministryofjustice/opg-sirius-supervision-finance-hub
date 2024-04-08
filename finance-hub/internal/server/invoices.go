package server

import (
	"github.com/opg-sirius-finance-hub/finance-hub/internal/util"
	"github.com/opg-sirius-finance-hub/shared"
	"net/http"
)

type Invoices []Invoice

type LedgerAllocation struct {
	Amount          string
	ReceivedDate    shared.Date
	TransactionType string
	Status          string
}

type SupervisionLevel struct {
	Level  string
	Amount string
	From   shared.Date
	To     shared.Date
}

type Invoice struct {
	Id                 int
	Ref                string
	Status             string
	Amount             string
	RaisedDate         string
	Received           string
	OutstandingBalance string
	Ledgers            []LedgerAllocation
	SupervisionLevels  []SupervisionLevel
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
			Ledgers:            transformLedgers(f.Ledgers),
			SupervisionLevels:  transformSupervisionLevels(f.SupervisionLevels),
		})
	}
	return out
}

func transformSupervisionLevels(in []shared.SupervisionLevel) []SupervisionLevel {
	var out []SupervisionLevel
	for _, f := range in {
		out = append(out, SupervisionLevel{
			Level:  f.Level,
			Amount: util.IntToDecimalString(f.Amount),
			From:   f.From,
			To:     f.To,
		})
	}
	return out
}

func transformLedgers(ledgers []shared.Ledger) []LedgerAllocation {
	var out []LedgerAllocation
	for _, l := range ledgers {
		out = append(out, LedgerAllocation{
			Amount:          util.IntToDecimalString(l.Amount),
			ReceivedDate:    l.ReceivedDate,
			TransactionType: l.TransactionType,
			Status:          l.Status,
		})
	}
	return out
}
