package service

import (
	"context"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/auth"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/store"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
	"time"
)

type reversalDetail struct {
	paymentType     shared.TransactionType
	erroredCourtRef pgtype.Text
	correctCourtRef pgtype.Text
	bankDate        pgtype.Date
	receivedDate    pgtype.Timestamp
	amount          int32
	createdBy       pgtype.Int4
}

func (s *Service) ProcessPaymentReversals(ctx context.Context, records [][]string, uploadType shared.ReportUploadType) (map[int]string, error) {
	// extract payments from the file

	failedLines := make(map[int]string)

	ctx, cancelTx := s.WithCancel(ctx)
	defer cancelTx()

	tx, err := s.BeginStoreTx(ctx)
	if err != nil {
		return nil, err
	}

	for index, record := range records {
		if index != 0 && record[0] != "" {
			details := getReversalLines(ctx, record, uploadType, index, &failedLines)

			if details != (reversalDetail{}) {
				if !s.validateReversalLine(ctx, details, index, &failedLines) {
					continue
				}

				err = s.processReversalUploadLine(ctx, tx, details)
				if err != nil {
					return nil, err
				}

				if uploadType == shared.ReportTypeUploadMisappliedPayments {
					err = s.ProcessPaymentsUploadLine(ctx, tx, shared.PaymentDetails{
						Amount:       details.amount,
						BankDate:     details.bankDate,
						CourtRef:     details.correctCourtRef,
						LedgerType:   details.paymentType,
						ReceivedDate: details.receivedDate,
						CreatedBy:    details.createdBy,
					})
					if err != nil {
						return nil, err
					}
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

func getReversalLines(ctx context.Context, record []string, uploadType shared.ReportUploadType, index int, failedLines *map[int]string) reversalDetail {
	var (
		paymentType     shared.TransactionType
		erroredCourtRef pgtype.Text
		correctCourtRef pgtype.Text
		bankDate        pgtype.Date
		receivedDate    pgtype.Timestamp
		createdBy       pgtype.Int4
		amount          int32
	)

	switch uploadType {
	case shared.ReportTypeUploadMisappliedPayments:
		paymentType = shared.ParseTransactionType(record[0])
		if paymentType == shared.TransactionTypeUnknown {
			(*failedLines)[index] = "PAYMENT_TYPE_PARSE_ERROR"
			return reversalDetail{}
		}

		_ = erroredCourtRef.Scan(record[1])
		_ = correctCourtRef.Scan(record[2])

		bd, err := time.Parse("2006-01-02", record[3])
		if err != nil {
			(*failedLines)[index] = "BANK_DATE_PARSE_ERROR"
			return reversalDetail{}
		}
		_ = bankDate.Scan(bd)

		rd, err := time.Parse("2006-01-02", record[4])
		if err != nil {
			(*failedLines)[index] = "RECEIVED_DATE_PARSE_ERROR"
			return reversalDetail{}
		}
		_ = receivedDate.Scan(rd)

		amount, err = parseAmount(record[5])
		if err != nil {
			(*failedLines)[index] = "AMOUNT_PARSE_ERROR"
			return reversalDetail{}
		}

	default:
		(*failedLines)[index] = "UNKNOWN_UPLOAD_TYPE"
	}

	_ = store.ToInt4(&createdBy, ctx.(auth.Context).User.ID)

	return reversalDetail{
		paymentType:     paymentType,
		erroredCourtRef: erroredCourtRef,
		correctCourtRef: correctCourtRef,
		bankDate:        bankDate,
		receivedDate:    receivedDate,
		amount:          amount,
	}
}

func (s *Service) validateReversalLine(ctx context.Context, details reversalDetail, index int, failedLines *map[int]string) bool {
	params := store.CheckDuplicateLedgerParams{
		CourtRef:     details.erroredCourtRef,
		Amount:       details.amount,
		Type:         details.paymentType.Key(),
		BankDate:     details.bankDate,
		ReceivedDate: details.receivedDate,
	}
	exists, _ := s.store.CheckDuplicateLedger(ctx, params)

	if !exists {
		(*failedLines)[index] = "NO_MATCHED_PAYMENT"
		return false
	}

	exists, _ = s.store.CheckClientExistsByCourtRef(ctx, details.correctCourtRef)

	if !exists {
		(*failedLines)[index] = "COURT_REF_MISMATCH"
		return false
	}

	return true
}

func (s *Service) processReversalUploadLine(ctx context.Context, tx *store.Tx, details reversalDetail) error {
	ledgerID, err := tx.CreateLedgerForCourtRef(ctx, store.CreateLedgerForCourtRefParams{
		CourtRef:     details.erroredCourtRef,
		Amount:       -details.amount,
		Type:         details.paymentType.Key(),
		Status:       "CONFIRMED",
		CreatedBy:    details.createdBy,
		BankDate:     details.bankDate,
		ReceivedDate: details.receivedDate,
	})

	if err != nil {
		return err
	}

	invoices, err := tx.GetInvoicesForReversalByCourtRef(ctx, details.erroredCourtRef)

	if err != nil {
		return err
	}

	remaining := details.amount

	for _, invoice := range invoices {
		allocationAmount := invoice.Received
		if allocationAmount > remaining {
			allocationAmount = remaining
		}

		var invoiceID pgtype.Int4
		_ = store.ToInt4(&invoiceID, invoice.ID)

		err = tx.CreateLedgerAllocation(ctx, store.CreateLedgerAllocationParams{
			InvoiceID: invoiceID,
			Amount:    -allocationAmount,
			Status:    "ALLOCATED",
			LedgerID:  ledgerID,
		})
		if err != nil {
			return err
		}

		remaining -= allocationAmount
		if remaining == 0 {
			break
		}
	}

	if remaining != 0 {
		err = tx.CreateLedgerAllocation(ctx, store.CreateLedgerAllocationParams{
			Amount:   remaining,
			Status:   "UNAPPLIED",
			LedgerID: ledgerID,
		})
		if err != nil {
			return err
		}
	}

	return nil
}
