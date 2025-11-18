package service

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/allpay"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/auth"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/event"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/store"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
)

func (s *Service) ProcessFailedDirectDebitCollections(ctx context.Context, collectionDate time.Time) error {
	logger := s.Logger(ctx)

	fromDate, err := s.govUK.SubWorkingDays(ctx, collectionDate, 7)
	if err != nil {
		return err
	}

	payments, err := s.allpay.FetchFailedPayments(ctx, allpay.FetchFailedPaymentsInput{
		From: fromDate,
		To:   collectionDate,
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

	processed := make(map[int32]allpay.FailedPayment, len(payments))

	for _, payment := range payments {
		details := s.validateFailedCollectionLine(ctx, payment)
		if details != nil {
			err = s.ProcessReversalUploadLine(ctx, tx, *details)
			if err != nil {
				logger.Error(fmt.Sprintf("error processing failed direct debit collection for client reference %s on collection date %s", payment.ClientReference, payment.CollectionDate), "error", err)
				continue
			}

			var courtRef pgtype.Text
			_ = courtRef.Scan(payment.ClientReference)
			client, _ := s.store.GetClientByCourtRef(ctx, courtRef) // unchecked errors as inputs have already been confirmed via payment processing
			processed[client.ClientID] = payment
		}
	}

	err = tx.Commit(ctx)
	if err != nil {
		logger.Error("error committing failed direct debit collections", "error", err)
		return err
	}

	for clientID, payment := range processed {
		err = s.dispatch.DirectDebitCollectionFailed(ctx, event.DirectDebitCollectionFailed{
			ClientID: int(clientID),
			Reason:   payment.ReasonCode,
		})
		if err != nil {
			logger.Error("error dispatching \"direct-debit-collection-failed\" event", "error", err)
		}
	}

	return nil
}

func (s *Service) validateFailedCollectionLine(ctx context.Context, payment allpay.FailedPayment) *shared.ReversalDetails {
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
		s.Logger(ctx).Error("unable to match failed collection to original direct debit payment", "courtRef", courtRef, "collectionDate", collectionDate)
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
		s.Logger(ctx).Info("failed direct debit collection already reversed", "courtRef", courtRef, "collectionDate", collectionDate)
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
