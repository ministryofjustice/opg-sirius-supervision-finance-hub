package service

import (
	"context"
	"encoding/csv"
	"fmt"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/notify"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/store"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
	"os"
	"slices"
	"strconv"
	"strings"
	"time"
)

func (s *Service) ProcessFinanceAdminUpload(ctx context.Context, detail shared.FinanceAdminUploadEvent) error {
	file, err := s.fileStorage.GetFile(ctx, os.Getenv("ASYNC_S3_BUCKET"), detail.Filename)

	if err != nil {
		payload := createUploadNotifyPayload(detail, fmt.Errorf("unable to download report"), map[int]string{})
		return s.notify.Send(ctx, payload)
	}

	csvReader := csv.NewReader(file)
	records, err := csvReader.ReadAll()
	if err != nil {
		payload := createUploadNotifyPayload(detail, fmt.Errorf("unable to read report"), map[int]string{})
		return s.notify.Send(ctx, payload)
	}

	failedLines, err := s.processPayments(ctx, records, detail.UploadType, detail.UploadDate)
	payload := createUploadNotifyPayload(detail, err, failedLines)

	return s.notify.Send(ctx, payload)
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

	intAmount, err := strconv.ParseInt(strings.Replace(amount, ".", "", 1), 10, 32)
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
			(*failedLines)[index] = "DATE_TIME_PARSE_ERROR"
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
	case "SOP_UNALLOCATED":
		courtRef = record[0]

		amount, err = parseAmount(record[1])
		if err != nil {
			(*failedLines)[index] = "AMOUNT_PARSE_ERROR"
			return shared.PaymentDetails{}
		}

		bankDate, err = time.Parse("2006-01-02 15:04:05", "2025-03-31 00:00:00")
		if err != nil {
			(*failedLines)[index] = "DATE_TIME_PARSE_ERROR"
			return shared.PaymentDetails{}
		}
	}

	return shared.PaymentDetails{Amount: amount, BankDate: bankDate, CourtRef: courtRef, LedgerType: ledgerType, UploadDate: uploadDate.Time}
}

func (s *Service) processPayments(ctx context.Context, records [][]string, uploadType string, uploadDate shared.Date) (map[int]string, error) {
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
			details := getPaymentDetails(record, uploadType, uploadDate, ledgerType, index, &failedLines)

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
	_ = store.ToInt4(&createdBy, s.env.SystemUserID)
	_ = uploadDate.Scan(details.UploadDate)

	ledgerId, _ := s.store.GetLedgerForPayment(ctx, store.GetLedgerForPaymentParams{
		CourtRef: courtRef,
		Amount:   details.Amount,
		Type:     details.LedgerType,
		Bankdate: bankDate,
	})

	if ledgerId != 0 {
		(*failedLines)[index] = "DUPLICATE_PAYMENT"
		return nil
	}

	ledgerId, err := tx.CreateLedgerForCourtRef(ctx, store.CreateLedgerForCourtRefParams{
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

	var ledgerID pgtype.Int4
	_ = store.ToInt4(&ledgerID, ledgerId)

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

func createUploadNotifyPayload(detail shared.FinanceAdminUploadEvent, err error, failedLines map[int]string) notify.Payload {
	var payload notify.Payload

	uploadType := shared.ParseReportUploadType(detail.UploadType)
	if err != nil {
		payload = notify.Payload{
			EmailAddress: detail.EmailAddress,
			TemplateId:   notify.ProcessingErrorTemplateId,
			Personalisation: struct {
				Error      string `json:"error"`
				UploadType string `json:"upload_type"`
			}{
				Error:      err.Error(),
				UploadType: uploadType.Translation(),
			},
		}
	} else if len(failedLines) != 0 {
		payload = notify.Payload{
			EmailAddress: detail.EmailAddress,
			TemplateId:   notify.ProcessingFailedTemplateId,
			Personalisation: struct {
				FailedLines []string `json:"failed_lines"`
				UploadType  string   `json:"upload_type"`
			}{
				FailedLines: formatFailedLines(failedLines),
				UploadType:  uploadType.Translation(),
			},
		}
	} else {
		payload = notify.Payload{
			EmailAddress: detail.EmailAddress,
			TemplateId:   notify.ProcessingSuccessTemplateId,
			Personalisation: struct {
				UploadType string `json:"upload_type"`
			}{uploadType.Translation()},
		}
	}

	return payload
}

func formatFailedLines(failedLines map[int]string) []string {
	var errorMessage string
	var formattedLines []string
	var keys []int
	for i := range failedLines {
		keys = append(keys, i)
	}

	slices.Sort(keys)

	for _, key := range keys {
		failedLine := failedLines[key]
		errorMessage = ""

		switch failedLine {
		case "DATE_PARSE_ERROR":
			errorMessage = "Unable to parse date - please use the format DD/MM/YYYY"
		case "DATE_TIME_PARSE_ERROR":
			errorMessage = "Unable to parse date - please use the format YYYY-MM-DD HH:MM:SS"
		case "AMOUNT_PARSE_ERROR":
			errorMessage = "Unable to parse amount - please use the format 320.00"
		case "DUPLICATE_PAYMENT":
			errorMessage = "Duplicate payment line"
		case "CLIENT_NOT_FOUND":
			errorMessage = "Could not find a client with this court reference"
		}

		formattedLines = append(formattedLines, fmt.Sprintf("Line %d: %s", key, errorMessage))
	}

	return formattedLines
}
