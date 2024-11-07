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

func (s *Service) ProcessFinanceAdminUpload(ctx context.Context, filename string, email string, uploadType string) error {
	file, err := s.filestorage.GetFile(ctx, os.Getenv("ASYNC_S3_BUCKET"), filename)

	if err != nil {
		failedEvent := event.FinanceAdminUploadProcessed{EmailAddress: email, UploadType: uploadType, Error: "Unable to download report"}
		return s.dispatch.FinanceAdminUploadProcessed(ctx, failedEvent)
	}

	csvReader := csv.NewReader(file)
	records, err := csvReader.ReadAll()
	if err != nil {
		failedEvent := event.FinanceAdminUploadProcessed{EmailAddress: email, UploadType: uploadType, Error: "Unable to read report"}
		return s.dispatch.FinanceAdminUploadProcessed(ctx, failedEvent)
	}

	failedLines, err := s.processPayments(ctx, records, uploadType)

	if err != nil {
		failedEvent := event.FinanceAdminUploadProcessed{EmailAddress: email, UploadType: uploadType, Error: err.Error()}
		return s.dispatch.FinanceAdminUploadProcessed(ctx, failedEvent)
	}

	if len(failedLines) > 0 {
		failedEvent := event.FinanceAdminUploadProcessed{EmailAddress: email, UploadType: uploadType, FailedLines: failedLines}
		return s.dispatch.FinanceAdminUploadProcessed(ctx, failedEvent)
	}

	return s.dispatch.FinanceAdminUploadProcessed(ctx, event.FinanceAdminUploadProcessed{EmailAddress: email, UploadType: uploadType})
}

func getLedgerType(uploadType string) (string, error) {
	switch uploadType {
	case "PAYMENTS_MOTO_CARD":
		return "MOTO card payment", nil
	case "PAYMENTS_ONLINE_CARD":
		return "Online card payment", nil
	case "PAYMENTS_SUPERVISION_BACS":
		return "BACS payment (Supervision account)", nil
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
}

func getPaymentDetails(record []string, uploadType string, ledgerType string, index int, failedLines *map[int]string) paymentDetails {
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
			(*failedLines)[index] = "DATE_PARSE_ERROR"
			return paymentDetails{}
		}
	case "PAYMENTS_SUPERVISION_BACS":
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

	return paymentDetails{amount, date, courtRef, ledgerType}
}

func (s *Service) processPayments(ctx context.Context, records [][]string, uploadType string) (map[int]string, error) {
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
			details := getPaymentDetails(record, uploadType, ledgerType, index, &failedLines)

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
		Datetime: pgtype.Timestamp{Time: details.date, Valid: true},
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
		Datetime:  pgtype.Timestamp{Time: details.date, Valid: true},
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
