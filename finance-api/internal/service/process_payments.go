package service

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/auth"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/store"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
	"strconv"
	"strings"
	"time"
)

func (s *Service) ProcessPayments(ctx context.Context, records [][]string, uploadType shared.ReportUploadType, bankDate shared.Date) (map[int]string, error) {
	failedLines := make(map[int]string)

	ctx, cancelTx := s.WithCancel(ctx)
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
			details := getPaymentDetails(record, uploadType, bankDate, ledgerType, index, &failedLines)

			if details != (shared.PaymentDetails{}) {
				err := s.ProcessPaymentsUploadLine(ctx, tx, details, index, &failedLines)
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

func getLedgerType(uploadType shared.ReportUploadType) (string, error) {
	switch uploadType {
	case shared.ReportTypeUploadPaymentsMOTOCard:
		return shared.TransactionTypeMotoCardPayment.Key(), nil
	case shared.ReportTypeUploadPaymentsOnlineCard:
		return shared.TransactionTypeOnlineCardPayment.Key(), nil
	case shared.ReportTypeUploadPaymentsSupervisionBACS:
		return shared.TransactionTypeSupervisionBACSPayment.Key(), nil
	case shared.ReportTypeUploadPaymentsOPGBACS:
		return shared.TransactionTypeOPGBACSPayment.Key(), nil
	case shared.ReportTypeUploadSOPUnallocated:
		return shared.TransactionTypeSOPUnallocatedPayment.Key(), nil
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

	intAmount, err := strconv.ParseInt(strings.Replace(amount, ".", "", 1), 10, 32)
	return int32(intAmount), err
}

func getPaymentDetails(record []string, uploadType shared.ReportUploadType, bankDate shared.Date, ledgerType string, index int, failedLines *map[int]string) shared.PaymentDetails {
	var courtRef string
	var receivedDate time.Time
	var amount int32
	var err error

	switch uploadType {
	case shared.ReportTypeUploadPaymentsMOTOCard, shared.ReportTypeUploadPaymentsOnlineCard:
		courtRef = strings.SplitN(record[0], "-", -1)[0]

		amount, err = parseAmount(record[2])
		if err != nil {
			(*failedLines)[index] = "AMOUNT_PARSE_ERROR"
			return shared.PaymentDetails{}
		}

		receivedDate, err = time.Parse("2006-01-02 15:04:05", record[1])
		if err != nil {
			(*failedLines)[index] = "DATE_TIME_PARSE_ERROR"
			return shared.PaymentDetails{}
		}
	case shared.ReportTypeUploadPaymentsSupervisionBACS, shared.ReportTypeUploadPaymentsOPGBACS:
		courtRef = record[10]

		amount, err = parseAmount(record[6])
		if err != nil {
			(*failedLines)[index] = "AMOUNT_PARSE_ERROR"
			return shared.PaymentDetails{}
		}

		receivedDate, err = time.Parse("02/01/2006", record[4])
		if err != nil {
			(*failedLines)[index] = "DATE_PARSE_ERROR"
			return shared.PaymentDetails{}
		}
	case shared.ReportTypeUploadSOPUnallocated:
		courtRef = record[0]

		amount, err = parseAmount(record[1])
		if err != nil {
			(*failedLines)[index] = "AMOUNT_PARSE_ERROR"
			return shared.PaymentDetails{}
		}

		receivedDate, err = time.Parse("2006-01-02 15:04:05", "2025-03-31 00:00:00")
		if err != nil {
			(*failedLines)[index] = "DATE_TIME_PARSE_ERROR"
			return shared.PaymentDetails{}
		}
	}

	return shared.PaymentDetails{Amount: amount, BankDate: bankDate.Time, CourtRef: courtRef, LedgerType: ledgerType, ReceivedDate: receivedDate}
}

func (s *Service) ProcessPaymentsUploadLine(ctx context.Context, tx *store.Tx, details shared.PaymentDetails, index int, failedLines *map[int]string) error {
	if details.Amount == 0 {
		return nil
	}

	var (
		courtRef   pgtype.Text
		bankDate   pgtype.Date
		createdBy  pgtype.Int4
		uploadDate pgtype.Timestamp
	)

	_ = courtRef.Scan(details.CourtRef)
	_ = bankDate.Scan(details.BankDate)
	_ = uploadDate.Scan(details.ReceivedDate)
	_ = store.ToInt4(&createdBy, ctx.(auth.Context).User.ID)

	exists, _ := s.store.CheckDuplicateLedger(ctx, store.CheckDuplicateLedgerParams{
		CourtRef:     courtRef,
		Amount:       details.Amount,
		Type:         details.LedgerType,
		BankDate:     bankDate,
		ReceivedDate: uploadDate,
	})

	if exists {
		(*failedLines)[index] = "DUPLICATE_PAYMENT"
		return nil
	}

	ledgerID, err := tx.CreateLedgerForCourtRef(ctx, store.CreateLedgerForCourtRefParams{
		CourtRef:  courtRef,
		Amount:    details.Amount,
		Type:      details.LedgerType,
		Status:    "CONFIRMED",
		CreatedBy: createdBy,
		Bankdate:  bankDate,
		Datetime:  uploadDate,
	})

	if err != nil {
		(*failedLines)[index] = "CLIENT_NOT_FOUND"
		return nil
	}

	invoices, err := tx.GetInvoicesForCourtRef(ctx, courtRef)

	if err != nil {
		return err
	}

	remaining := details.Amount

	for _, invoice := range invoices {
		allocationAmount := invoice.Outstanding
		if allocationAmount > remaining {
			allocationAmount = remaining
		}

		var invoiceID pgtype.Int4
		_ = store.ToInt4(&invoiceID, invoice.ID)

		err = tx.CreateLedgerAllocation(ctx, store.CreateLedgerAllocationParams{
			InvoiceID: invoiceID,
			Amount:    allocationAmount,
			Status:    "ALLOCATED",
			LedgerID:  ledgerID,
		})
		if err != nil {
			return err
		}

		remaining -= allocationAmount
		if remaining <= 0 {
			break
		}
	}

	if remaining > 0 {
		err = tx.CreateLedgerAllocation(ctx, store.CreateLedgerAllocationParams{
			Amount:   -remaining,
			Status:   "UNAPPLIED",
			LedgerID: ledgerID,
		})

		if err != nil {
			return err
		}
	}

	return nil
}
