package service

import (
	"context"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/apierror"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/store"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
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
	case shared.InvoiceTypeGA:
		data.Amount = shared.Nillable[int]{Value: 20000, Valid: true}
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
	validationErrors := s.validateManualInvoice(data)

	if len(validationErrors) != 0 {
		return apierror.ValidationError{Errors: validationErrors}
	}

	ctx, cancelTx := context.WithCancel(ctx)
	defer cancelTx()

	tx, err := s.BeginStoreTx(ctx)
	if err != nil {
		return err
	}

	counter, err := tx.GetInvoiceCounter(ctx, strconv.Itoa(data.StartDate.Value.Time.Year())+"InvoiceNumber")
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
	invoice, err = tx.AddInvoice(ctx, invoiceParams)
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

		err = tx.AddInvoiceRange(ctx, addInvoiceRangeQueryArgs)
		if err != nil {
			return err
		}
	}

	reductionForDateParams := store.GetFeeReductionForDateParams{
		ClientID:     int32(clientId),
		Datereceived: invoice.Raiseddate,
	}

	feeReduction, _ := tx.GetFeeReductionForDate(ctx, reductionForDateParams)

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
		ledgerId, err := tx.CreateLedger(ctx, ledger)
		if err != nil {
			return err
		}

		for _, allocation := range allocations {
			allocation.LedgerID = pgtype.Int4{Int32: ledgerId, Valid: true}
			err = tx.CreateLedgerAllocation(ctx, allocation)
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

func (s *Service) validateManualInvoice(data shared.AddManualInvoice) apierror.ValidationErrors {
	validationErrors := apierror.ValidationErrors{}

	if data.InvoiceType.RequiresDateValidation() {
		if !data.RaisedDate.Value.Time.Before(time.Now()) {
			validationErrors["RaisedDate"] = map[string]string{"RaisedDate": "Raised BankDate not in the past"}
		}
	}

	if data.InvoiceType.RequiresSameFinancialYearValidation() {
		isSameFinancialYear := validateSameFinancialYear(data.StartDate.Value, data.EndDate.Value)
		if !isSameFinancialYear {
			validationErrors["StartDate"] = map[string]string{"StartDate": "Start BankDate and end BankDate must be in same financial year"}
		}
	}

	isStartDateValid := validateStartDate(data.StartDate.Value, data.EndDate.Value)
	if !isStartDateValid {
		validationErrors["StartDate"] = map[string]string{"StartDate": "Start BankDate must be before end BankDate"}
	}

	isEndDateValid := validateEndDate(data.StartDate.Value, data.EndDate.Value)
	if !isEndDateValid {
		validationErrors["EndDate"] = map[string]string{"EndDate": "End BankDate must be after start BankDate"}
	}

	return validationErrors
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
	return !startDate.Time.After(endDate.Time)
}

func validateSameFinancialYear(startDate shared.Date, endDate shared.Date) bool {
	return startDate.IsSameFinancialYear(endDate)
}
