package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/ministryofjustice/opg-go-common/telemetry"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/auth"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/event"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/store"
	"log/slog"
)

func (s *Service) ReapplyCredit(ctx context.Context, clientID int32) error {
	var userID pgtype.Int4
	_ = store.ToInt4(&userID, ctx.(auth.Context).User.ID)

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

	ledgerID, err := s.store.CreateLedger(ctx, ledger)
	if err != nil {
		logger := telemetry.LoggerFromContext(ctx)
		logger.Error(fmt.Sprintf("Error in reapply for client %d", clientID), slog.String("err", err.Error()))
		return err
	}

	allocation.LedgerID = ledgerID

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
