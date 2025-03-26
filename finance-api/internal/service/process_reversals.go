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
	paymentType     string
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

				err := s.processReversalUploadLine(ctx, tx, details, index, &failedLines)
				if err != nil {
					return nil, err
				}
				// if misapplied payment, run again for correct client
				// need to ensure reversal took place and no failed lines added
				// add note in the ledger for billing history?
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
		paymentType     string
		erroredCourtRef pgtype.Text
		correctCourtRef pgtype.Text
		bankDate        pgtype.Date
		receivedDate    pgtype.Timestamp
		createdBy       pgtype.Int4
		amount          int32
	)

	switch uploadType {
	case shared.ReportTypeUploadMisappliedPayments:
		paymentType = record[0]
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
		Type:         details.paymentType,
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

func (s *Service) processReversalUploadLine(ctx context.Context, tx *store.Tx, details reversalDetail, index int, failedLines *map[int]string) error {
	// match payments and validate:
	//   - payment exists
	var (
		erroredCourtRef pgtype.Text
		correctCourtRef pgtype.Text
		bankDate        pgtype.Date
		receivedDate    pgtype.Timestamp
		createdBy       pgtype.Int4
	)

	_ = erroredCourtRef.Scan(details.erroredCourtRef)
	_ = correctCourtRef.Scan(details.correctCourtRef)
	_ = bankDate.Scan(details.bankDate)
	_ = receivedDate.Scan(details.receivedDate)
	_ = store.ToInt4(&createdBy, ctx.(auth.Context).User.ID)

	//   - matched by date or PIS if cheque
	//   - if misapplied, new client exists
	// for each payment:
	//   - create ledger
	//   - create the reversed allocation
	//   - update the payment with the reversed allocation
	//   - run reapply
	//   - if misapplied:
	//		- create ledger for correct client
	//		- create allocation for correct client
	//		- run reapply
	return nil
}
