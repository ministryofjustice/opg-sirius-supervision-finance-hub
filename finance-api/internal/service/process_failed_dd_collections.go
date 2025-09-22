package service

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/allpay"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/auth"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/store"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/validation"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
)

func (s *Service) ProcessFailedDirectDebitCollections(ctx context.Context, date time.Time) error {
	logger := s.Logger(ctx)

	tx, err := s.BeginStoreTx(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	payments, err := s.allpay.FetchFailedPayments(ctx, allpay.FetchFailedPaymentsInput{
		To:   date,
		From: date,
	})
	if err != nil {
		return err
	}

	// TODO: Remove duplicates?

	for _, payment := range payments {
		var (
			courtRef       pgtype.Text
			collectionDate pgtype.Date
			collectionTime pgtype.Timestamp
			createdBy      pgtype.Int4
		)

		_ = courtRef.Scan(payment.ClientReference)
		_ = store.ToInt4(&createdBy, ctx.(auth.Context).User.ID)
		_ = collectionDate.Scan(payment.CollectionDate)
		_ = collectionTime.Scan(payment.CollectionDate)

		details := shared.ReversalDetails{
			PaymentType:     shared.TransactionTypeDirectDebitPayment,
			ErroredCourtRef: courtRef,
			BankDate:        collectionDate,
			ReceivedDate:    collectionTime, // TODO: Does this need to be the processedDate?
			Amount:          int32(payment.Amount),
			CreatedBy:       createdBy,
			SkipBankDate:    pgtype.Bool{Bool: true, Valid: true},
		}

		// TODO: Does this need validating? - yes

		err = s.ProcessReversalUploadLine(ctx, tx, details)
		logger.Error("process failed direct debit collections", "error", err) // TODO: Need to work out how to alert on these errors
	}

	return nil
}

func (s *Service) validateFailedCollectionLine(ctx context.Context, details shared.ReversalDetails, uploadType shared.ReportUploadType, processedRecords []shared.ReversalDetails, index int, failedLines *map[int]string) bool {
	ledgerCount, _ := s.store.CountDuplicateLedger(ctx, store.CountDuplicateLedgerParams{
		CourtRef:     details.ErroredCourtRef,
		Amount:       details.Amount,
		Type:         details.PaymentType.Key(),
		BankDate:     details.BankDate,
		ReceivedDate: details.ReceivedDate,
		PisNumber:    details.PisNumber,
		SkipBankDate: details.SkipBankDate,
	})

	if ledgerCount == 0 {
		(*failedLines)[index] = validation.UploadErrorNoMatchedPayment
		return false
	}

	reversalCount, _ := s.store.CountDuplicateLedger(ctx, store.CountDuplicateLedgerParams{
		CourtRef:     details.ErroredCourtRef,
		Amount:       -details.Amount,
		Type:         details.PaymentType.Key(),
		BankDate:     details.BankDate,
		ReceivedDate: details.ReceivedDate,
		PisNumber:    details.PisNumber,
	})

	if reversalCount >= ledgerCount {
		(*failedLines)[index] = validation.UploadErrorDuplicateReversal
		return false
	}

	if !hasPaymentToReverse(processedRecords, details, int(ledgerCount)) { // TODO: Not needed if duplicates removed?
		(*failedLines)[index] = validation.UploadErrorDuplicateReversal
		return false
	}

	return true
}
