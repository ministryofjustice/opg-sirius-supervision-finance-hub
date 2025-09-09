package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/auth"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/event"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/store"
)

func (s *Service) ReapplyCredit(ctx context.Context, clientID int32, tx *store.Tx) error {
	if tx == nil {
		return s.reapplyCreditTx(ctx, clientID)
	}

	logger := s.Logger(ctx)
	logger.Info(fmt.Sprintf("reapplying credit for client %d", clientID))

	var userID pgtype.Int4
	_ = store.ToInt4(&userID, ctx.(auth.Context).User.ID)
	creditPosition, err := tx.GetCreditBalanceAndOldestOpenInvoice(ctx, clientID)

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

	reapplyAmount := getReapplyAmount(creditPosition.Credit, creditPosition.Outstanding)
	allocation := store.CreateLedgerAllocationParams{
		InvoiceID: creditPosition.InvoiceID,
		Amount:    reapplyAmount,
		Status:    "REAPPLIED",
	}

	var notes pgtype.Text
	_ = notes.Scan("Excess credit applied to invoice")

	ledger := store.CreateLedgerParams{
		ClientID:  clientID,
		Amount:    reapplyAmount,
		Notes:     notes,
		Type:      "CREDIT REAPPLY",
		Status:    "CONFIRMED",
		CreatedBy: userID,
	}

	ledgerID, err := tx.CreateLedger(ctx, ledger)
	if err != nil {
		s.Logger(ctx).Error(fmt.Sprintf("Error in reapply for client %d", clientID), slog.String("err", err.Error()))
		return err
	}

	allocation.LedgerID = ledgerID

	err = tx.CreateLedgerAllocation(ctx, allocation)
	if err != nil {
		s.Logger(ctx).Error(fmt.Sprintf("Error create ledger allocation for client %d", clientID), slog.String("err", err.Error()))
		return err
	}

	logger.Info(fmt.Sprintf("%d credit applied to invoice %d", reapplyAmount, creditPosition.InvoiceID.Int32))

	// there may still be credit on account, so repeat to find the next applicable invoice
	return s.ReapplyCredit(ctx, clientID, tx)
}

func getReapplyAmount(credit int32, outstanding int32) int32 {
	if credit >= outstanding {
		return outstanding
	} else {
		return credit
	}
}

/**
 * reapplyCreditTx is a wrapper function to supply a transaction where ReapplyCredit is called without an existing transaction
 */
func (s *Service) reapplyCreditTx(ctx context.Context, clientID int32) error {
	tx, err := s.BeginStoreTx(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	err = s.ReapplyCredit(ctx, clientID, tx)
	if err != nil {
		return err
	}
	return tx.Commit(ctx)
}
