package service

import (
	"fmt"
	"github.com/jackc/pgx/v5/pgtype"
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

func generateLedgerEntries(vars addLedgerVars) (store.CreateLedgerParams, []store.CreateLedgerAllocationParams) {
	var total int32
	allocations := []store.CreateLedgerAllocationParams{
		{
			InvoiceID: pgtype.Int4{Int32: vars.invoiceId, Valid: true},
			Amount:    vars.amount,
			Status:    "ALLOCATED",
			Notes:     pgtype.Text{},
		},
	}
	total += vars.amount

	diff := vars.outstandingBalance - vars.amount
	if diff < 0 {
		allocations = append(allocations, store.CreateLedgerAllocationParams{
			InvoiceID: pgtype.Int4{Int32: vars.invoiceId, Valid: true},
			Amount:    diff,
			Status:    "UNAPPLIED",
			Notes:     pgtype.Text{String: fmt.Sprintf("Unapplied funds as a result of applying %s", strings.ToLower(vars.transactionType.Key())), Valid: true},
		})
		total += diff
	}

	ledger := store.CreateLedgerParams{
		ClientID:       vars.clientId,
		Amount:         total,
		Notes:          pgtype.Text{String: "Credit due to approved " + strings.ToLower(vars.transactionType.Key()), Valid: true},
		Type:           transformEnumToLedgerType(vars.transactionType),
		Status:         "APPROVED",
		FeeReductionID: pgtype.Int4{Int32: vars.feeReductionId, Valid: vars.feeReductionId != 0},
		//TODO make sure we have correct createdby ID in ticket PFS-136
		CreatedBy: pgtype.Int4{Int32: 1, Valid: true},
	}

	return ledger, allocations
}

func transformEnumToLedgerType(e shared.Enum) string {
	if t, ok := e.(shared.FeeReductionType); ok {
		return "CREDIT " + t.Key()
	}
	return e.Key()
}
