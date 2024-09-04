package service

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/ministryofjustice/opg-go-common/telemetry"
	"github.com/opg-sirius-finance-hub/finance-api/internal/store"
	"github.com/opg-sirius-finance-hub/shared"
)

func (s *Service) AddInvoiceAdjustment(ctx context.Context, clientId int, invoiceId int, ledgerEntry *shared.AddInvoiceAdjustmentRequest) (*shared.InvoiceReference, error) {
	logger := telemetry.LoggerFromContext(ctx)

	ctx, cancelTx := context.WithCancel(ctx)
	defer cancelTx()

	tx, err := s.tx.Begin(ctx)
	if err != nil {
		return nil, err
	}

	tStore := s.store.WithTx(tx)

	balance, err := tStore.GetInvoiceBalanceDetails(ctx, int32(invoiceId))
	if err != nil {
		return nil, err
	}

	err = s.validateAdjustmentAmount(ledgerEntry, balance)
	if err != nil {
		return nil, err
	}

	ledgerParams := store.CreateLedgerParams{
		ClientID:       int32(clientId),
		Amount:         s.calculateAdjustmentAmount(ledgerEntry, balance),
		Notes:          pgtype.Text{String: ledgerEntry.AdjustmentNotes, Valid: true},
		Type:           ledgerEntry.AdjustmentType.Key(),
		Status:         "PENDING",
		FeeReductionID: pgtype.Int4{},
		//TODO make sure we have correct createdby ID in ticket PFS-136
		CreatedBy: pgtype.Int4{Int32: 1, Valid: true},
	}

	ledgerId, err := tStore.CreateLedger(ctx, ledgerParams)
	if err != nil {
		logger.Error("Error creating ledger: ", err)
		return nil, err
	}

	allocationParams := store.CreateLedgerAllocationParams{
		LedgerID:  pgtype.Int4{Int32: ledgerId, Valid: true},
		InvoiceID: pgtype.Int4{Int32: int32(invoiceId), Valid: true},
		Amount:    ledgerParams.Amount,
		Status:    "PENDING",
		Notes:     pgtype.Text{String: ledgerEntry.AdjustmentNotes, Valid: true},
	}

	invoiceReference, err := tStore.CreateLedgerAllocation(ctx, allocationParams)
	if err != nil {
		logger.Error("Error creating ledger allocation: ", err)
		return nil, err
	}

	return &shared.InvoiceReference{InvoiceRef: invoiceReference}, nil
}

func (s *Service) validateAdjustmentAmount(adjustment *shared.AddInvoiceAdjustmentRequest, balance store.GetInvoiceBalanceDetailsRow) error {
	switch adjustment.AdjustmentType {
	case shared.AdjustmentTypeCreditMemo:
		if int32(adjustment.Amount)-balance.Outstanding > balance.Initial {
			return shared.BadRequest{Field: "Amount", Reason: fmt.Sprintf("Amount entered must be equal to or less than £%s", shared.IntToDecimalString(int(balance.Initial+balance.Outstanding)))}
		}
	case shared.AdjustmentTypeDebitMemo:
		var maxBalance int32
		if balance.Feetype == "AD" {
			maxBalance = 10000
		} else {
			maxBalance = 32000
		}
		if int32(adjustment.Amount)+balance.Outstanding > maxBalance {
			return shared.BadRequest{Field: "Amount", Reason: fmt.Sprintf("Amount entered must be equal to or less than £%s", shared.IntToDecimalString(int(maxBalance-balance.Outstanding)))}
		}
	case shared.AdjustmentTypeWriteOff:
		if balance.Outstanding < 1 {
			return shared.BadRequest{Field: "Amount", Reason: "No outstanding balance to write off"}
		}
	default:
		return shared.BadRequest{Field: "AdjustmentType", Reason: "Unimplemented adjustment type"}
	}
	return nil
}

func (s *Service) calculateAdjustmentAmount(adjustment *shared.AddInvoiceAdjustmentRequest, balance store.GetInvoiceBalanceDetailsRow) int32 {
	switch adjustment.AdjustmentType {
	case shared.AdjustmentTypeWriteOff:
		return balance.Outstanding
	case shared.AdjustmentTypeDebitMemo:
		return -int32(adjustment.Amount)
	default:
		return int32(adjustment.Amount)
	}
}
