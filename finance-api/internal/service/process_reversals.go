package service

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/auth"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/store"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/validation"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
)

func (s *Service) ProcessPaymentReversals(ctx context.Context, records [][]string, uploadType shared.ReportUploadType, uploadDate shared.Date) (map[int]string, error) {
	failedLines := make(map[int]string)

	tx, err := s.BeginStoreTx(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	var processedRecords []shared.ReversalDetails

	for index, record := range records {
		if index != 0 && safeRead(record, 0) != "" {
			details := getReversalLines(ctx, record, uploadType, index, &failedLines)

			if details != (shared.ReversalDetails{}) {
				if !s.validateReversalLine(ctx, details, uploadType, processedRecords, index, &failedLines) {
					continue
				}

				if uploadType == shared.ReportTypeUploadMisappliedPayments {
					if !s.validateApplyLine(ctx, details, index, &failedLines) {
						continue
					}
				}

				if uploadType == shared.ReportTypeUploadReverseFulfilledRefunds {
					var (
						bankDate     pgtype.Date
						receivedDate pgtype.Timestamp
					)
					_ = bankDate.Scan(uploadDate.Time)
					_ = receivedDate.Scan(uploadDate.Time)
					details.BankDate = bankDate
					details.ReceivedDate = receivedDate
				}

				err = s.ProcessReversalUploadLine(ctx, tx, details)
				if err != nil {
					return nil, err
				}

				processedRecords = append(processedRecords, details)

				if uploadType == shared.ReportTypeUploadMisappliedPayments {
					_, err = s.ProcessPaymentsUploadLine(ctx, tx, shared.PaymentDetails{
						Amount:       details.Amount,
						BankDate:     details.BankDate,
						CourtRef:     details.CorrectCourtRef,
						LedgerType:   details.PaymentType,
						ReceivedDate: details.ReceivedDate,
						CreatedBy:    details.CreatedBy,
						PisNumber:    details.PisNumber,
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
		paymentType = shared.ParseTransactionType(safeRead(record, 0))
		if !paymentType.Valid() {
			(*failedLines)[index] = validation.UploadErrorPaymentTypeParse
			return shared.ReversalDetails{}
		}

		_ = erroredCourtRef.Scan(safeRead(record, 1))
		_ = correctCourtRef.Scan(safeRead(record, 2))

		bd, err := time.Parse("02/01/2006", safeRead(record, 3))
		if err != nil {
			(*failedLines)[index] = validation.UploadErrorDateParse
			return shared.ReversalDetails{}
		}
		_ = bankDate.Scan(bd)

		rd, err := time.Parse("02/01/2006", safeRead(record, 4))
		if err != nil {
			(*failedLines)[index] = validation.UploadErrorDateParse
			return shared.ReversalDetails{}
		}
		_ = receivedDate.Scan(rd)

		amount, err = parseAmount(safeRead(record, 5))
		if err != nil {
			(*failedLines)[index] = validation.UploadErrorAmountParse
			return shared.ReversalDetails{}
		}

		_ = pisNumber.Scan(safeRead(record, 6)) // will have no value for non-cheque payments
	case shared.ReportTypeUploadDuplicatedPayments:
		paymentType = shared.ParseTransactionType(safeRead(record, 0))
		if !paymentType.Valid() {
			(*failedLines)[index] = validation.UploadErrorPaymentTypeParse
			return shared.ReversalDetails{}
		}

		_ = erroredCourtRef.Scan(safeRead(record, 1))

		bd, err := time.Parse("02/01/2006", safeRead(record, 2))
		if err != nil {
			(*failedLines)[index] = validation.UploadErrorDateParse
			return shared.ReversalDetails{}
		}
		_ = bankDate.Scan(bd)

		rd, err := time.Parse("02/01/2006", safeRead(record, 3))
		if err != nil {
			(*failedLines)[index] = validation.UploadErrorDateParse
			return shared.ReversalDetails{}
		}
		_ = receivedDate.Scan(rd)

		amount, err = parseAmount(safeRead(record, 4))
		if err != nil {
			(*failedLines)[index] = validation.UploadErrorAmountParse
			return shared.ReversalDetails{}
		}

		_ = pisNumber.Scan(safeRead(record, 5)) // will have no value for non-cheque payments

	case shared.ReportTypeUploadBouncedCheque:
		paymentType = shared.TransactionTypeSupervisionChequePayment
		_ = erroredCourtRef.Scan(safeRead(record, 0))

		bd, err := time.Parse("02/01/2006", safeRead(record, 1))
		if err != nil {
			(*failedLines)[index] = validation.UploadErrorDateParse
			return shared.ReversalDetails{}
		}
		_ = bankDate.Scan(bd)

		rd, err := time.Parse("02/01/2006", safeRead(record, 2))
		if err != nil {
			(*failedLines)[index] = validation.UploadErrorDateParse
			return shared.ReversalDetails{}
		}
		_ = receivedDate.Scan(rd)

		amount, err = parseAmount(safeRead(record, 3))
		if err != nil {
			(*failedLines)[index] = validation.UploadErrorAmountParse
			return shared.ReversalDetails{}
		}

		_ = pisNumber.Scan(safeRead(record, 4))

	case shared.ReportTypeUploadFailedDirectDebitCollections:
		paymentType = shared.TransactionTypeDirectDebitPayment
		_ = erroredCourtRef.Scan(safeRead(record, 0))

		bd, err := time.Parse("02/01/2006", safeRead(record, 1))
		if err != nil {
			(*failedLines)[index] = validation.UploadErrorDateParse
			return shared.ReversalDetails{}
		}
		_ = bankDate.Scan(bd)

		rd, err := time.Parse("02/01/2006", safeRead(record, 2))
		if err != nil {
			(*failedLines)[index] = validation.UploadErrorDateParse
			return shared.ReversalDetails{}
		}
		_ = receivedDate.Scan(rd)

		amount, err = parseAmount(safeRead(record, 3))
		if err != nil {
			(*failedLines)[index] = validation.UploadErrorAmountParse
			return shared.ReversalDetails{}
		}

	case shared.ReportTypeUploadReverseFulfilledRefunds:
		paymentType = shared.TransactionTypeRefund
		_ = erroredCourtRef.Scan(safeRead(record, 0))

		bd, err := time.Parse("02/01/2006", safeRead(record, 2))
		if err != nil {
			(*failedLines)[index] = validation.UploadErrorDateParse
			return shared.ReversalDetails{}
		}
		_ = bankDate.Scan(bd)
		_ = receivedDate.Scan(bd)

		amount, err = parseAmount(safeRead(record, 1))

		if err != nil {
			(*failedLines)[index] = validation.UploadErrorAmountParse
			return shared.ReversalDetails{}
		}

	default:
		(*failedLines)[index] = validation.UploadErrorUnknownUploadType
	}

	_ = store.ToInt4(&createdBy, ctx.(auth.Context).User.ID)

	return shared.ReversalDetails{
		PaymentType:      paymentType,
		ErroredCourtRef:  erroredCourtRef,
		CorrectCourtRef:  correctCourtRef,
		BankDate:         bankDate,
		ReceivedDate:     receivedDate,
		Amount:           amount,
		PisNumber:        pisNumber,
		CreatedBy:        createdBy,
		SkipBankDate:     shouldSkipBankDate(uploadType, pisNumber),
		SkipReceivedDate: shouldSkipReceivedDate(uploadType),
	}
}

func (s *Service) validateReversalLine(ctx context.Context, details shared.ReversalDetails, uploadType shared.ReportUploadType, processedRecords []shared.ReversalDetails, index int, failedLines *map[int]string) bool {
	ledgerCount, _ := s.store.CountDuplicateLedger(ctx, store.CountDuplicateLedgerParams{
		CourtRef:         details.ErroredCourtRef,
		Amount:           details.Amount,
		Type:             details.PaymentType.Key(),
		BankDate:         details.BankDate,
		ReceivedDate:     details.ReceivedDate,
		PisNumber:        details.PisNumber,
		SkipBankDate:     details.SkipBankDate,
		SkipReceivedDate: details.SkipReceivedDate,
	})

	if ledgerCount == 0 {
		(*failedLines)[index] = validation.UploadErrorNoMatchedPayment
		return false
	}

	if uploadType == shared.ReportTypeUploadMisappliedPayments {
		exists, _ := s.store.CheckClientExistsByCourtRef(ctx, details.CorrectCourtRef)
		if !exists {
			(*failedLines)[index] = validation.UploadErrorReversalClientNotFound
			return false
		}
	}

	reversalCount, _ := s.store.CountDuplicateLedger(ctx, store.CountDuplicateLedgerParams{
		CourtRef:     details.ErroredCourtRef,
		Amount:       -details.Amount,
		Type:         details.PaymentType.Key(),
		BankDate:     details.BankDate,
		ReceivedDate: details.ReceivedDate,
		PisNumber:    details.PisNumber,
	})

	if reversalCount >= ledgerCount {
		(*failedLines)[index] = validation.UploadErrorDuplicateReversal
		return false
	}

	if !hasPaymentToReverse(processedRecords, details, int(ledgerCount)) {
		(*failedLines)[index] = validation.UploadErrorDuplicateReversal
		return false
	}

	reversible, _ := s.store.GetReversibleBalanceByCourtRef(ctx, details.ErroredCourtRef)

	if reversible < details.Amount {
		(*failedLines)[index] = validation.UploadErrorMaximumDebt
		return false
	}

	return true
}

func (s *Service) validateApplyLine(ctx context.Context, details shared.ReversalDetails, index int, failedLines *map[int]string) bool {
	exists, _ := s.store.CheckClientExistsByCourtRef(ctx, details.CorrectCourtRef)

	if !exists {
		(*failedLines)[index] = validation.UploadErrorReversalClientNotFound
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
		Notes:        details.Notes,
	})

	if err != nil {
		return err
	}

	// get credit balance and apply there first
	credit, _ := tx.GetCreditBalanceByCourtRef(ctx, details.ErroredCourtRef)

	remaining := details.Amount

	if credit > 0 {
		allocationAmount := credit
		if allocationAmount > remaining {
			allocationAmount = remaining
		}
		err = tx.CreateLedgerAllocation(ctx, store.CreateLedgerAllocationParams{
			Amount:   allocationAmount,
			Status:   "UNAPPLIED",
			LedgerID: ledgerID,
		})
		if err != nil {
			return err
		}

		remaining -= allocationAmount
	}

	if remaining == 0 {
		return nil
	}

	invoices, err := tx.GetInvoicesForReversalByCourtRef(ctx, details.ErroredCourtRef)

	if err != nil {
		return err
	}

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
		s.Logger(ctx).Error("process reversal upload line failed as amount remaining after applying to all available invoices", "amount", remaining)
		return errors.New("unexpected error - remaining not zero")
	}

	return nil
}

func hasPaymentToReverse(processedRecords []shared.ReversalDetails, details shared.ReversalDetails, totalPayments int) bool {
	reversals := 0
	for _, s := range processedRecords {
		if s == details {
			reversals++
		}
	}
	return reversals < totalPayments
}

/**
 * shouldSkipBankDate returns true if the upload is for failed direct debit or cheque payment, as they are different
 * transactions on OPG bank statements
 */
func shouldSkipBankDate(uploadType shared.ReportUploadType, pisNumber pgtype.Int4) pgtype.Bool {
	var skip pgtype.Bool
	_ = skip.Scan(uploadType == shared.ReportTypeUploadFailedDirectDebitCollections || pisNumber.Valid)
	return skip
}

func shouldSkipReceivedDate(uploadType shared.ReportUploadType) pgtype.Bool {
	var skip pgtype.Bool
	_ = skip.Scan(uploadType == shared.ReportTypeUploadReverseFulfilledRefunds)
	return skip
}
