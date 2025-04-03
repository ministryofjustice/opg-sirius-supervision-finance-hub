package service

import (
	"context"
	"encoding/csv"
	"fmt"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/auth"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/notify"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/store"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
	"log/slog"
	"os"
	"slices"
	"strconv"
	"strings"
	"time"
)

func (s *Service) ProcessFinanceAdminUpload(ctx context.Context, detail shared.FinanceAdminUploadEvent) error {
	var pisNumber shared.Nillable[int]

	file, err := s.fileStorage.GetFile(ctx, os.Getenv("ASYNC_S3_BUCKET"), detail.Filename)
	s.Logger(ctx).Info("processing file " + detail.Filename)

	if err != nil {
		s.Logger(ctx).Error("Error unable to download report", slog.String("err", err.Error()))
		payload := createUploadNotifyPayload(detail, fmt.Errorf("unable to download report"), map[int]string{})
		return s.notify.Send(ctx, payload)
	}

	csvReader := csv.NewReader(file)
	records, err := csvReader.ReadAll()
	if err != nil {
		s.Logger(ctx).Error("Error unable to read report", slog.String("err", err.Error()))
		payload := createUploadNotifyPayload(detail, fmt.Errorf("unable to read report"), map[int]string{})
		return s.notify.Send(ctx, payload)
	}

	if detail.PisNumber != 0 {
		pisNumber = shared.Nillable[int]{Value: detail.PisNumber, Valid: true}
	}

	failedLines, err := s.processPayments(ctx, records, detail.UploadType, detail.UploadDate, pisNumber)
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
	case "PAYMENTS_SUPERVISION_CHEQUE":
		return shared.TransactionTypeSupervisionChequePayment.Key(), nil
	case "SOP_UNALLOCATED":
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

func getPaymentDetails(record []string, uploadType string, bankDate shared.Date, ledgerType string, index int, failedLines *map[int]string, pisNumber shared.Nillable[int]) shared.PaymentDetails {
	var courtRef string
	var amountString string
	var receivedDateString string

	switch uploadType {
	case shared.ReportTypeUploadPaymentsMOTOCard.Key(), shared.ReportTypeUploadPaymentsOnlineCard.Key():
		courtRef = strings.SplitN(record[0], "-", -1)[0]
		amountString = record[2]
		receivedDateString = record[1]
	case shared.ReportTypeUploadPaymentsSupervisionBACS.Key(), shared.ReportTypeUploadPaymentsOPGBACS.Key():
		courtRef = record[10]
		amountString = record[6]
		receivedDateString = record[4]
  case "PAYMENTS_SUPERVISION_CHEQUE":
		courtRef = record[0]
		amountString = record[2]
		receivedDateString = record[4]
	case "SOP_UNALLOCATED":
		courtRef = record[0]
		amountString = record[1]
		receivedDateString = "31/03/2025"
	}

	amount, err := parseAmount(amountString)
	if err != nil {
		(*failedLines)[index] = "AMOUNT_PARSE_ERROR"
		return shared.PaymentDetails{}
	}

	receivedDate, err := time.Parse("02/01/2006", receivedDateString)
	if err != nil {
		(*failedLines)[index] = "DATE_PARSE_ERROR"
		return shared.PaymentDetails{}
	}

	return shared.PaymentDetails{Amount: amount, BankDate: bankDate.Time, CourtRef: courtRef, LedgerType: ledgerType, ReceivedDate: receivedDate, PisNumber: pisNumber}
}

func (s *Service) processPayments(ctx context.Context, records [][]string, uploadType string, bankDate shared.Date, pisNumber shared.Nillable[int]) (map[int]string, error) {
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
			details := getPaymentDetails(record, uploadType, bankDate, ledgerType, index, &failedLines, pisNumber)

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
		pisNumber  pgtype.Int4
		uploadDate pgtype.Timestamp
	)

	_ = courtRef.Scan(details.CourtRef)
	_ = bankDate.Scan(details.BankDate)
	_ = uploadDate.Scan(details.ReceivedDate)
	_ = store.ToInt4(&createdBy, ctx.(auth.Context).User.ID)

	if details.PisNumber.Valid {
		_ = store.ToInt4(&pisNumber, details.PisNumber.Value)
	}

	ledgerID, _ := s.store.GetLedgerForPayment(ctx, store.GetLedgerForPaymentParams{
		CourtRef: courtRef,
		Amount:   details.Amount,
		Type:     details.LedgerType,
		Bankdate: bankDate,
	})

	if ledgerID != 0 {
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
		PisNumber: pisNumber,
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
