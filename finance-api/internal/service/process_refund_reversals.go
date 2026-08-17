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

type refundReversalDetails struct {
	shared.PaymentDetails
	fulfilledDate pgtype.Date
}

type refundReversalUploadLine struct {
	index   int
	details refundReversalDetails
}

func (s *Service) ProcessRefundReversals(ctx context.Context, records [][]string, bankDate shared.Date) (map[int]string, error) {
	failedLines := make(map[int]string)
	clientLines := make(map[int32][]refundReversalUploadLine)

	for index, record := range records {
		if !isHeaderRow(shared.ReportTypeUploadReverseFulfilledRefunds, index) && safeRead(record, 0) != "" {
			details := getRefundReversalDetails(ctx, record, bankDate, index, &failedLines)

			if details != (refundReversalDetails{}) {
				if !s.validateRefundReversalLine(ctx, details, index, &failedLines) {
					continue
				}

				// we know client exists because validateRefundReversalLine has already checked
				client, _ := s.store.GetClientIdsByCourtRef(ctx, details.CourtRef)
				clientLines[client.ClientID] = append(clientLines[client.ClientID], refundReversalUploadLine{
					index:   index,
					details: details,
				})
			}
		}
	}

	for clientID, lines := range clientLines {
		s.processRefundReversalsForClient(ctx, clientID, lines, &failedLines)
	}

	return failedLines, nil
}

func (s *Service) processRefundReversalsForClient(ctx context.Context, clientID int32, lines []refundReversalUploadLine, failedLines *map[int]string) {
	tx, err := s.BeginStoreTxForClient(ctx, clientID)
	if err != nil {
		s.markRefundReversalLinesFailed(lines, failedLines, validation.UploadErrorProcessing)
		return
	}
	defer tx.Rollback(ctx)

	for _, line := range lines {
		_, err = s.ProcessPaymentsUploadLine(ctx, tx, line.details.PaymentDetails)
		if err != nil {
			s.markRefundReversalLinesFailed(lines, failedLines, validation.UploadErrorProcessing)
			return
		}
	}

	err = tx.Commit(ctx)
	if err != nil {
		s.markRefundReversalLinesFailed(lines, failedLines, validation.UploadErrorProcessing)
	}
}

func (s *Service) markRefundReversalLinesFailed(lines []refundReversalUploadLine, failedLines *map[int]string, reason string) {
	for _, line := range lines {
		if _, exists := (*failedLines)[line.index]; exists {
			continue
		}

		(*failedLines)[line.index] = reason
	}
}

func getRefundReversalDetails(ctx context.Context, record []string, formDate shared.Date, index int, failedLines *map[int]string) refundReversalDetails {
	var (
		courtRef      pgtype.Text
		bankDate      pgtype.Date
		fulfilledDate pgtype.Date
		receivedDate  pgtype.Timestamp
		createdBy     pgtype.Int4
		amount        int32
		err           error
	)

	if formDate.IsNull() {
		(*failedLines)[0] = validation.UploadErrorDateParse
	}
	_ = bankDate.Scan(formDate.Time)
	_ = receivedDate.Scan(formDate.Time)

	_ = store.ToInt4(&createdBy, ctx.(auth.Context).User.ID)

	_ = courtRef.Scan(safeRead(record, 0))

	refundDate, err := time.Parse("02/01/2006", safeRead(record, 2))
	if err != nil {
		(*failedLines)[index] = validation.UploadErrorDateParse
		return refundReversalDetails{}
	}
	_ = fulfilledDate.Scan(refundDate)

	amount, err = parseAmount(safeRead(record, 1))

	if err != nil {
		(*failedLines)[index] = validation.UploadErrorAmountParse
		return refundReversalDetails{}
	}

	return refundReversalDetails{
		PaymentDetails: shared.PaymentDetails{Amount: amount, BankDate: bankDate, CourtRef: courtRef, LedgerType: shared.TransactionTypeRefund, ReceivedDate: receivedDate, CreatedBy: createdBy},
		fulfilledDate:  fulfilledDate,
	}
}

func (s *Service) validateRefundReversalLine(ctx context.Context, details refundReversalDetails, index int, failedLines *map[int]string) bool {
	var amount pgtype.Int4
	_ = store.ToInt4(&amount, details.Amount)

	exists, err := s.store.CheckRefundForReversalExists(ctx, store.CheckRefundForReversalExistsParams{
		CourtRef:      details.CourtRef,
		FulfilledDate: details.fulfilledDate,
		Amount:        amount,
	})
	if err != nil {
		(*failedLines)[index] = validation.UploadErrorProcessing
		return false
	}

	if !exists {
		(*failedLines)[index] = validation.UploadErrorRefundForReversalNotFound
		return false
	}

	// check for already processed reversal
	return s.validatePaymentLine(ctx, details.PaymentDetails, index, failedLines)
}
