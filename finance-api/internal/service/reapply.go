package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/ministryofjustice/opg-go-common/telemetry"
	"github.com/opg-sirius-finance-hub/finance-api/internal/event"
	"github.com/opg-sirius-finance-hub/finance-api/internal/store"
	"log/slog"
)

func (s *Service) ReapplyCredit(ctx context.Context, clientID int32) error {
	creditPosition, err := s.store.GetCreditBalanceAndOldestOpenInvoice(ctx, clientID)

	switch {
	case errors.Is(err, pgx.ErrNoRows):
		return nil
	case err != nil:
		return err
	case creditPosition.Credit < 1:
		return nil
	case !creditPosition.InvoiceID.Valid:
		return s.dispatch.CreditOnAccount(ctx, event.CreditOnAccount{
			ClientID:        int(clientID),
			CreditRemaining: int(creditPosition.Credit),
		})
	}

	reapplyAmount := getReapplyAmount(creditPosition.Credit, creditPosition.Outstanding.Int32)
	allocation := store.CreateLedgerAllocationParams{
		InvoiceID: creditPosition.InvoiceID,
		Amount:    reapplyAmount,
		Status:    "REAPPLIED",
	}

	ledger := store.CreateLedgerParams{
		ClientID: clientID,
		Amount:   reapplyAmount,
		Notes:    pgtype.Text{String: "Excess credit applied to invoice", Valid: true},
		Type:     "CREDIT REAPPLY",
		Status:   "CONFIRMED",
		//TODO when adding this in PFS-136, the id with need to be a system user
		CreatedBy: pgtype.Int4{Int32: 1, Valid: true},
	}

	ledgerId, err := s.store.CreateLedger(ctx, ledger)
	if err != nil {
		logger := telemetry.LoggerFromContext(ctx)
		logger.Error(fmt.Sprintf("Error in reapply for client %d", clientID), slog.String("err", err.Error()))
		return err
	}

	allocation.LedgerID = pgtype.Int4{Int32: ledgerId, Valid: true}
	err = s.store.CreateLedgerAllocation(ctx, allocation)
	if err != nil {
		return err
	}

	// there may still be credit on account, so repeat to find the next applicable invoice
	return s.ReapplyCredit(ctx, clientID)
}

func getReapplyAmount(credit int32, outstanding int32) int32 {
	if credit >= outstanding {
		return outstanding
	} else {
		return credit
	}
}
