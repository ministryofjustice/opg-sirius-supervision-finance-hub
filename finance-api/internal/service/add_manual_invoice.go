package service

import (
	"context"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/opg-sirius-finance-hub/finance-api/internal/store"
	"github.com/opg-sirius-finance-hub/shared"
	"strconv"
	"strings"
	"time"
)

func processInvoiceData(data shared.AddManualInvoice) shared.AddManualInvoice {
	switch data.InvoiceType {
	case shared.InvoiceTypeAD:
		data.Amount = shared.NillableInt{Value: 100, Valid: true}
		data.StartDate = data.RaisedDate
		data.EndDate = data.RaisedDate
	case shared.InvoiceTypeS2, shared.InvoiceTypeB2:
		if data.RaisedYear.Value != 0 {
			data.RaisedDate = shared.NillableDate{
				Value: shared.NewDate(strconv.Itoa(data.RaisedYear.Value) + "-03" + "-31"),
				Valid: true,
			}
		}
		data.EndDate = data.RaisedDate
		data.SupervisionLevel = shared.NillableString{Value: "GENERAL", Valid: true}
	case shared.InvoiceTypeS3, shared.InvoiceTypeB3:
		if data.RaisedYear.Value != 0 {
			data.RaisedDate = shared.NillableDate{
				Value: shared.NewDate(strconv.Itoa(data.RaisedYear.Value) + "-03" + "-31"),
				Valid: true,
			}
		}
		data.EndDate = data.RaisedDate
		data.SupervisionLevel = shared.NillableString{Value: "MINIMAL", Valid: true}
	}

	return data
}

func (s *Service) AddManualInvoice(ctx context.Context, clientId int, data shared.AddManualInvoice) (txErr error) {
	var validationsErrors []string

	data = processInvoiceData(data)

	if data.InvoiceType.RequiresDateValidation() {
		if !data.RaisedDate.Value.Time.Before(time.Now()) {
			validationsErrors = append(validationsErrors, "RaisedDateForAnInvoice")
		}
	}

	isStartDateValid := validateStartDate(&data.StartDate.Value, &data.EndDate.Value)
	if !isStartDateValid {
		validationsErrors = append(validationsErrors, "StartDate")
	}

	isEndDateValid := validateEndDate(&data.StartDate.Value, &data.EndDate.Value)
	if !isEndDateValid {
		validationsErrors = append(validationsErrors, "EndDate")
	}

	if len(validationsErrors) != 0 {
		return shared.BadRequests{Reasons: validationsErrors}
	}

	ctx, cancelTx := context.WithCancel(ctx)
	defer cancelTx()

	tx, err := s.tx.Begin(ctx)
	if err != nil {
		return err
	}

	transaction := s.store.WithTx(tx)
	getInvoiceCounterForYear, err := transaction.UpsertCounterForInvoiceRefYear(ctx, strconv.Itoa(data.StartDate.Value.Time.Year())+"InvoiceNumber")
	if err != nil {
		return err
	}

	invoiceRef := addLeadingZeros(int(getInvoiceCounterForYear.Counter))

	addManualInvoiceQueryArgs := store.AddManualInvoiceParams{
		PersonID:      pgtype.Int4{Int32: int32(clientId), Valid: true},
		Feetype:       data.InvoiceType.Key(),
		Reference:     data.InvoiceType.Key() + invoiceRef + "/" + strconv.Itoa(data.StartDate.Value.Time.Year()%100),
		Startdate:     pgtype.Date{Time: data.StartDate.Value.Time, Valid: true},
		Enddate:       pgtype.Date{Time: data.EndDate.Value.Time, Valid: true},
		Amount:        int32(data.Amount.Value),
		Confirmeddate: pgtype.Date{Time: time.Now(), Valid: true},
		Raiseddate:    pgtype.Date{Time: data.RaisedDate.Value.Time, Valid: true},
		Source:        pgtype.Text{String: "Created manually", Valid: true},
		Createddate:   pgtype.Date{Time: time.Now(), Valid: true},
	}

	var invoice store.Invoice
	invoice, err = transaction.AddManualInvoice(ctx, addManualInvoiceQueryArgs)
	if err != nil {
		return err
	}

	if data.SupervisionLevel.Valid {
		addInvoiceRangeQueryArgs := store.AddInvoiceRangeParams{
			InvoiceID:        pgtype.Int4{Int32: invoice.ID, Valid: true},
			Supervisionlevel: strings.ToUpper(data.SupervisionLevel.Value),
			Fromdate:         pgtype.Date{Time: data.StartDate.Value.Time, Valid: true},
			Todate:           pgtype.Date{Time: data.EndDate.Value.Time, Valid: true},
			Amount:           invoice.Amount,
		}

		err = transaction.AddInvoiceRange(ctx, addInvoiceRangeQueryArgs)
		if err != nil {
			return err
		}
	}

	AddFeeReductionToInvoiceParams := store.GetFeeReductionForDateParams{
		ClientID:     int32(clientId),
		Datereceived: invoice.Raiseddate,
	}

	feeReduction, _ := transaction.GetFeeReductionForDate(ctx, AddFeeReductionToInvoiceParams)

	if feeReduction.FeeReductionID != 0 {
		err = s.AddLedgerAndAllocations(shared.ParseFeeReductionType(feeReduction.Type), feeReduction.FeeReductionID, feeReduction.FinanceClientID.Int32, invoice, transaction, ctx)
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

func (s *Service) AddLedgerAndAllocations(feeReductionFeeType shared.FeeReductionType, feeReductionId int32, feeReductionFinanceClientID int32, invoice store.Invoice, transaction *store.Queries, ctx context.Context) error {
	var amount int32 = 0
	switch feeReductionFeeType {
	case shared.FeeReductionTypeExemption, shared.FeeReductionTypeHardship:
		amount = invoice.Amount
	case shared.FeeReductionTypeRemission:
		invoiceFeeRangeParams := store.GetInvoiceFeeRangeAmountParams{
			InvoiceID:        pgtype.Int4{Int32: invoice.ID, Valid: true},
			Supervisionlevel: "GENERAL",
		}
		amount, _ = transaction.GetInvoiceFeeRangeAmount(ctx, invoiceFeeRangeParams)
		if invoice.Feetype == "AD" {
			amount = invoice.Amount / 2
		}
	}
	if amount != 0 {
		ledgerQueryArgs := store.CreateLedgerForFeeReductionParams{
			Method:          feeReductionFeeType.String() + " credit for invoice " + invoice.Reference,
			Amount:          amount,
			Notes:           pgtype.Text{String: "Credit due to manual invoice " + strings.ToLower(feeReductionFeeType.Key()), Valid: true},
			Type:            "CREDIT " + feeReductionFeeType.Key(),
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
