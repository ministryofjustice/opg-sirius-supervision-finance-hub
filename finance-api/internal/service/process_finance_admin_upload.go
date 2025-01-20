package service

import (
	"context"
	"encoding/csv"
	"fmt"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/event"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/store"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
	"os"
	"strconv"
	"strings"
	"time"
)

func (s *Service) ProcessFinanceAdminUpload(ctx context.Context, detail shared.FinanceAdminUploadEvent) error {
	file, err := s.fileStorage.GetFile(ctx, os.Getenv("ASYNC_S3_BUCKET"), detail.Filename)
	uploadProcessedEvent := event.FinanceAdminUploadProcessed{
		EmailAddress: detail.EmailAddress,
		UploadType:   detail.UploadType,
		Filename:     detail.Filename,
	}

	if err != nil {
		uploadProcessedEvent.Error = "Unable to download report"
		return s.dispatch.FinanceAdminUploadProcessed(ctx, uploadProcessedEvent)
	}

	csvReader := csv.NewReader(file)
	records, err := csvReader.ReadAll()
	if err != nil {
		uploadProcessedEvent.Error = "Unable to read report"
		return s.dispatch.FinanceAdminUploadProcessed(ctx, uploadProcessedEvent)
	}

	failedLines, err := s.processPayments(ctx, records, detail.UploadType, detail.UploadDate)

	if err != nil {
		uploadProcessedEvent.Error = err.Error()
		return s.dispatch.FinanceAdminUploadProcessed(ctx, uploadProcessedEvent)
	}

	if len(failedLines) > 0 {
		uploadProcessedEvent.FailedLines = failedLines
		return s.dispatch.FinanceAdminUploadProcessed(ctx, uploadProcessedEvent)
	}

	return s.dispatch.FinanceAdminUploadProcessed(ctx, uploadProcessedEvent)
}

func getLedgerType(uploadType string) (string, error) {
	switch uploadType {
	case "PAYMENTS_MOTO_CARD":
		return shared.TransactionTypeMotoCardPayment.Key(), nil
	case "PAYMENTS_ONLINE_CARD":
		return shared.TransactionTypeOnlineCardPayment.Key(), nil
	case "PAYMENTS_SUPERVISION_BACS":
		return shared.TransactionTypeSupervisionBACSPayment.Key(), nil
	case "PAYMENTS_OPG_BACS":
		return shared.TransactionTypeOPGBACSPayment.Key(), nil
	}
	return "", fmt.Errorf("unknown upload type")
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

type paymentDetails struct {
	amount     int32
	date       time.Time
	courtRef   string
	ledgerType string
	uploadDate shared.Date
}

func getPaymentDetails(record []string, uploadType string, uploadDate shared.Date, ledgerType string, index int, failedLines *map[int]string) paymentDetails {
	var courtRef string
	var date time.Time
	var amount int32
	var err error

	switch uploadType {
	case "PAYMENTS_MOTO_CARD", "PAYMENTS_ONLINE_CARD":
		courtRef = strings.SplitN(record[0], "-", -1)[0]

		amount, err = parseAmount(record[2])
		if err != nil {
			(*failedLines)[index] = "AMOUNT_PARSE_ERROR"
			return paymentDetails{}
		}

		date, err = time.Parse("2006-01-02 15:04:05", record[1])
		if err != nil {
			(*failedLines)[index] = "DATE_TIME_PARSE_ERROR"
			return paymentDetails{}
		}
	case "PAYMENTS_SUPERVISION_BACS", "PAYMENTS_OPG_BACS":
		courtRef = record[10]

		amount, err = parseAmount(record[6])
		if err != nil {
			(*failedLines)[index] = "AMOUNT_PARSE_ERROR"
			return paymentDetails{}
		}

		date, err = time.Parse("02/01/2006", record[4])
		if err != nil {
			(*failedLines)[index] = "DATE_PARSE_ERROR"
			return paymentDetails{}
		}
	}

	return paymentDetails{amount, date, courtRef, ledgerType, uploadDate}
}

func (s *Service) processPayments(ctx context.Context, records [][]string, uploadType string, uploadDate shared.Date) (map[int]string, error) {
	failedLines := make(map[int]string)

	ctx, cancelTx := context.WithCancel(ctx)
	defer cancelTx()

	tx, err := s.tx.Begin(ctx)
	if err != nil {
		return nil, err
	}

	transaction := s.store.WithTx(tx)

	ledgerType, err := getLedgerType(uploadType)
	if err != nil {
		return nil, err
	}

	for index, record := range records {
		if index != 0 && record[0] != "" {
			details := getPaymentDetails(record, uploadType, uploadDate, ledgerType, index, &failedLines)

			if details != (paymentDetails{}) {
				err := s.processPaymentsUploadLine(ctx, details, index, &failedLines, transaction)
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

func (s *Service) processPaymentsUploadLine(ctx context.Context, details paymentDetails, index int, failedLines *map[int]string, transaction *store.Queries) error {
	if details.amount == 0 {
		return nil
	}

	ledgerId, _ := s.store.GetLedgerForPayment(ctx, store.GetLedgerForPaymentParams{
		CourtRef: pgtype.Text{String: details.courtRef, Valid: true},
		Amount:   details.amount,
		Type:     details.ledgerType,
		Bankdate: pgtype.Date{Time: details.date, Valid: true},
	})

	if ledgerId != 0 {
		(*failedLines)[index] = "DUPLICATE_PAYMENT"
		return nil
	}

	ledgerId, err := transaction.CreateLedgerForCourtRef(ctx, store.CreateLedgerForCourtRefParams{
		CourtRef:  pgtype.Text{String: details.courtRef, Valid: true},
		Amount:    details.amount,
		Type:      details.ledgerType,
		Status:    "CONFIRMED",
		CreatedBy: pgtype.Int4{Int32: 1, Valid: true},
		Bankdate:  pgtype.Date{Time: details.date, Valid: true},
		Datetime:  pgtype.Timestamp{Time: details.uploadDate.Time, Valid: true},
	})

	if err != nil {
		(*failedLines)[index] = "CLIENT_NOT_FOUND"
		return nil
	}

	invoices, err := transaction.GetInvoicesForCourtRef(ctx, pgtype.Text{String: details.courtRef, Valid: true})

	if err != nil {
		return err
	}

	for _, invoice := range invoices {
		allocationAmount := invoice.Outstanding
		if allocationAmount > details.amount {
			allocationAmount = details.amount
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

		details.amount -= allocationAmount
	}

	if details.amount > 0 {
		err = transaction.CreateLedgerAllocation(ctx, store.CreateLedgerAllocationParams{
			Amount:   details.amount,
			Status:   "UNAPPLIED",
			LedgerID: pgtype.Int4{Int32: ledgerId, Valid: true},
		})

		if err != nil {
			return err
		}
	}

	return nil
}
