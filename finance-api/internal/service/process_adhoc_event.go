package service

import (
	"context"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/auth"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/store"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
	"time"
)

func (s *Service) ProcessAdhocEvent(ctx context.Context) error {

	ctx, cancelTx := s.WithCancel(ctx)
	defer cancelTx()

	tx, err := s.BeginStoreTx(ctx)
	if err != nil {
		return err
	}

	invoices, err := tx.GetNegativeInvoices(ctx)
	if err != nil {
		return err
	}

	for _, invoice := range invoices {
		now := time.Now().UTC()
		todaysDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
		details := shared.PaymentDetails{
			Amount: int32(invoice.Ledgerallocationamountneeded),
			ReceivedDate: pgtype.Timestamp{
				Time:             todaysDate,
				InfinityModifier: 0,
				Valid:            true,
			},
			CourtRef:   pgtype.Text{String: invoice.CourtRef.String, Valid: true},
			LedgerType: shared.TransactionTypeUnappliedPayment,
			BankDate: pgtype.Date{
				Time:             todaysDate,
				InfinityModifier: 0,
				Valid:            true,
			},
			CreatedBy: pgtype.Int4{
				Int32: ctx.(auth.Context).User.ID,
				Valid: true,
			},
		}

		params := store.CreateLedgerForCourtRefParams{
			CourtRef:     details.CourtRef,
			Amount:       details.Amount,
			Type:         details.LedgerType.Key(),
			Status:       "CONFIRMED",
			CreatedBy:    details.CreatedBy,
			BankDate:     details.BankDate,
			ReceivedDate: details.ReceivedDate,
			PisNumber:    details.PisNumber,
		}

		ledgerID, err := tx.CreateLedgerForCourtRef(ctx, params)
		if err != nil {
			return err
		}

		var invoiceID pgtype.Int4
		_ = store.ToInt4(&invoiceID, invoice.ID)

		err = tx.CreateLedgerAllocation(ctx, store.CreateLedgerAllocationParams{
			InvoiceID: invoiceID,
			Amount:    details.Amount,
			Status:    "ALLOCATED",
			LedgerID:  ledgerID,
		})
		if err != nil {
			return err
		}
	}

	err = tx.Commit(ctx)
	if err != nil {
		return err
	}

	return nil
}
