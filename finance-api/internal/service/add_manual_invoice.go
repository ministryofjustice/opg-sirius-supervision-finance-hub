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
		data.Amount = shared.Nillable[int]{Value: 100, Valid: true}
		data.StartDate = data.RaisedDate
		data.EndDate = data.RaisedDate
	case shared.InvoiceTypeS2, shared.InvoiceTypeB2:
		data.EndDate = data.RaisedDate
		data.SupervisionLevel = shared.Nillable[string]{Value: "GENERAL", Valid: true}
	case shared.InvoiceTypeS3, shared.InvoiceTypeB3:
		data.EndDate = data.RaisedDate
		data.SupervisionLevel = shared.Nillable[string]{Value: "MINIMAL", Valid: true}
	}

	return data
}

func (s *Service) AddManualInvoice(ctx context.Context, clientId int, data shared.AddManualInvoice) error {
  data = processInvoiceData(data)
	validationsErrors := s.validateManualInvoice(data)

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
	counter, err := transaction.GetInvoiceCounter(ctx, strconv.Itoa(data.StartDate.Time.Year())+"InvoiceNumber")
	if err != nil {
		return err
	}

	invoiceRef := addLeadingZeros(counter)

	invoiceParams := store.AddInvoiceParams{
		PersonID:   pgtype.Int4{Int32: int32(clientId), Valid: true},
		Feetype:    data.InvoiceType.Key(),
		Reference:  data.InvoiceType.Key() + invoiceRef + "/" + strconv.Itoa(data.StartDate.Time.Year()%100),
		Startdate:  pgtype.Date{Time: data.StartDate.Time, Valid: true},
		Enddate:    pgtype.Date{Time: data.EndDate.Time, Valid: true},
		Amount:     int32(data.Amount),
		Raiseddate: pgtype.Date{Time: data.RaisedDate.Time, Valid: true},
		Source:     pgtype.Text{String: "Created manually", Valid: true},
		//TODO make sure we have correct createdby ID in ticket PFS-136
		CreatedbyID: pgtype.Int4{Int32: int32(1), Valid: true},
	}

	var invoice store.Invoice
	invoice, err = transaction.AddInvoice(ctx, invoiceParams)
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
		err = s.AddFeeReductionLedger(ctx, transaction, shared.ParseFeeReductionType(feeReduction.Type), feeReduction.FeeReductionID, int32(clientId), invoice)
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

func (s *Service) validateManualInvoice(data shared.AddManualInvoice) []string {
	var validationsErrors []string

	if data.InvoiceType.RequiresDateValidation() {
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
	return validationsErrors
}

func (s *Service) AddFeeReductionLedger(ctx context.Context, transaction *store.Queries, feeReductionFeeType shared.FeeReductionType, feeReductionId int32, clientId int32, invoice store.Invoice) error {
	amount := s.calculateFeeReduction(ctx, transaction, feeReductionFeeType, invoice)
	if amount != 0 {
		ledgerParams := store.CreateLedgerParams{
			ClientID:       clientId,
			Amount:         amount,
			Notes:          pgtype.Text{String: "Credit due to manual invoice " + strings.ToLower(feeReductionFeeType.Key()), Valid: true},
			Type:           "CREDIT " + feeReductionFeeType.Key(),
			Status:         "APPROVED",
			Method:         feeReductionFeeType.String() + " credit for invoice " + invoice.Reference,
			FeeReductionID: pgtype.Int4{Int32: feeReductionId, Valid: true},
			//TODO make sure we have correct createdby ID in ticket PFS-136
			CreatedbyID: pgtype.Int4{Int32: 1},
		}

		ledgerId, err := transaction.CreateLedger(ctx, ledgerParams)
		if err != nil {
			return err
		}

		allocationParams := store.CreateLedgerAllocationParams{
			LedgerID:  pgtype.Int4{Int32: ledgerId, Valid: true},
			InvoiceID: pgtype.Int4{Int32: invoice.ID, Valid: true},
			Amount:    amount,
			Status:    "ALLOCATED",
			Notes:     pgtype.Text{},
		}

		_, err = transaction.CreateLedgerAllocation(ctx, allocationParams)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) calculateFeeReduction(ctx context.Context, transaction *store.Queries, feeReductionFeeType shared.FeeReductionType, invoice store.Invoice) int32 {
	var amount int32
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
	default:
		amount = 0
	}
	return amount
}

func addLeadingZeros(counter string) string {
	strNumberLength := len(counter)
	if strNumberLength != 6 {
		numOfZeros := 6 - strNumberLength
		return strings.Repeat("0", numOfZeros) + counter
	}
	return counter
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
