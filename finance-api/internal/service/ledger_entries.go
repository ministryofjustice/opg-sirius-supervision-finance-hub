package service

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/auth"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/store"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
	"strings"
)

type addLedgerVars struct {
	amount             int32
	transactionType    shared.Enum
	feeReductionId     int32
	clientId           int32
	invoiceId          int32
	outstandingBalance int32
}

func generateLedgerEntries(ctx context.Context, vars addLedgerVars) (store.CreateLedgerParams, []store.CreateLedgerAllocationParams) {
	var (
		total     int32
		invoiceID pgtype.Int4
	)

	_ = store.ToInt4(&invoiceID, vars.invoiceId)

	allocations := []store.CreateLedgerAllocationParams{
		{
			InvoiceID: invoiceID,
			Amount:    vars.amount,
			Status:    "ALLOCATED",
			Notes:     pgtype.Text{},
		},
	}
	total += vars.amount

	diff := vars.outstandingBalance - vars.amount
	if diff < 0 {
		var notes pgtype.Text
		_ = notes.Scan(fmt.Sprintf("Unapplied funds as a result of applying %s", strings.ToLower(vars.transactionType.Key())))

		allocations = append(allocations, store.CreateLedgerAllocationParams{
			InvoiceID: invoiceID,
			Amount:    diff,
			Status:    "UNAPPLIED",
			Notes:     notes,
		})
		total += diff
	}

	var (
		createdBy pgtype.Int4
		notes     pgtype.Text
	)

	_ = notes.Scan("Credit due to approved " + strings.ToLower(vars.transactionType.Key()))
	_ = store.ToInt4(&createdBy, ctx.(auth.Context).User.ID)

	ledger := store.CreateLedgerParams{
		ClientID:       vars.clientId,
		Amount:         total,
		Notes:          notes,
		Type:           transformEnumToLedgerType(vars.transactionType),
		Status:         "CONFIRMED",
		FeeReductionID: pgtype.Int4{Int32: vars.feeReductionId, Valid: vars.feeReductionId != 0},
		CreatedBy:      createdBy,
	}

	return ledger, allocations
}

func transformEnumToLedgerType(e shared.Enum) string {
	if t, ok := e.(shared.FeeReductionType); ok {
		return "CREDIT " + t.Key()
	}
	return e.Key()
}
