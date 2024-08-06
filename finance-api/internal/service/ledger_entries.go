package service

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/opg-sirius-finance-hub/finance-api/internal/store"
	"github.com/opg-sirius-finance-hub/shared"
	"strings"
)

func addLedger(ctx context.Context, transaction *store.Queries, amount int32, feeReductionFeeType shared.FeeReductionType, feeReductionId int32, clientId int32, invoiceId int32, outstandingBalance int32) error {
	var total int32
	allocations := []store.CreateLedgerAllocationParams{
		{
			InvoiceID: pgtype.Int4{Int32: invoiceId, Valid: true},
			Amount:    amount,
			Status:    "ALLOCATED",
			Notes:     pgtype.Text{},
		},
	}
	total += amount

	diff := outstandingBalance - amount
	if diff < 0 {
		allocations = append(allocations, store.CreateLedgerAllocationParams{
			InvoiceID: pgtype.Int4{Int32: invoiceId, Valid: true},
			Amount:    diff,
			Status:    "UNAPPLIED",
			Notes:     pgtype.Text{String: fmt.Sprintf("Unapplied funds as a result of applying %s credit", strings.ToLower(feeReductionFeeType.Key())), Valid: true},
		})
		total += diff
	}

	ledgerParams := store.CreateLedgerParams{
		ClientID:       clientId,
		Amount:         total,
		Notes:          pgtype.Text{String: "Credit due to " + strings.ToLower(feeReductionFeeType.Key()), Valid: true},
		Type:           "CREDIT " + feeReductionFeeType.Key(),
		Status:         "APPROVED",
		FeeReductionID: pgtype.Int4{Int32: feeReductionId, Valid: true},
		//TODO make sure we have correct createdby ID in ticket PFS-136
		CreatedbyID: pgtype.Int4{Int32: 1},
	}

	ledgerId, err := transaction.CreateLedger(ctx, ledgerParams)
	if err != nil {
		return err
	}

	for _, allocation := range allocations {
		allocation.LedgerID = pgtype.Int4{Int32: ledgerId, Valid: true}
		_, err = transaction.CreateLedgerAllocation(ctx, allocation)
		if err != nil {
			return err
		}
	}

	return nil
}
