package service

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/apierror"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/auth"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/store"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
)

func processInvoiceData(data shared.AddManualInvoice) shared.AddManualInvoice {
	switch data.InvoiceType {
	case shared.InvoiceTypeAD:
		data.Amount = shared.Nillable[int32]{Value: 10000, Valid: true}
		data.StartDate = data.RaisedDate
		data.EndDate = data.RaisedDate
	case shared.InvoiceTypeGA:
		data.Amount = shared.Nillable[int32]{Value: 20000, Valid: true}
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

func (s *Service) AddManualInvoice(ctx context.Context, clientId int32, data shared.AddManualInvoice) error {
	data = processInvoiceData(data)
	validationErrors := s.validateManualInvoice(data)
	if len(validationErrors) != 0 {
		return apierror.ValidationError{Errors: validationErrors}
	}

	tx, err := s.BeginStoreTx(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	counter, err := tx.GetInvoiceCounter(ctx, strconv.Itoa(data.StartDate.Value.Time.Year())+"InvoiceNumber")
	if err != nil {
		s.Logger(ctx).Error("Get invoice counter has an issue " + err.Error())
		return err
	}

	invoiceRef := addLeadingZeros(counter)

	var (
		personID   pgtype.Int4
		startDate  pgtype.Date
		endDate    pgtype.Date
		raisedDate pgtype.Date
		source     pgtype.Text
		createdBy  pgtype.Int4
	)

	_ = store.ToInt4(&personID, clientId)
	_ = startDate.Scan(data.StartDate.Value.Time)
	_ = endDate.Scan(data.EndDate.Value.Time)
	_ = raisedDate.Scan(data.RaisedDate.Value.Time)
	_ = source.Scan("Created manually")
	_ = store.ToInt4(&createdBy, ctx.(auth.Context).User.ID)

	calculateFinanceYear := data.StartDate.Value.CalculateFinanceYear()

	invoiceParams := store.AddInvoiceParams{
		PersonID:   personID,
		Feetype:    data.InvoiceType.Key(),
		Reference:  data.InvoiceType.Key() + invoiceRef + "/" + calculateFinanceYear,
		Startdate:  startDate,
		Enddate:    endDate,
		Amount:     data.Amount.Value,
		Raiseddate: raisedDate,
		Source:     source,
		CreatedBy:  createdBy,
	}

	var invoice store.Invoice
	invoice, err = tx.AddInvoice(ctx, invoiceParams)
	if err != nil {
		s.Logger(ctx).Error("Add invoice has an issue " + err.Error() + fmt.Sprintf("for client %d", clientId))
		return err
	}

	var (
		invoiceID pgtype.Int4
		fromDate  pgtype.Date
		toDate    pgtype.Date
	)

	_ = store.ToInt4(&invoiceID, invoice.ID)
	_ = fromDate.Scan(data.StartDate.Value.Time)
	_ = toDate.Scan(data.EndDate.Value.Time)

	if data.SupervisionLevel.Valid {
		addInvoiceRangeQueryArgs := store.AddInvoiceRangeParams{
			InvoiceID:        invoiceID,
			Supervisionlevel: strings.ToUpper(data.SupervisionLevel.Value),
			Fromdate:         fromDate,
			Todate:           toDate,
			Amount:           invoice.Amount,
		}

		err = tx.AddInvoiceRange(ctx, addInvoiceRangeQueryArgs)
		if err != nil {
			s.Logger(ctx).Error(fmt.Sprintf("Error in add invoice range for invoice %d for client %d", invoiceID.Int32, clientId), slog.String("err", err.Error()))
			return err
		}
	}

	reductionForDateParams := store.GetFeeReductionForDateParams{
		ClientID:     clientId,
		DateReceived: invoice.Raiseddate,
	}

	feeReduction, _ := tx.GetFeeReductionForDate(ctx, reductionForDateParams)

	if feeReduction.ID != 0 {
		var generalSupervisionFee int32
		if data.SupervisionLevel.Value == "GENERAL" {
			generalSupervisionFee = invoice.Amount
		}
		reduction := calculateFeeReduction(shared.ParseFeeReductionType(feeReduction.Type), invoice.Amount, invoice.Feetype, generalSupervisionFee)
		ledger, allocations := generateLedgerEntries(ctx, addLedgerVars{
			amount:             reduction,
			transactionType:    shared.ParseFeeReductionType(feeReduction.Type),
			feeReductionId:     feeReduction.ID,
			clientId:           clientId,
			invoiceId:          invoice.ID,
			outstandingBalance: invoice.Amount,
		})
		ledgerID, err := tx.CreateLedger(ctx, ledger)
		if err != nil {
			s.Logger(ctx).Error(fmt.Sprintf("Error in add fee reduction for ledger %d for client %d", ledgerID, clientId), slog.String("err", err.Error()))
			return err
		}

		for _, allocation := range allocations {
			allocation.LedgerID = ledgerID
			err = tx.CreateLedgerAllocation(ctx, allocation)
			if err != nil {
				s.Logger(ctx).Error(fmt.Sprintf("Error in add fee reduction for ledger allocation %d for client %d", ledgerID, clientId), slog.String("err", err.Error()))
				return err
			}
		}
	}

	err = s.ReapplyCredit(ctx, clientId, tx)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (s *Service) validateManualInvoice(data shared.AddManualInvoice) apierror.ValidationErrors {
	validationErrors := apierror.ValidationErrors{}

	if data.InvoiceType.RequiresDateValidation() {
		if !data.RaisedDate.Value.Time.Before(time.Now()) {
			validationErrors["RaisedDate"] = map[string]string{"RaisedDate": "Raised date not in the past"}
		}
	}

	if data.InvoiceType.RequiresSameFinancialYearValidation() {
		isSameFinancialYear := validateSameFinancialYear(data.StartDate.Value, data.EndDate.Value)
		if !isSameFinancialYear {
			validationErrors["StartDate"] = map[string]string{"StartDate": "Start date and end date must be in same financial year"}
		}
	}

	isStartDateValid := validateStartDate(data.StartDate.Value, data.EndDate.Value)
	if !isStartDateValid {
		validationErrors["StartDate"] = map[string]string{"StartDate": "Start date must be before end date"}
	}

	isEndDateValid := validateEndDate(data.StartDate.Value, data.EndDate.Value)
	if !isEndDateValid {
		validationErrors["EndDate"] = map[string]string{"EndDate": "End date must be after start date"}
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
