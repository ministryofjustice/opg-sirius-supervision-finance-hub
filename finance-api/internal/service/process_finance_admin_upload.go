package service

import (
	"context"
	"encoding/csv"
	"fmt"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/opg-sirius-finance-hub/finance-api/internal/event"
	"github.com/opg-sirius-finance-hub/finance-api/internal/store"
	"os"
	"strconv"
	"strings"
	"time"
)

func (s *Service) ProcessFinanceAdminUpload(ctx context.Context, filename string, email string, reportType string) error {
	file, err := s.filestorage.GetFile(ctx, os.Getenv("ASYNC_S3_BUCKET"), filename)

	if err != nil {
		failedEvent := event.FinanceAdminUploadProcessed{EmailAddress: email, Error: "Unable to download report"}
		return s.dispatch.FinanceAdminUploadProcessed(ctx, failedEvent)
	}

	csvReader := csv.NewReader(file)
	records, err := csvReader.ReadAll()
	if err != nil {
		failedEvent := event.FinanceAdminUploadProcessed{EmailAddress: email, Error: "Unable to read report"}
		return s.dispatch.FinanceAdminUploadProcessed(ctx, failedEvent)
	}

	var failedLines map[int]string

	switch reportType {
	case "PAYMENTS_MOTO_CARD":
	case "PAYMENTS_ONLINE_CARD":
		failedLines, err = s.processMotoCardPayments(ctx, records)
	default:
		return fmt.Errorf("unknown report type: %s", reportType)
	}

	if err != nil {
		return err
	}

	if len(failedLines) > 0 {
		failedEvent := event.FinanceAdminUploadProcessed{EmailAddress: email, FailedLines: failedLines}
		return s.dispatch.FinanceAdminUploadProcessed(ctx, failedEvent)
	}

	return s.dispatch.FinanceAdminUploadProcessed(ctx, event.FinanceAdminUploadProcessed{EmailAddress: email})
}

func parseAmount(amount string) (int32, error) {
	index := strings.Index(amount, ".")

	if index != -1 && len(amount)-index == 2 {
		amount = amount + "0"
	} else if index == -1 {
		amount = amount + "00"
	}

	intAmount, err := strconv.Atoi(strings.Replace(amount, ".", "", 1))
	return int32(intAmount), err
}

func (s *Service) processMotoCardPayments(ctx context.Context, records [][]string) (map[int]string, error) {
	failedLines := make(map[int]string)

	for index, record := range records {
		if index != 0 {
			err := s.processMotoCardPaymentsUploadLine(ctx, record, index, &failedLines)
			if err != nil {
				return nil, err
			}
		}
	}

	return failedLines, nil
}

func (s *Service) processMotoCardPaymentsUploadLine(ctx context.Context, record []string, index int, failedLines *map[int]string) error {
	ctx, cancelTx := context.WithCancel(ctx)
	defer cancelTx()

	if record[0] == "" {
		return nil
	}

	courtReference := strings.SplitN(record[0], "-", -1)[0]

	amount, err := parseAmount(record[2])
	if err != nil {
		(*failedLines)[index] = "AMOUNT_PARSE_ERROR"
		return nil
	}

	parsedDate, err := time.Parse("2006-01-02 15:04:05", record[1])
	if err != nil {
		(*failedLines)[index] = "DATE_PARSE_ERROR"
		return nil
	}

	if amount == 0 {
		return nil
	}

	ledgerId, _ := s.store.GetLedgerForPayment(ctx, store.GetLedgerForPaymentParams{
		CourtRef: pgtype.Text{String: courtReference, Valid: true},
		Amount:   amount,
		Type:     "MOTO card payment",
		Datetime: pgtype.Timestamp{Time: parsedDate, Valid: true},
	})

	if ledgerId != 0 {
		(*failedLines)[index] = "DUPLICATE_PAYMENT"
		return nil
	}

	tx, err := s.tx.Begin(ctx)
	if err != nil {
		return err
	}

	transaction := s.store.WithTx(tx)

	ledgerId, err = transaction.CreateLedgerForCourtRef(ctx, store.CreateLedgerForCourtRefParams{
		CourtRef:  pgtype.Text{String: courtReference, Valid: true},
		Amount:    amount,
		Type:      "MOTO card payment",
		Status:    "CONFIRMED",
		CreatedBy: pgtype.Int4{Int32: 1, Valid: true},
		Datetime:  pgtype.Timestamp{Time: parsedDate, Valid: true},
	})

	if err != nil {
		(*failedLines)[index] = "CLIENT_NOT_FOUND"
		return nil
	}

	invoices, err := transaction.GetInvoicesForCourtRef(ctx, pgtype.Text{String: courtReference, Valid: true})

	if err != nil {
		return err
	}

	for _, invoice := range invoices {
		allocationAmount := invoice.Outstanding
		if allocationAmount > amount {
			allocationAmount = amount
		}

		err = transaction.CreateLedgerAllocation(ctx, store.CreateLedgerAllocationParams{
			InvoiceID: pgtype.Int4{Int32: invoice.ID, Valid: true},
			Amount:    allocationAmount,
			Status:    "ALLOCATED",
			LedgerID:  pgtype.Int4{Int32: ledgerId, Valid: true},
		})
		if err != nil {
			return err
		}

		amount -= allocationAmount
	}

	if amount > 0 {
		err = transaction.CreateLedgerAllocation(ctx, store.CreateLedgerAllocationParams{
			Amount:   amount,
			Status:   "UNAPPLIED",
			LedgerID: pgtype.Int4{Int32: ledgerId, Valid: true},
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
