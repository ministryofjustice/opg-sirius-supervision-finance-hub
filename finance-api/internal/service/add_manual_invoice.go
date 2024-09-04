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
		data.Amount = shared.Nillable[int]{Value: 10000, Valid: true}
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
	counter, err := transaction.GetInvoiceCounter(ctx, strconv.Itoa(data.StartDate.Value.Time.Year())+"InvoiceNumber")
	if err != nil {
		return err
	}

	invoiceRef := addLeadingZeros(counter)

	invoiceParams := store.AddInvoiceParams{
		PersonID:   pgtype.Int4{Int32: int32(clientId), Valid: true},
		Feetype:    data.InvoiceType.Key(),
		Reference:  data.InvoiceType.Key() + invoiceRef + "/" + strconv.Itoa(data.StartDate.Value.Time.Year()%100),
		Startdate:  pgtype.Date{Time: data.StartDate.Value.Time, Valid: true},
		Enddate:    pgtype.Date{Time: data.EndDate.Value.Time, Valid: true},
		Amount:     int32(data.Amount.Value),
		Raiseddate: pgtype.Date{Time: data.RaisedDate.Value.Time, Valid: true},
		Source:     pgtype.Text{String: "Created manually", Valid: true},
		//TODO make sure we have correct createdby ID in ticket PFS-136
		CreatedBy: pgtype.Int4{Int32: int32(1), Valid: true},
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

	reductionForDateParams := store.GetFeeReductionForDateParams{
		ClientID:     int32(clientId),
		Datereceived: invoice.Raiseddate,
	}

	feeReduction, _ := transaction.GetFeeReductionForDate(ctx, reductionForDateParams)

	if feeReduction.ID != 0 {
		var generalSupervisionFee int32
		if data.SupervisionLevel.Value == "GENERAL" {
			generalSupervisionFee = invoice.Amount
		}
		reduction := calculateFeeReduction(shared.ParseFeeReductionType(feeReduction.Type), invoice.Amount, invoice.Feetype, generalSupervisionFee)
		ledger, allocations := generateLedgerEntries(addLedgerVars{
			amount:             reduction,
			transactionType:    shared.ParseFeeReductionType(feeReduction.Type),
			feeReductionId:     feeReduction.ID,
			clientId:           int32(clientId),
			invoiceId:          invoice.ID,
			outstandingBalance: invoice.Amount,
		})
		ledgerId, err := transaction.CreateLedger(ctx, ledger)
		if err != nil {
			return err
		}

		for _, allocation := range allocations {
			allocation.LedgerID = pgtype.Int4{Int32: ledgerId, Valid: true}
			_, err = transaction.CreateLedgerAllocation(ctx, allocation)
			if err != nil {
				return err
			}
		}
	}

	commitErr := tx.Commit(ctx)
	if commitErr != nil {
		return commitErr
	}

	return s.ReapplyCredit(ctx, int32(clientId))
}

func (s *Service) validateManualInvoice(data shared.AddManualInvoice) []string {
	var validationsErrors []string

	if data.InvoiceType.RequiresDateValidation() {
		if !data.RaisedDate.Value.Time.Before(time.Now()) {
			validationsErrors = append(validationsErrors, "RaisedDateForAnInvoice")
		}
	}

	isStartDateValid := validateStartDate(data.StartDate.Value, data.EndDate.Value)
	if !isStartDateValid {
		validationsErrors = append(validationsErrors, "StartDate")
	}

	isEndDateValid := validateEndDate(data.StartDate.Value, data.EndDate.Value)
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

func validateEndDate(startDate shared.Date, endDate shared.Date) bool {
	return !endDate.Time.Before(startDate.Time)
}

func validateStartDate(startDate shared.Date, endDate shared.Date) bool {
	if startDate.Time.After(endDate.Time) {
		return false
	}

	if !startDate.IsSameFinancialYear(endDate) {
		return false
	}

	return true
}
