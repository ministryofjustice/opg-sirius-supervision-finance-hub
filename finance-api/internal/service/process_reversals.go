package service

import (
	"context"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/auth"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/store"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
	"time"
)

func (s *Service) ProcessPaymentReversals(ctx context.Context, records [][]string, uploadType shared.ReportUploadType) (map[int]string, error) {
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

			if details != (shared.ReversalDetails{}) {
				if !s.validateReversalLine(ctx, details, index, &failedLines) {
					continue
				}

				err = s.ProcessReversalUploadLine(ctx, tx, details)
				if err != nil {
					return nil, err
				}

				if uploadType == shared.ReportTypeUploadMisappliedPayments {
					err = s.ProcessPaymentsUploadLine(ctx, tx, shared.PaymentDetails{
						Amount:       details.Amount,
						BankDate:     details.BankDate,
						CourtRef:     details.CorrectCourtRef,
						LedgerType:   details.PaymentType,
						ReceivedDate: details.ReceivedDate,
						CreatedBy:    details.CreatedBy,
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

func getReversalLines(ctx context.Context, record []string, uploadType shared.ReportUploadType, index int, failedLines *map[int]string) shared.ReversalDetails {
	var (
		paymentType     shared.TransactionType
		erroredCourtRef pgtype.Text
		correctCourtRef pgtype.Text
		bankDate        pgtype.Date
		receivedDate    pgtype.Timestamp
		createdBy       pgtype.Int4
		pisNumber       pgtype.Int4
		amount          int32
	)

	switch uploadType {
	case shared.ReportTypeUploadMisappliedPayments:
		paymentType = shared.ParseTransactionType(record[0])
		if !paymentType.Valid() {
			(*failedLines)[index] = "PAYMENT_TYPE_PARSE_ERROR"
			return shared.ReversalDetails{}
		}

		_ = erroredCourtRef.Scan(record[1])
		_ = correctCourtRef.Scan(record[2])

		bd, err := time.Parse("2006-01-02", record[3])
		if err != nil {
			(*failedLines)[index] = "BANK_DATE_PARSE_ERROR"
			return shared.ReversalDetails{}
		}
		_ = bankDate.Scan(bd)

		rd, err := time.Parse("2006-01-02", record[4])
		if err != nil {
			(*failedLines)[index] = "RECEIVED_DATE_PARSE_ERROR"
			return shared.ReversalDetails{}
		}
		_ = receivedDate.Scan(rd)

		amount, err = parseAmount(record[5])
		if err != nil {
			(*failedLines)[index] = "AMOUNT_PARSE_ERROR"
			return shared.ReversalDetails{}
		}

		_ = pisNumber.Scan(record[6]) // will have no value for non-cheque payments

	default:
		(*failedLines)[index] = "UNKNOWN_UPLOAD_TYPE"
	}

	_ = store.ToInt4(&createdBy, ctx.(auth.Context).User.ID)

	return shared.ReversalDetails{
		PaymentType:     paymentType,
		ErroredCourtRef: erroredCourtRef,
		CorrectCourtRef: correctCourtRef,
		BankDate:        bankDate,
		ReceivedDate:    receivedDate,
		Amount:          amount,
		PisNumber:       pisNumber,
	}
}

func (s *Service) validateReversalLine(ctx context.Context, details shared.ReversalDetails, index int, failedLines *map[int]string) bool {
	params := store.CheckDuplicateLedgerParams{
		CourtRef:     details.ErroredCourtRef,
		Amount:       details.Amount,
		Type:         details.PaymentType.Key(),
		BankDate:     details.BankDate,
		ReceivedDate: details.ReceivedDate,
		PisNumber:    details.PisNumber,
	}
	exists, _ := s.store.CheckDuplicateLedger(ctx, params)

	if !exists {
		(*failedLines)[index] = "NO_MATCHED_PAYMENT"
		return false
	}

	exists, _ = s.store.CheckClientExistsByCourtRef(ctx, details.CorrectCourtRef)

	if !exists {
		(*failedLines)[index] = "COURT_REF_MISMATCH"
		return false
	}

	return true
}

func (s *Service) ProcessReversalUploadLine(ctx context.Context, tx *store.Tx, details shared.ReversalDetails) error {
	ledgerID, err := tx.CreateLedgerForCourtRef(ctx, store.CreateLedgerForCourtRefParams{
		CourtRef:     details.ErroredCourtRef,
		Amount:       -details.Amount,
		Type:         details.PaymentType.Key(),
		Status:       "CONFIRMED",
		CreatedBy:    details.CreatedBy,
		BankDate:     details.BankDate,
		ReceivedDate: details.ReceivedDate,
		PisNumber:    details.PisNumber,
	})

	if err != nil {
		return err
	}

	invoices, err := tx.GetInvoicesForReversalByCourtRef(ctx, details.ErroredCourtRef)

	if err != nil {
		return err
	}

	remaining := details.Amount

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
