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

func (s *Service) AddManualInvoice(ctx context.Context, clientId int, data shared.AddManualInvoice) error {
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

	AddFeeReductionToInvoiceParams := store.GetFeeReductionForDateParams{
		ClientID:     int32(clientId),
		Datereceived: invoice.Raiseddate,
	}

	feeReduction, _ := transaction.GetFeeReductionForDate(ctx, AddFeeReductionToInvoiceParams)

	if feeReduction.FeeReductionID != 0 {
		amount := calculateFeeReduction(ctx, transaction, shared.ParseFeeReductionType(feeReduction.Type), invoice.ID, invoice.Amount, invoice.Feetype)
		if amount > 0 {
			err = addLedger(ctx, transaction, amount, shared.ParseFeeReductionType(feeReduction.Type), feeReduction.FeeReductionID, int32(clientId), invoice.ID, invoice.Amount)
			if err != nil {
				return err
			}
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
