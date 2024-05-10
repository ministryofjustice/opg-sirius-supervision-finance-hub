package server

import (
	"github.com/opg-sirius-finance-hub/finance-hub/internal/util"
	"github.com/opg-sirius-finance-hub/shared"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"net/http"
)

type Invoices []Invoice

type LedgerAllocations []LedgerAllocation

type LedgerAllocation struct {
	Amount          string
	ReceivedDate    shared.Date
	TransactionType string
	Status          string
}

type SupervisionLevels []SupervisionLevel

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
	Ledgers            LedgerAllocations
	SupervisionLevels  SupervisionLevels
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
	for _, invoice := range in {
		out = append(out, Invoice{
			Id:                 invoice.Id,
			Ref:                invoice.Ref,
			Status:             invoice.Status,
			Amount:             util.IntToDecimalString(invoice.Amount),
			RaisedDate:         invoice.RaisedDate.String(),
			Received:           util.IntToDecimalString(invoice.Received),
			OutstandingBalance: util.IntToDecimalString(invoice.OutstandingBalance),
			Ledgers:            h.transformLedgers(invoice.Ledgers),
			SupervisionLevels:  h.transformSupervisionLevels(invoice.SupervisionLevels),
		})
	}
	return out
}

func (h *InvoicesHandler) transformSupervisionLevels(in []shared.SupervisionLevel) SupervisionLevels {
	var out SupervisionLevels
	caser := cases.Title(language.English)
	for _, supervisionLevel := range in {
		out = append(out, SupervisionLevel{
			Level:  caser.String(supervisionLevel.Level),
			Amount: util.IntToDecimalString(supervisionLevel.Amount),
			From:   supervisionLevel.From,
			To:     supervisionLevel.To,
		})
	}
	return out
}

func (h *InvoicesHandler) transformLedgers(ledgers []shared.Ledger) LedgerAllocations {
	var out LedgerAllocations
	for _, ledger := range ledgers {
		out = append(out, LedgerAllocation{
			Amount:          util.IntToDecimalString(ledger.Amount),
			ReceivedDate:    ledger.ReceivedDate,
			TransactionType: ledger.TransactionType,
			Status:          ledger.Status,
		})
	}
	return out
}
