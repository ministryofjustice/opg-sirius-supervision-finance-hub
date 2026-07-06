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

func (s *Service) PostLedgerActions(ctx context.Context, clientID int32, tx *store.Tx) error {
	if tx == nil {
		return s.postLedgerActionsTx(ctx, clientID)
	}
	return s.postLedgerActions(ctx, clientID, tx)
}

func (s *Service) postLedgerActions(ctx context.Context, clientID int32, tx *store.Tx) error {
	err := s.reapplyCredit(ctx, clientID, tx)
	if err != nil {
		return err
	}
	return s.resetRefunds(ctx, clientID, tx)
}

func (s *Service) reapplyCredit(ctx context.Context, clientID int32, tx *store.Tx) error {
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

	logger := s.Logger(ctx)
	logger.Info(fmt.Sprintf("reapplying credit for client %d", clientID))

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
	return s.PostLedgerActions(ctx, clientID, tx)
}

func getReapplyAmount(credit int32, outstanding int32) int32 {
	if credit >= outstanding {
		return outstanding
	}

	return credit
}

func (s *Service) resetRefunds(ctx context.Context, clientID int32, tx *store.Tx) error {
	refunds, err := tx.ResetApprovedRefunds(ctx, clientID)
	if err != nil {
		return err
	}
	if len(refunds) > 0 {
		return s.dispatch.RefundReset(ctx, event.RefundReset{ClientID: clientID})
	}
	return nil

}

/**
 * postLedgerActionsTx is a wrapper function to supply a transaction where PostLedgerActions is called without an existing transaction
 */
func (s *Service) postLedgerActionsTx(ctx context.Context, clientID int32) error {
	tx, err := s.BeginStoreTx(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	err = s.postLedgerActions(ctx, clientID, tx)
	if err != nil {
		return err
	}
	return tx.Commit(ctx)
}
