package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/ministryofjustice/opg-go-common/telemetry"
	"github.com/opg-sirius-finance-hub/finance-api/internal/store"
)

func (s *Service) reapplyCredit(ctx context.Context, clientID int32) error {
	creditPosition, err := s.store.GetCreditBalanceAndOldestOpenInvoice(ctx, clientID)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil
	} else if err != nil {
		return err
	} else if creditPosition.Credit < 1 {
		return nil
	}

	reapplyAmount := getReapplyAmount(creditPosition.Credit, creditPosition.Outstanding)
	allocation := store.CreateLedgerAllocationParams{

		InvoiceID: pgtype.Int4{Int32: creditPosition.InvoiceID, Valid: true},
		Amount:    reapplyAmount,
		Status:    "REAPPLIED",
		Notes:     pgtype.Text{},
	}

	ledger := store.CreateLedgerParams{
		ClientID: clientID,
		Amount:   reapplyAmount,
		Notes:    pgtype.Text{String: "Excess credit applied to invoice", Valid: true},
		Type:     "CREDIT REAPPLY",
		Status:   "CONFIRMED",
		//TODO when adding this in PFS-136, the id with need to be a system user
		CreatedbyID: pgtype.Int4{Int32: 1},
	}

	ledgerId, err := s.store.CreateLedger(ctx, ledger)
	if err != nil {
		logger := telemetry.LoggerFromContext(ctx)
		logger.Error(fmt.Sprintf("Error in reapply for client %d: %s", clientID, err.Error()))
		return err
	}

	allocation.LedgerID = pgtype.Int4{Int32: ledgerId, Valid: true}
	_, err = s.store.CreateLedgerAllocation(ctx, allocation)
	if err != nil {
		return err
	}

	// there may still be credit on account, so repeat to find the next applicable invoice
	return s.reapplyCredit(ctx, clientID)
}

func getReapplyAmount(credit int32, outstanding int32) int32 {
	if credit >= outstanding {
		return outstanding
	} else {
		return credit
	}
}
