package server

import (
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"

	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
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
	ClientId           int
}

type InvoicesVars struct {
	Invoices Invoices
	ClientId string
	AppVars
}

type InvoicesHandler struct {
	router
}

func (h *InvoicesHandler) render(v AppVars, w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	clientID := getClientID(r)

	invoices, err := h.Client().GetInvoices(ctx, clientID)
	if err != nil {
		return err
	}

	data := &InvoicesVars{h.transform(invoices, clientID), strconv.Itoa(clientID), v}
	data.selectTab("invoices")
	fmt.Printf("in invoices render")

	return h.execute(w, r, data)
}

func (h *InvoicesHandler) transform(in shared.Invoices, clientId int) Invoices {
	var out Invoices
	caser := cases.Title(language.English)

	for _, invoice := range in {
		out = append(out, Invoice{
			Id:                 invoice.Id,
			Ref:                invoice.Ref,
			Status:             caser.String(invoice.Status),
			Amount:             shared.IntToDecimalString(invoice.Amount),
			RaisedDate:         invoice.RaisedDate.String(),
			Received:           shared.IntToDecimalString(invoice.Received),
			OutstandingBalance: shared.IntToDecimalString(invoice.OutstandingBalance),
			Ledgers:            h.transformLedgers(invoice.Ledgers, caser),
			SupervisionLevels:  h.transformSupervisionLevels(invoice.SupervisionLevels, caser),
			ClientId:           clientId,
		})
	}
	return out
}

func (h *InvoicesHandler) transformSupervisionLevels(in []shared.SupervisionLevel, caser cases.Caser) SupervisionLevels {
	var out SupervisionLevels
	for _, supervisionLevel := range in {
		out = append(out, SupervisionLevel{
			Level:  caser.String(supervisionLevel.Level),
			Amount: shared.IntToDecimalString(supervisionLevel.Amount),
			From:   supervisionLevel.From,
			To:     supervisionLevel.To,
		})
	}
	return out
}

func (h *InvoicesHandler) transformLedgers(ledgers []shared.Ledger, caser cases.Caser) LedgerAllocations {
	var out LedgerAllocations
	for _, ledger := range ledgers {
		out = append(out, LedgerAllocation{
			Amount:          shared.IntToDecimalString(int(math.Abs(float64(ledger.Amount)))),
			ReceivedDate:    ledger.ReceivedDate,
			TransactionType: translate(ledger.TransactionType, ledger.Status, ledger.Amount),
			Status:          caser.String(ledger.Status),
		})
	}
	return out
}

func translate(transactionType string, status string, amount int) string {
	switch status {
	case "UNAPPLIED":
		return "Unapplied Payment"
	case "REAPPLIED":
		return "Reapplied Payment"
	}

	switch transactionType {
	case shared.AdjustmentTypeWriteOff.Key():
		return "Write Off"
	case shared.AdjustmentTypeCreditMemo.Key():
		return "Manual Credit"
	case shared.AdjustmentTypeDebitMemo.Key():
		return "Manual Debit"
	}

	parsedTransactionType := shared.ParseTransactionType(transactionType)
	if parsedTransactionType != shared.TransactionTypeUnknown {
		if amount < 0 && !parsedTransactionType.IsReversalType() { // don't need to rename reversal types
			return fmt.Sprintf("%s reversal", parsedTransactionType.String())
		}
		return parsedTransactionType.String()
	}

	caser := cases.Title(language.English)
	words := strings.Fields(transactionType)
	for i, word := range words {
		if strings.ToUpper(word) != word {
			words[i] = caser.String(word)
		}
	}
	return strings.Join(words, " ")
}
