package service

import (
	"context"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/allpay"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/auth"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/store"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
)

func (s *Service) ProcessFailedDirectDebitCollections(ctx context.Context, collectionDate time.Time) error {
	logger := s.Logger(ctx)

	fromDate, err := s.govUK.AddWorkingDays(ctx, collectionDate, 7)
	if err != nil {
		return err
	}

	payments, err := s.allpay.FetchFailedPayments(ctx, allpay.FetchFailedPaymentsInput{
		To:   collectionDate,
		From: fromDate,
	})
	if err != nil {
		return err
	}

	if len(payments) == 0 {
		logger.Info("no failed payments for collection date", "collectionDate", collectionDate)
		return nil
	}

	tx, err := s.BeginStoreTx(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	for _, payment := range payments {
		details := s.validateFailedCollectionLine(ctx, logger, payment)
		if details != nil {
			err = s.ProcessReversalUploadLine(ctx, tx, *details)
			logger.Error("process failed direct debit collections", "error", err)
		}
	}

	return tx.Commit(ctx)
}

func (s *Service) validateFailedCollectionLine(ctx context.Context, logger *slog.Logger, payment allpay.FailedPayment) *shared.ReversalDetails {
	var (
		courtRef       pgtype.Text
		collectionDate pgtype.Date
		collectionTime pgtype.Timestamp
		processedDate  pgtype.Date
		processedTime  pgtype.Timestamp
		createdBy      pgtype.Int4
		amount         int32
		notes          pgtype.Text
	)

	cd, _ := time.Parse("02/01/2006 15:04:05", payment.CollectionDate)
	pd, _ := time.Parse("02/01/2006 15:04:05", payment.ProcessedDate)

	_ = courtRef.Scan(payment.ClientReference)
	_ = store.ToInt4(&createdBy, ctx.(auth.Context).User.ID)
	_ = collectionDate.Scan(cd)
	_ = collectionTime.Scan(cd)
	_ = processedDate.Scan(pd)
	_ = processedTime.Scan(pd)
	amount = payment.Amount
	_ = notes.Scan(payment.ReasonCode)

	ledgerCount, _ := s.store.CountDuplicateLedger(ctx, store.CountDuplicateLedgerParams{
		CourtRef:     courtRef,
		Amount:       amount,
		Type:         shared.TransactionTypeDirectDebitPayment.Key(),
		BankDate:     collectionDate,
		ReceivedDate: collectionTime,
	})

	if ledgerCount == 0 {
		logger.Error("unable to match failed collection to original direct debit payment", "courtRef", courtRef, "collectionDate", collectionDate)
		return nil
	}

	reversalCount, _ := s.store.CountDuplicateLedger(ctx, store.CountDuplicateLedgerParams{
		CourtRef:     courtRef,
		Amount:       -amount,
		Type:         shared.TransactionTypeDirectDebitPayment.Key(),
		BankDate:     processedDate,
		ReceivedDate: processedTime,
	})

	if reversalCount >= ledgerCount {
		logger.Info("failed direct debit collection already reversed", "courtRef", courtRef, "collectionDate", collectionDate)
		return nil
	}

	return &shared.ReversalDetails{
		PaymentType:     shared.TransactionTypeDirectDebitPayment,
		ErroredCourtRef: courtRef,
		BankDate:        processedDate,
		ReceivedDate:    processedTime,
		Amount:          amount,
		CreatedBy:       createdBy,
		Notes:           notes,
	}
}
