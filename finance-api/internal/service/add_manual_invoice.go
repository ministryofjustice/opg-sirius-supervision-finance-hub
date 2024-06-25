package service

import (
	"context"
	"database/sql"
	"errors"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/opg-sirius-finance-hub/finance-api/internal/store"
	"github.com/opg-sirius-finance-hub/shared"
	"log"
	"strconv"
	"strings"
	"time"
)

func (s *Service) AddManualInvoice(clientId int, data shared.AddManualInvoice) error {
	var validationsErrors []string

	if data.InvoiceType.IsRaisedDateValid() {
		if !data.RaisedDate.Time.Before(time.Now()) {
			validationsErrors = append(validationsErrors, "RaisedDateForAnInvoice")
		}
	}

	isStartDateValid := validateStartDate(data.StartDate, data.EndDate)
	if !isStartDateValid {
		validationsErrors = append(validationsErrors, "StartDate")
	}

	isEndDateValid := validateEndDate(data.StartDate, data.EndDate)
	if !isEndDateValid {
		validationsErrors = append(validationsErrors, "EndDate")
	}

	if len(validationsErrors) != 0 {
		return shared.BadRequests{Reasons: validationsErrors}
	}

	ctx := context.Background()

	tx, err := s.tx.Begin(ctx)
	if err != nil {
		return err
	}

	defer func() {
		if err := tx.Rollback(ctx); !errors.Is(err, sql.ErrTxDone) {
			log.Println("Error rolling back add fee reduction transaction:", err)
		}
	}()

	transaction := s.store.WithTx(tx)
	getInvoiceCounterForYear, err := transaction.UpsertCounterForInvoiceRefYear(ctx, strconv.Itoa(data.StartDate.Time.Year())+"InvoiceNumber")
	if err != nil {
		return err
	}

	invoiceRef := addLeadingZeros(int(getInvoiceCounterForYear.Counter))

	addManualInvoiceQueryArgs := store.AddManualInvoiceParams{
		PersonID:      pgtype.Int4{Int32: int32(clientId), Valid: true},
		Feetype:       data.InvoiceType.Key(),
		Reference:     data.InvoiceType.Key() + invoiceRef + "/" + strconv.Itoa(data.StartDate.Time.Year()%100),
		Startdate:     pgtype.Date{Time: data.StartDate.Time, Valid: true},
		Enddate:       pgtype.Date{Time: data.EndDate.Time, Valid: true},
		Amount:        int32(data.Amount),
		Confirmeddate: pgtype.Date{Time: time.Now(), Valid: true},
		Raiseddate:    pgtype.Date{Time: data.RaisedDate.Time, Valid: true},
		Source:        pgtype.Text{String: "Created manually", Valid: true},
		Createddate:   pgtype.Date{Time: time.Now(), Valid: true},
	}

	var invoice store.Invoice
	invoice, err = transaction.AddManualInvoice(ctx, addManualInvoiceQueryArgs)
	if err != nil {
		return err
	}

	if data.SupervisionLevel != "" {
		addInvoiceRangeQueryArgs := store.AddInvoiceRangeParams{
			InvoiceID:        pgtype.Int4{Int32: invoice.ID, Valid: true},
			Supervisionlevel: strings.ToUpper(data.SupervisionLevel),
			Fromdate:         pgtype.Date{Time: data.StartDate.Time, Valid: true},
			Todate:           pgtype.Date{Time: data.EndDate.Time, Valid: true},
			Amount:           invoice.Amount,
		}

		err = transaction.AddInvoiceRange(ctx, addInvoiceRangeQueryArgs)
		if err != nil {
			return err
		}
	}

	AddFeeReductionToInvoiceParams := store.GetFeeReductionByInvoiceIdParams{
		ClientID: int32(clientId),
		ID:       invoice.ID,
	}

	feeReduction, _ := transaction.GetFeeReductionByInvoiceId(ctx, AddFeeReductionToInvoiceParams)

	if feeReduction.FeeReductionID != 0 {
		err = s.AddLedgerAndAllocations(feeReduction.Type, feeReduction.FeeReductionID, feeReduction.FinanceClientID.Int32, invoice, transaction, ctx)
		if err != nil {
			return err
		}
	}

	commitErr := tx.Commit(ctx)
	if commitErr != nil {
		return commitErr
	}

	return nil
}

func (s *Service) AddLedgerAndAllocations(feeReductionFeeType string, feeReductionId int32, feeReductionFinanceClientID int32, invoice store.Invoice, transaction *store.Queries, ctx context.Context) error {
	var amount int32 = 0
	switch strings.ToLower(feeReductionFeeType) {
	case "exemption", "hardship":
		amount = invoice.Amount
	case "remission":
		invoiceFeeRangeParams := store.GetInvoiceFeeRangeAmountParams{
			InvoiceID:        pgtype.Int4{Int32: invoice.ID, Valid: true},
			Supervisionlevel: "GENERAL",
		}
		amount, _ = transaction.GetInvoiceFeeRangeAmount(ctx, invoiceFeeRangeParams)
	}
	if amount != 0 {
		ledgerQueryArgs := store.CreateLedgerForFeeReductionParams{
			Method:          feeReductionFeeType + " credit for invoice " + invoice.Reference,
			Amount:          amount,
			Notes:           pgtype.Text{String: "Credit due to manual invoice " + feeReductionFeeType},
			Type:            "CREDIT " + feeReductionFeeType,
			FinanceClientID: pgtype.Int4{Int32: feeReductionFinanceClientID, Valid: true},
			FeeReductionID:  pgtype.Int4{Int32: feeReductionId, Valid: true},
			//TODO make sure we have correct createdby ID in ticket PFS-88
			CreatedbyID: pgtype.Int4{Int32: 1},
		}
		var ledger store.Ledger
		ledger, err := transaction.CreateLedgerForFeeReduction(ctx, ledgerQueryArgs)
		if err != nil {
			return err
		}

		queryArgs := store.CreateLedgerAllocationForFeeReductionParams{
			LedgerID:  pgtype.Int4{Int32: ledger.ID, Valid: true},
			InvoiceID: pgtype.Int4{Int32: invoice.ID, Valid: true},
			Amount:    ledger.Amount,
		}

		_, err = transaction.CreateLedgerAllocationForFeeReduction(ctx, queryArgs)
		if err != nil {
			return err
		}
	}
	return nil
}

func addLeadingZeros(number int) string {
	strNumber := strconv.Itoa(number)
	strNumberLength := len(strNumber)
	if strNumberLength != 6 {
		numOfZeros := 6 - strNumberLength
		return strings.Repeat("0", numOfZeros) + strNumber
	}
	return strNumber
}

func validateEndDate(startDate *shared.Date, endDate *shared.Date) bool {
	return !endDate.Time.Before(startDate.Time)
}

func validateStartDate(startDate *shared.Date, endDate *shared.Date) bool {
	if startDate.Time.After(endDate.Time) {
		return false
	}

	if !startDate.IsSameFinancialYear(endDate) {
		return false
	}

	return true
}
