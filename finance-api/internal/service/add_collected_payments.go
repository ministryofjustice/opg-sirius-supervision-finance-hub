package service

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/auth"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/store"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
)

func (s *Service) AddCollectedPayments(ctx context.Context, date time.Time) error {
	logger := s.Logger(ctx)

	var (
		collectionDate pgtype.Date
		receivedDate   pgtype.Timestamp
		createdBy      pgtype.Int4
	)

	_ = collectionDate.Scan(date)
	_ = receivedDate.Scan(date)
	_ = store.ToInt4(&createdBy, ctx.(auth.Context).User.ID)

	pendingCollections, err := s.store.GetPendingCollectionsForDate(ctx, collectionDate)
	if err != nil {
		logger.Error("unable to get pending collections for date", "date", collectionDate, "error", err)
		return err
	}

	if len(pendingCollections) < 0 {
		logger.Info("no pending collections for date", "date", collectionDate)
		return nil
	}

	ctx, cancelTx := s.WithCancel(ctx)
	defer cancelTx()

	tx, err := s.BeginStoreTx(ctx)
	if err != nil {
		logger.Error("unable to begin store transaction", "error", err)
		return err
	}

	for _, pc := range pendingCollections {
		payment := shared.PaymentDetails{
			Amount:       pc.Amount,
			BankDate:     collectionDate,
			CourtRef:     pc.CourtRef,
			LedgerType:   shared.TransactionTypeDirectDebitPayment,
			ReceivedDate: receivedDate,
			CreatedBy:    createdBy,
		}
		paymentLedger, err := s.ProcessPaymentsUploadLine(ctx, tx, payment)
		if err != nil {
			logger.Error("unable to process collected payments upload line for date and court reference", "date", collectionDate, "courtRef", pc.CourtRef.String, "error", err)
			return err
		}
		var ledgerID pgtype.Int4
		_ = store.ToInt4(&ledgerID, paymentLedger)

		err = tx.MarkPendingCollectionAsCollected(ctx, store.MarkPendingCollectionAsCollectedParams{
			LedgerID: ledgerID,
			ID:       pc.ID,
		})
		if err != nil {
			logger.Error("unable to mark collected payments as collected", "id", pc.ID, "error", err)
			return err
		}
	}

	err = tx.Commit(ctx)
	if err != nil {
		logger.Error("unable to commit transaction", "error", err)
		return err
	}

	return nil
}
