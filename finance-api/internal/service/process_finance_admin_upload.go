package service

import (
	"context"
	"encoding/csv"
	"fmt"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/event"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/store"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
	"strconv"
	"strings"
	"time"
)

func (s *Service) ProcessFinanceAdminUpload(ctx context.Context, detail shared.FinanceAdminUploadEvent) error {
	file, err := s.fileStorage.GetFile(ctx, s.env.AsyncBucket, detail.Filename)
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

	failedLines, err := s.processPayments(ctx, records, detail.UploadType, detail.UploadDate, detail.PisNumber)

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

func getPaymentDetails(record []string, uploadType string, uploadDate shared.Date, ledgerType string, index int, failedLines *map[int]string) shared.PaymentDetails {
	var courtRef string
	var bankDate time.Time
	var amount int32
	var err error

	switch uploadType {
	case "PAYMENTS_MOTO_CARD", "PAYMENTS_ONLINE_CARD":
		courtRef = strings.SplitN(record[0], "-", -1)[0]

		amount, err = parseAmount(record[2])
		if err != nil {
			(*failedLines)[index] = "AMOUNT_PARSE_ERROR"
			return shared.PaymentDetails{}
		}

		bankDate, err = time.Parse("2006-01-02 15:04:05", record[1])
		if err != nil {
			(*failedLines)[index] = "DATE_PARSE_ERROR"
			return shared.PaymentDetails{}
		}
	case "PAYMENTS_SUPERVISION_BACS", "PAYMENTS_OPG_BACS":
		courtRef = record[10]

		amount, err = parseAmount(record[6])
		if err != nil {
			(*failedLines)[index] = "AMOUNT_PARSE_ERROR"
			return shared.PaymentDetails{}
		}

		bankDate, err = time.Parse("02/01/2006", record[4])
		if err != nil {
			(*failedLines)[index] = "DATE_PARSE_ERROR"
			return shared.PaymentDetails{}
		}
	case "PAYMENTS_SUPERVISION_CHEQUE":
		courtRef = record[0]

		amount, err = parseAmount(record[2])
		if err != nil {
			(*failedLines)[index] = "AMOUNT_PARSE_ERROR"
			return shared.PaymentDetails{}
		}

		bankDate, err = time.Parse("02/01/2006", record[4])
		if err != nil {
			(*failedLines)[index] = "DATE_PARSE_ERROR"
			return shared.PaymentDetails{}
		}
	}

	return shared.PaymentDetails{Amount: amount, BankDate: bankDate, CourtRef: courtRef, LedgerType: ledgerType, UploadDate: uploadDate.Time}
}

func (s *Service) processPayments(ctx context.Context, records [][]string, uploadType string, uploadDate shared.Date, pisNumber int) (map[int]string, error) {
	failedLines := make(map[int]string)

	ctx, cancelTx := context.WithCancel(ctx)
	defer cancelTx()

	tx, err := s.BeginStoreTx(ctx)
	if err != nil {
		return nil, err
	}

	ledgerType, err := getLedgerType(uploadType)
	if err != nil {
		return nil, err
	}

	for index, record := range records {
		if index != 0 && record[0] != "" {
			details := getPaymentDetails(record, uploadType, uploadDate, ledgerType, index, &failedLines)

			if details != (shared.PaymentDetails{}) {
				err := s.ProcessPaymentsUploadLine(ctx, tx, details, index, &failedLines, pisNumber)
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

func (s *Service) ProcessPaymentsUploadLine(ctx context.Context, tx *store.Tx, details shared.PaymentDetails, index int, failedLines *map[int]string, pisNumber int) error {
	if details.Amount == 0 {
		return nil
	}

	ledgerId, _ := s.store.GetLedgerForPayment(ctx, store.GetLedgerForPaymentParams{
		CourtRef: pgtype.Text{String: details.CourtRef, Valid: true},
		Amount:   details.Amount,
		Type:     details.LedgerType,
		Bankdate: pgtype.Date{Time: details.BankDate, Valid: true},
	})

	if ledgerId != 0 {
		(*failedLines)[index] = "DUPLICATE_PAYMENT"
		return nil
	}

	ledgerId, err := tx.CreateLedgerForCourtRef(ctx, store.CreateLedgerForCourtRefParams{
		CourtRef:  pgtype.Text{String: details.CourtRef, Valid: true},
		Amount:    details.Amount,
		Type:      details.LedgerType,
		Status:    "CONFIRMED",
		CreatedBy: pgtype.Int4{Int32: 1, Valid: true},
		Bankdate:  pgtype.Date{Time: details.BankDate, Valid: true},
		Datetime:  pgtype.Timestamp{Time: details.UploadDate, Valid: true},
		PisNumber: pgtype.Int4{Int32: int32(pisNumber), Valid: true},
	})

	if err != nil {
		(*failedLines)[index] = "CLIENT_NOT_FOUND"
		return nil
	}

	invoices, err := tx.GetInvoicesForCourtRef(ctx, pgtype.Text{String: details.CourtRef, Valid: true})

	if err != nil {
		return err
	}

	remaining := details.Amount

	for _, invoice := range invoices {
		allocationAmount := invoice.Outstanding
		if allocationAmount > remaining {
			allocationAmount = remaining
		}

		err = tx.CreateLedgerAllocation(ctx, store.CreateLedgerAllocationParams{
			InvoiceID: pgtype.Int4{Int32: invoice.ID, Valid: true},
			Amount:    allocationAmount,
			Status:    "ALLOCATED",
			LedgerID:  pgtype.Int4{Int32: ledgerId, Valid: true},
		})
		if err != nil {
			return err
		}

		remaining -= allocationAmount
	}

	if remaining > 0 {
		err = tx.CreateLedgerAllocation(ctx, store.CreateLedgerAllocationParams{
			Amount:   -remaining,
			Status:   "UNAPPLIED",
			LedgerID: pgtype.Int4{Int32: ledgerId, Valid: true},
		})

		if err != nil {
			return err
		}
	}

	return nil
}
