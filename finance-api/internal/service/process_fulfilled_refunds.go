package service

import (
	"context"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/auth"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/store"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/validation"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
	"time"
)

func (s *Service) ProcessFulfilledRefunds(ctx context.Context, records [][]string, bankDate shared.Date) (map[int]string, error) {
	failedLines := make(map[int]string)

	ctx, cancelTx := s.WithCancel(ctx)
	defer cancelTx()

	tx, err := s.BeginStoreTx(ctx)
	if err != nil {
		return nil, err
	}

	for index, record := range records {
		if !isHeaderRow(shared.ReportTypeUploadFulfilledRefunds, index) && safeRead(record, 0) != "" {
			details := getRefundDetails(ctx, record, bankDate, index, &failedLines)

			if details != (shared.FulfilledRefundDetails{}) {
				id, _ := tx.GetProcessingRefund(ctx, store.GetProcessingRefundParams{
					CourtRef:      details.CourtRef,
					Amount:        details.Amount,
					AccountName:   details.AccountName,
					AccountNumber: details.AccountNumber,
					SortCode:      details.SortCode,
				})
				if id == 0 {
					failedLines[index] = validation.UploadErrorRefundNotFound
					continue
				}

				err := s.ProcessFulfilledRefundsLine(ctx, tx, id, details)
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
		Amount:       details.Amount.Int32,
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
