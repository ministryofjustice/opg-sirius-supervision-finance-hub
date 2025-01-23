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
		payload := createUploadNotifyPayload(detail, "Unable to download report", map[int]string{})
		err := s.notify.Send(ctx, payload)
		if err != nil {
			return err
		}
		return nil
	}

	csvReader := csv.NewReader(file)
	records, err := csvReader.ReadAll()
	if err != nil {
		payload := createUploadNotifyPayload(detail, "Unable to read report", map[int]string{})
		err := s.notify.Send(ctx, payload)
		if err != nil {
			return err
		}
		return nil
	}

	failedLines, err := s.processPayments(ctx, records, detail.UploadType, detail.UploadDate)

	if err != nil {
		payload := createUploadNotifyPayload(detail, err.Error(), map[int]string{})
		err := s.notify.Send(ctx, payload)
		if err != nil {
			return err
		}
		return nil
	}

	if len(failedLines) > 0 {
		payload := createUploadNotifyPayload(detail, "", failedLines)
		err := s.notify.Send(ctx, payload)
		if err != nil {
			return err
		}
		return nil
	}

	payload := createUploadNotifyPayload(detail, "", map[int]string{})
	err = s.notify.Send(ctx, payload)
	if err != nil {
		return err
	}
	return nil
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
	}

	return shared.PaymentDetails{Amount: amount, BankDate: bankDate, CourtRef: courtRef, LedgerType: ledgerType, UploadDate: uploadDate.Time}
}

func (s *Service) processPayments(ctx context.Context, records [][]string, uploadType string, uploadDate shared.Date) (map[int]string, error) {
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
	})

	if err != nil {
		(*failedLines)[index] = "CLIENT_NOT_FOUND"
		return nil
	}

	invoices, err := tx.GetInvoicesForCourtRef(ctx, pgtype.Text{String: details.CourtRef, Valid: true})

	if err != nil {
		return err
	}

	for _, invoice := range invoices {
		allocationAmount := invoice.Outstanding
		if allocationAmount > details.Amount {
			allocationAmount = details.Amount
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

		details.Amount -= allocationAmount
	}

	if details.Amount > 0 {
		err = tx.CreateLedgerAllocation(ctx, store.CreateLedgerAllocationParams{
			Amount:   details.Amount,
			Status:   "UNAPPLIED",
			LedgerID: pgtype.Int4{Int32: ledgerId, Valid: true},
		})

		if err != nil {
			return err
		}
	}

	return nil
}

func createUploadNotifyPayload(detail shared.FinanceAdminUploadEvent, error string, failedLines map[int]string) notify.Payload {
	var payload notify.Payload

	uploadType := shared.ParseReportUploadType(detail.UploadType)
	if error != "" {
		payload = notify.Payload{
			EmailAddress: detail.EmailAddress,
			TemplateId:   notify.ProcessingErrorTemplateId,
			Personalisation: struct {
				Error      string `json:"error"`
				UploadType string `json:"upload_type"`
			}{
				Error:      error,
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
