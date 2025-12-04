package service

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/auth"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/store"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/validation"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
)

func (s *Service) ProcessRefundReversals(ctx context.Context, records [][]string, bankDate shared.Date) (map[int]string, error) {
	failedLines := make(map[int]string)

	tx, err := s.BeginStoreTx(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	for index, record := range records {
		if !isHeaderRow(shared.ReportTypeUploadReverseFulfilledRefunds, index) && safeRead(record, 0) != "" {
			details := getRefundReversalDetails(ctx, record, bankDate, index, &failedLines)

			if details != (shared.PaymentDetails{}) {
				if !s.validateRefundReversalLine(ctx, details, index, &failedLines) {
					continue
				}

				_, err := s.ProcessPaymentsUploadLine(ctx, tx, details)
				if err != nil {
					return nil, err
				}
			}
		}
	}

	err = tx.Commit(ctx)
	if err != nil {
		return nil, err
	}

	return failedLines, nil
}

func getRefundReversalDetails(ctx context.Context, record []string, formDate shared.Date, index int, failedLines *map[int]string) shared.PaymentDetails {
	var (
		courtRef     pgtype.Text
		bankDate     pgtype.Date
		receivedDate pgtype.Timestamp
		createdBy    pgtype.Int4
		amount       int32
		err          error
	)

	if !formDate.IsNull() {
		_ = bankDate.Scan(formDate.Time)
	}
	_ = store.ToInt4(&createdBy, ctx.(auth.Context).User.ID)

	_ = courtRef.Scan(safeRead(record, 0))

	bd, err := time.Parse("02/01/2006", safeRead(record, 2))
	if err != nil {
		(*failedLines)[index] = validation.UploadErrorDateParse
		return shared.PaymentDetails{}
	}
	_ = bankDate.Scan(bd)
	_ = receivedDate.Scan(bd)

	amount, err = parseAmount(safeRead(record, 1))

	if err != nil {
		(*failedLines)[index] = validation.UploadErrorAmountParse
		return shared.PaymentDetails{}
	}

	return shared.PaymentDetails{Amount: amount, BankDate: bankDate, CourtRef: courtRef, LedgerType: shared.TransactionTypeRefund, ReceivedDate: receivedDate, CreatedBy: createdBy}
}

func (s *Service) validateRefundReversalLine(ctx context.Context, details shared.PaymentDetails, index int, failedLines *map[int]string) bool {
	var amount pgtype.Int4
	_ = store.ToInt4(&amount, details.Amount)

	exists, _ := s.store.CheckRefundForReversalExists(ctx, store.CheckRefundForReversalExistsParams{
		CourtRef:      details.CourtRef,
		FulfilledDate: details.BankDate,
		Amount:        amount,
	})

	if !exists {
		(*failedLines)[index] = validation.UploadErrorRefundForReversalNotFound
		return false
	}

	// check for already processed reversal
	// this doesn't work as it is looking for a negative amount?
	return s.validatePaymentLine(ctx, details, index, failedLines)
}
