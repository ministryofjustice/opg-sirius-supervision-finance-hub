package service

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/auth"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/event"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/store"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/validation"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
)

func (s *Service) ProcessPayments(ctx context.Context, records [][]string, uploadType shared.ReportUploadType, bankDate shared.Date, pisNumber int) (map[int]string, error) {
	failedLines := make(map[int]string)

	tx, err := s.BeginStoreTx(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	for index, record := range records {
		if !isHeaderRow(uploadType, index) && safeRead(record, 0) != "" {
			details := getPaymentDetails(ctx, record, uploadType, bankDate, pisNumber, index, &failedLines)

			if details != (shared.PaymentDetails{}) {
				if !s.validatePaymentLine(ctx, details, index, &failedLines) {
					continue
				}

				_, err := s.ProcessPaymentsUploadLine(ctx, tx, details)
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

func getLedgerType(uploadType shared.ReportUploadType) (shared.TransactionType, error) {
	switch uploadType {
	case shared.ReportTypeUploadPaymentsMOTOCard:
		return shared.TransactionTypeMotoCardPayment, nil
	case shared.ReportTypeUploadPaymentsOnlineCard:
		return shared.TransactionTypeOnlineCardPayment, nil
	case shared.ReportTypeUploadPaymentsSupervisionBACS:
		return shared.TransactionTypeSupervisionBACSPayment, nil
	case shared.ReportTypeUploadPaymentsOPGBACS:
		return shared.TransactionTypeOPGBACSPayment, nil
	case shared.ReportTypeUploadPaymentsSupervisionCheque:
		return shared.TransactionTypeSupervisionChequePayment, nil
	case shared.ReportTypeUploadDirectDebitsCollections:
		return shared.TransactionTypeDirectDebitPayment, nil
	}
	return shared.TransactionTypeUnknown, fmt.Errorf("unknown upload type")
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

func getPaymentDetails(ctx context.Context, record []string, uploadType shared.ReportUploadType, formDate shared.Date, pisNumber int, index int, failedLines *map[int]string) shared.PaymentDetails {
	var (
		paymentType  shared.TransactionType
		courtRef     pgtype.Text
		bankDate     pgtype.Date
		receivedDate pgtype.Timestamp
		createdBy    pgtype.Int4
		pis          pgtype.Int4
		amount       int32
		err          error
	)

	_ = bankDate.Scan(formDate.Time)
	_ = store.ToInt4(&createdBy, ctx.(auth.Context).User.ID)
	_ = store.ToInt4(&pis, pisNumber)

	paymentType, err = getLedgerType(uploadType)
	if err != nil {
		(*failedLines)[index] = validation.UploadErrorDateParse
		return shared.PaymentDetails{}
	}

	switch uploadType {
	case shared.ReportTypeUploadPaymentsMOTOCard, shared.ReportTypeUploadPaymentsOnlineCard:
		_ = courtRef.Scan(strings.Split(safeRead(record, 0), "-")[0])

		amount, err = parseAmount(safeRead(record, 2))
		if err != nil {
			(*failedLines)[index] = validation.UploadErrorAmountParse
			return shared.PaymentDetails{}
		}

		rd, err := time.Parse("02/01/2006", safeRead(record, 1))
		if err != nil {
			(*failedLines)[index] = validation.UploadErrorDateParse
			return shared.PaymentDetails{}
		}
		_ = receivedDate.Scan(rd)
	case shared.ReportTypeUploadPaymentsSupervisionBACS, shared.ReportTypeUploadPaymentsOPGBACS:
		_ = courtRef.Scan(safeRead(record, 10))

		amount, err = parseAmount(safeRead(record, 6))
		if err != nil {
			(*failedLines)[index] = validation.UploadErrorAmountParse
			return shared.PaymentDetails{}
		}

		rd, err := time.Parse("02/01/2006", safeRead(record, 4))
		if err != nil {
			(*failedLines)[index] = validation.UploadErrorDateParse
			return shared.PaymentDetails{}
		}
		_ = receivedDate.Scan(rd)
	case shared.ReportTypeUploadPaymentsSupervisionCheque:
		_ = courtRef.Scan(safeRead(record, 0))

		amount, err = parseAmount(safeRead(record, 2))
		if err != nil {
			(*failedLines)[index] = validation.UploadErrorAmountParse
			return shared.PaymentDetails{}
		}

		rd, err := time.Parse("02/01/2006", safeRead(record, 4))
		if err != nil {
			(*failedLines)[index] = validation.UploadErrorDateParse
			return shared.PaymentDetails{}
		}
		_ = receivedDate.Scan(rd)
	case shared.ReportTypeUploadDirectDebitsCollections:
		_ = courtRef.Scan(strings.TrimSpace(safeRead(record, 1)))
		amount, err = parseAmount(strings.TrimSpace(safeRead(record, 2)))
		if err != nil {
			(*failedLines)[index] = validation.UploadErrorAmountParse
			return shared.PaymentDetails{}
		}
		rd, err := time.Parse("02/01/2006", strings.TrimSpace(safeRead(record, 4)))
		if err != nil {
			(*failedLines)[index] = validation.UploadErrorDateParse
			return shared.PaymentDetails{}
		}
		_ = receivedDate.Scan(rd)
	}

	return shared.PaymentDetails{Amount: amount, BankDate: bankDate, CourtRef: courtRef, LedgerType: paymentType, ReceivedDate: receivedDate, CreatedBy: createdBy, PisNumber: pis}
}

/*
Payment lines are invalid if either the payment cannot be matched to a client, or if the payment has already been processed
(to prevent duplicates being created if the file is uploaded more than once).
Duplicate payments within the same file will be processed due to the transaction only being committed once all lines have
been processed, which is expected behaviour as there are legitimate reasons for a payment being duplicated, but duplicates will always appear in the same file.
*/
func (s *Service) validatePaymentLine(ctx context.Context, details shared.PaymentDetails, index int, failedLines *map[int]string) bool {
	count, _ := s.store.CountDuplicateLedger(ctx, store.CountDuplicateLedgerParams{
		CourtRef:     details.CourtRef,
		Amount:       details.Amount,
		Type:         details.LedgerType.Key(),
		BankDate:     details.BankDate,
		ReceivedDate: details.ReceivedDate,
		PisNumber:    details.PisNumber,
	})

	if count > 0 {
		(*failedLines)[index] = validation.UploadErrorDuplicatePayment
		return false
	}

	exists, _ := s.store.CheckClientExistsByCourtRef(ctx, details.CourtRef)

	if !exists {
		(*failedLines)[index] = validation.UploadErrorClientNotFound
		return false
	}

	return true
}

func (s *Service) ProcessPaymentsUploadLine(ctx context.Context, tx *store.Tx, details shared.PaymentDetails) (int32, error) {
	if details.Amount <= 0 {
		return 0, nil
	}

	invoices, err := tx.GetUnpaidInvoicesByCourtRef(ctx, details.CourtRef)

	if err != nil {
		return 0, err
	}

	remaining := details.Amount
	var allocations []store.CreateLedgerAllocationParams

	for _, invoice := range invoices {
		allocationAmount := invoice.Outstanding
		if allocationAmount > remaining {
			allocationAmount = remaining
		}

		var invoiceID pgtype.Int4
		_ = store.ToInt4(&invoiceID, invoice.ID)

		allocations = append(allocations, store.CreateLedgerAllocationParams{
			InvoiceID: invoiceID,
			Amount:    allocationAmount,
			Status:    "ALLOCATED",
		})

		remaining -= allocationAmount
		if remaining <= 0 {
			break
		}
	}

	if remaining > 0 {
		allocations = append(allocations, store.CreateLedgerAllocationParams{
			Amount: -remaining,
			Status: "UNAPPLIED",
		})
	}

	if len(allocations) == 0 {
		// this should never occur but back out and log if it does
		s.Logger(ctx).Error("no allocations generated for ledger in payment line", "courtref", details.CourtRef, "amount", details.Amount)
		return 0, nil
	}

	params := store.CreateLedgerForCourtRefParams{
		CourtRef:     details.CourtRef,
		Amount:       details.Amount,
		Type:         details.LedgerType.Key(),
		Status:       "CONFIRMED",
		CreatedBy:    details.CreatedBy,
		BankDate:     details.BankDate,
		ReceivedDate: details.ReceivedDate,
		PisNumber:    details.PisNumber,
	}
	ledgerID, err := tx.CreateLedgerForCourtRef(ctx, params)

	if err != nil {
		return 0, err
	}

	for _, allocation := range allocations {
		allocation.LedgerID = ledgerID
		err = tx.CreateLedgerAllocation(ctx, allocation)
		if err != nil {
			return 0, err
		}
	}

	if remaining > 0 {
		client, _ := tx.GetClientByCourtRef(ctx, details.CourtRef)
		err = s.dispatch.CreditOnAccount(ctx, event.CreditOnAccount{
			ClientID:        int(client.ClientID),
			CreditRemaining: int(remaining),
		})
		return ledgerID, err
	}

	return ledgerID, nil
}

func isHeaderRow(uploadType shared.ReportUploadType, index int) bool {
	return uploadType.HasHeader() && index == 0
}

func safeRead(record []string, index int) string {
	if index >= len(record) || index < 0 {
		return ""
	}
	return record[index]
}
