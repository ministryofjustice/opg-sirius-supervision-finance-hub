package service

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/auth"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/store"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/validation"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
)

type fulfilledRefundUploadLine struct {
	index   int
	details shared.FulfilledRefundDetails
}

func (s *Service) ProcessFulfilledRefunds(ctx context.Context, records [][]string, bankDate shared.Date) (map[int]string, error) {
	failedLines := make(map[int]string)
	clientLines := make(map[int32][]fulfilledRefundUploadLine)

	for index, record := range records {
		if !isHeaderRow(shared.ReportTypeUploadFulfilledRefunds, index) && safeRead(record, 0) != "" {
			details := getRefundDetails(ctx, record, bankDate, index, &failedLines)

			if details != (shared.FulfilledRefundDetails{}) {
				client, err := s.store.GetClientIdsByCourtRef(ctx, details.CourtRef)
				if errors.Is(err, pgx.ErrNoRows) || client.ClientID == 0 {
					failedLines[index] = validation.UploadErrorRefundNotFound
					continue
				}
				if err != nil {
					failedLines[index] = validation.UploadErrorProcessing
					continue
				}

				clientLines[client.ClientID] = append(clientLines[client.ClientID], fulfilledRefundUploadLine{
					index:   index,
					details: details,
				})
			}
		}
	}

	for clientID, lines := range clientLines {
		s.processFulfilledRefundsForClient(ctx, clientID, lines, &failedLines)
	}

	return failedLines, nil
}

func (s *Service) processFulfilledRefundsForClient(ctx context.Context, clientID int32, lines []fulfilledRefundUploadLine, failedLines *map[int]string) {
	tx, err := s.BeginStoreTxForClient(ctx, clientID)
	if err != nil {
		s.markFulfilledRefundLinesFailed(lines, failedLines, validation.UploadErrorProcessing)
		return
	}
	defer tx.Rollback(ctx)

	for _, line := range lines {
		refundID, err := tx.GetProcessingRefund(ctx, store.GetProcessingRefundParams{
			CourtRef:      line.details.CourtRef,
			Amount:        line.details.Amount,
			AccountName:   line.details.AccountName,
			AccountNumber: line.details.AccountNumber,
			SortCode:      line.details.SortCode,
		})
		if errors.Is(err, pgx.ErrNoRows) || refundID == 0 {
			(*failedLines)[line.index] = validation.UploadErrorRefundNotFound
			continue
		}
		if err != nil {
			s.markFulfilledRefundLinesFailed(lines, failedLines, validation.UploadErrorProcessing)
			return
		}

		err = s.ProcessFulfilledRefundsLine(ctx, tx, refundID, line.details)
		if err != nil {
			s.markFulfilledRefundLinesFailed(lines, failedLines, validation.UploadErrorProcessing)
			return
		}
	}

	err = tx.Commit(ctx)
	if err != nil {
		s.markFulfilledRefundLinesFailed(lines, failedLines, validation.UploadErrorProcessing)
	}
}

func (s *Service) markFulfilledRefundLinesFailed(lines []fulfilledRefundUploadLine, failedLines *map[int]string, reason string) {
	for _, line := range lines {
		if _, exists := (*failedLines)[line.index]; exists {
			continue
		}

		(*failedLines)[line.index] = reason
	}
}

func getRefundDetails(ctx context.Context, record []string, formDate shared.Date, index int, failedLines *map[int]string) shared.FulfilledRefundDetails {
	var (
		courtRef      pgtype.Text
		amount        pgtype.Int4
		accountName   pgtype.Text
		accountNumber pgtype.Text
		sortCode      pgtype.Text
		bankDate      pgtype.Date
		uploadedBy    pgtype.Int4
		err           error
	)

	_ = bankDate.Scan(formDate.Time)
	_ = store.ToInt4(&uploadedBy, ctx.(auth.Context).User.ID)
	_ = courtRef.Scan(safeRead(record, 0))
	a, err := parseAmount(safeRead(record, 1))
	if err != nil {
		(*failedLines)[index] = validation.UploadErrorAmountParse
		return shared.FulfilledRefundDetails{}
	}
	_ = store.ToInt4(&amount, a)

	_ = accountName.Scan(safeRead(record, 2))
	_ = accountNumber.Scan(safeRead(record, 3))
	_ = sortCode.Scan(safeRead(record, 4))

	return shared.FulfilledRefundDetails{
		CourtRef:      courtRef,
		Amount:        amount,
		AccountName:   accountName,
		AccountNumber: accountNumber,
		SortCode:      sortCode,
		BankDate:      bankDate,
		UploadedBy:    uploadedBy,
	}
}

func (s *Service) ProcessFulfilledRefundsLine(ctx context.Context, tx *store.Tx, refundID int32, details shared.FulfilledRefundDetails) error {
	var now pgtype.Timestamp
	_ = now.Scan(time.Now())

	params := store.CreateLedgerForCourtRefParams{
		CourtRef:     details.CourtRef,
		Amount:       -details.Amount.Int32,
		Type:         shared.TransactionTypeRefund.Key(),
		Status:       "CONFIRMED",
		CreatedBy:    details.UploadedBy,
		BankDate:     details.BankDate,
		ReceivedDate: now,
	}
	ledgerID, err := tx.CreateLedgerForCourtRef(ctx, params)

	if err != nil {
		return err
	}

	err = tx.CreateLedgerAllocation(ctx, store.CreateLedgerAllocationParams{
		Amount:   details.Amount.Int32,
		Status:   "REAPPLIED",
		LedgerID: ledgerID,
	})
	if err != nil {
		return err
	}

	err = tx.MarkRefundsAsFulfilled(ctx, refundID)
	if err != nil {
		return err
	}

	return tx.RemoveBankDetails(ctx, refundID)
}
