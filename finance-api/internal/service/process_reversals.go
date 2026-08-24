package service

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/auth"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/event"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/store"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/validation"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
)

type reversalUploadLine struct {
	index   int
	details shared.ReversalDetails
}

func (s *Service) ProcessPaymentReversals(ctx context.Context, records [][]string, uploadType shared.ReportUploadType) (map[int]string, error) {
	failedLines := make(map[int]string)
	clientLines := make(map[string][]reversalUploadLine)

	for index, record := range records {
		if index != 0 && safeRead(record, 0) != "" {
			details := getReversalLines(ctx, record, uploadType, index, &failedLines)

			if details != (shared.ReversalDetails{}) {
				clientLines[details.ErroredCourtRef.String] = append(clientLines[details.ErroredCourtRef.String], reversalUploadLine{
					index:   index,
					details: details,
				})
			}
		}
	}

	for _, lines := range clientLines {
		s.processReversalsForClient(ctx, uploadType, lines, &failedLines)
	}

	return failedLines, nil
}

func (s *Service) processReversalsForClient(ctx context.Context, uploadType shared.ReportUploadType, lines []reversalUploadLine, failedLines *map[int]string) {
	var (
		clientIDs        []int32
		processedRecords []shared.ReversalDetails
		validLines       []reversalUploadLine
	)

	for _, line := range lines {
		if !s.validateReversalLine(ctx, line.details, uploadType, processedRecords, line.index, failedLines) {
			continue
		}

		if uploadType == shared.ReportTypeUploadMisappliedPayments && !s.validateApplyLine(ctx, line.details, line.index, failedLines) {
			continue
		}

		// we know client exists because validateReversalLine has already checked
		client, _ := s.store.GetClientIdsByCourtRef(ctx, line.details.ErroredCourtRef)
		clientIDs = append(clientIDs, client.ClientID)

		if uploadType == shared.ReportTypeUploadMisappliedPayments {
			// we know client exists because validateApplyLine has already checked
			correctClient, _ := s.store.GetClientIdsByCourtRef(ctx, line.details.CorrectCourtRef)
			clientIDs = append(clientIDs, correctClient.ClientID)
		}

		validLines = append(validLines, line)
		processedRecords = append(processedRecords, line.details)
	}

	if len(validLines) == 0 {
		return
	}

	tx, err := s.BeginStoreTxForClients(ctx, clientIDs)
	if err != nil {
		s.markReversalLinesFailed(validLines, failedLines, validation.UploadErrorProcessing)
		return
	}
	defer tx.Rollback(ctx)

	for _, line := range validLines {
		err = s.ProcessReversalUploadLine(ctx, tx, line.details)
		if err != nil {
			s.markReversalLinesFailed(validLines, failedLines, validation.UploadErrorProcessing)
			return
		}

		if uploadType == shared.ReportTypeUploadMisappliedPayments {
			_, err = s.ProcessPaymentsUploadLine(ctx, tx, shared.PaymentDetails{
				Amount:       line.details.Amount,
				BankDate:     line.details.BankDate,
				CourtRef:     line.details.CorrectCourtRef,
				LedgerType:   line.details.PaymentType,
				ReceivedDate: line.details.ReceivedDate,
				CreatedBy:    line.details.CreatedBy,
				PisNumber:    line.details.PisNumber,
			})
			if err != nil {
				s.markReversalLinesFailed(validLines, failedLines, validation.UploadErrorProcessing)
				return
			}
		}
	}

	err = tx.Commit(ctx)
	if err != nil {
		s.markReversalLinesFailed(validLines, failedLines, validation.UploadErrorProcessing)
	}
}

func (s *Service) markReversalLinesFailed(lines []reversalUploadLine, failedLines *map[int]string, reason string) {
	for _, line := range lines {
		if _, exists := (*failedLines)[line.index]; exists {
			continue
		}

		(*failedLines)[line.index] = reason
	}
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
	default:
		(*failedLines)[index] = validation.UploadErrorUnknownUploadType
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
		CreatedBy:       createdBy,
		SkipBankDate:    shouldSkipBankDate(uploadType, pisNumber),
	}
}

func (s *Service) validateReversalLine(ctx context.Context, details shared.ReversalDetails, uploadType shared.ReportUploadType, processedRecords []shared.ReversalDetails, index int, failedLines *map[int]string) bool {
	ledgerCount, err := s.store.CountDuplicateLedger(ctx, store.CountDuplicateLedgerParams{
		CourtRef:     details.ErroredCourtRef,
		Amount:       details.Amount,
		Type:         details.PaymentType.Key(),
		BankDate:     details.BankDate,
		ReceivedDate: details.ReceivedDate,
		PisNumber:    details.PisNumber,
		SkipBankDate: details.SkipBankDate,
	})
	if err != nil {
		(*failedLines)[index] = validation.UploadErrorProcessing
		return false
	}

	if ledgerCount == 0 {
		(*failedLines)[index] = validation.UploadErrorNoMatchedPayment
		return false
	}

	if uploadType == shared.ReportTypeUploadMisappliedPayments {
		exists, err := s.store.CheckClientExistsByCourtRef(ctx, details.CorrectCourtRef)
		if err != nil {
			(*failedLines)[index] = validation.UploadErrorProcessing
			return false
		}
		if !exists {
			(*failedLines)[index] = validation.UploadErrorReversalClientNotFound
			return false
		}
	}

	reversalCount, err := s.store.CountDuplicateLedger(ctx, store.CountDuplicateLedgerParams{
		CourtRef:     details.ErroredCourtRef,
		Amount:       -details.Amount,
		Type:         details.PaymentType.Key(),
		BankDate:     details.BankDate,
		ReceivedDate: details.ReceivedDate,
		PisNumber:    details.PisNumber,
	})
	if err != nil {
		(*failedLines)[index] = validation.UploadErrorProcessing
		return false
	}

	if reversalCount >= ledgerCount {
		(*failedLines)[index] = validation.UploadErrorDuplicateReversal
		return false
	}

	if !hasPaymentToReverse(processedRecords, details, int(ledgerCount)) {
		(*failedLines)[index] = validation.UploadErrorDuplicateReversal
		return false
	}

	reversible, err := s.store.GetReversibleBalanceByCourtRef(ctx, details.ErroredCourtRef)
	if err != nil {
		(*failedLines)[index] = validation.UploadErrorProcessing
		return false
	}

	if reversible < details.Amount {
		(*failedLines)[index] = validation.UploadErrorMaximumDebt
		return false
	}

	return true
}

func (s *Service) validateApplyLine(ctx context.Context, details shared.ReversalDetails, index int, failedLines *map[int]string) bool {
	exists, err := s.store.CheckClientExistsByCourtRef(ctx, details.CorrectCourtRef)
	if err != nil {
		(*failedLines)[index] = validation.UploadErrorProcessing
		return false
	}

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

	if details.PaymentType == shared.TransactionTypeRefund {
		return tx.CreateLedgerAllocation(ctx, store.CreateLedgerAllocationParams{
			Amount:   remaining,
			Status:   "UNAPPLIED",
			LedgerID: ledgerID,
		})
	}

	if remaining != 0 {
		s.Logger(ctx).Error("process reversal upload line failed as amount remaining after applying to all available invoices", "amount", remaining)
		return errors.New("unexpected error - remaining not zero")
	}

	if details.PaymentType == shared.TransactionTypeDirectDebitPayment {
		client, err := tx.GetClientIdsByCourtRef(ctx, details.ErroredCourtRef)
		if err != nil {
			return err
		}
		err = s.dispatch.DirectDebitCollectionFailed(ctx, event.DirectDebitCollectionFailed{
			ClientID: int(client.ClientID),
		})
		if err != nil {
			s.Logger(ctx).Error("error dispatching \"direct-debit-collection-failed\" event", "error", err)
		}
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
 * shouldSkipBankDate returns true if the upload is for failed Direct Debit or cheque payment, as they are different
 * transactions on OPG bank statements
 */
func shouldSkipBankDate(uploadType shared.ReportUploadType, pisNumber pgtype.Int4) pgtype.Bool {
	var skip pgtype.Bool
	_ = skip.Scan(uploadType == shared.ReportTypeUploadFailedDirectDebitCollections || pisNumber.Valid)
	return skip
}
