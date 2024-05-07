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

func (s *Service) AddFeeReduction(body shared.AddFeeReduction) (shared.ValidationError, error) {
	ctx := context.Background()

	checkForOverlappingFeeReductionParams := store.CheckForOverlappingFeeReductionParams{
		ClientID:  int32(body.ClientId),
		Startdate: calculateStartDate(body.StartYear),
		Enddate:   calculateEndDate(body.StartYear, body.LengthOfAward),
	}

	hasFeeReduction, err := s.Store.CheckForOverlappingFeeReduction(ctx, checkForOverlappingFeeReductionParams)
	if hasFeeReduction == 1 {
		var validationError shared.ValidationError
		var validationErrors = make(shared.ValidationErrors)
		if validationErrors["Overlap"] == nil {
			validationErrors["Overlap"] = make(map[string]string)
		}
		validationErrors["Overlap"]["StartOrEndDate"] = "A fee reduction already exists for the period specified"
		validationError.Message = "This has not worked"
		validationError.Errors = validationErrors
		return validationError, err
	}

	validationError := s.VError.ValidateStruct(body)

	if len(validationError.Errors) > 0 {
		return validationError, nil
	}

	addFeeReductionQueryArgs := store.AddFeeReductionParams{
		ClientID:     int32(body.ClientId),
		Type:         strings.ToUpper(body.FeeType),
		Evidencetype: pgtype.Text{},
		Startdate:    calculateStartDate(body.StartYear),
		Enddate:      calculateEndDate(body.StartYear, body.LengthOfAward),
		Notes:        body.FeeReductionNotes,
		Deleted:      false,
		Datereceived: pgtype.Date{Time: body.DateReceive.Time, Valid: true},
	}

	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return shared.ValidationError{}, err
	}

	defer tx.Rollback(ctx)

	var feeReduction store.FeeReduction
	_, err = s.Store.AddFeeReduction(ctx, addFeeReductionQueryArgs)
	if err != nil {
		return shared.ValidationError{}, err
	}

	var invoices []store.Invoice
	invoices, err = s.Store.AddFeeReductionToInvoice(ctx, feeReduction.ID)
	if err != nil {
		return shared.ValidationError{}, err
	}

	for _, invoice := range invoices {
		var amount int32 = 0
		switch body.FeeType {
		case "exemption", "hardship":
			amount = invoice.Amount
		case "remission":
			invoiceFeeRangeParams := store.GetInvoiceFeeRangersAmountParams{
				InvoiceID:        pgtype.Int4{Int32: invoice.ID, Valid: true},
				Supervisionlevel: "GENERAL",
			}
			amount, err = s.Store.GetInvoiceFeeRangersAmount(ctx, invoiceFeeRangeParams)
		}
		ledgerQueryArgs := store.CreateLedgerForFeeReductionParams{
			Method:          strings.ToUpper(string(body.FeeType[0])) + body.FeeType[1:] + " credit for invoice " + invoice.Reference,
			Amount:          amount,
			Notes:           pgtype.Text{String: "Credit due to approved " + strings.ToUpper(string(body.FeeType[0])) + body.FeeType[1:]},
			Type:            "CREDIT " + strings.ToUpper(body.FeeType),
			FinanceClientID: pgtype.Int4{Int32: invoice.FinanceClientID.Int32, Valid: true},
			FeeReductionID:  pgtype.Int4{Int32: invoice.FeeReductionID.Int32, Valid: true},
			CreatedbyID:     pgtype.Int4{Int32: 1},
		}
		var ledger store.Ledger
		ledger, err = s.Store.CreateLedgerForFeeReduction(ctx, ledgerQueryArgs)
		if err != nil {
			return shared.ValidationError{}, err
		}

		queryArgs := store.CreateLedgerAllocationForFeeReductionParams{
			LedgerID:  pgtype.Int4{Int32: ledger.ID, Valid: true},
			InvoiceID: pgtype.Int4{Int32: invoice.ID, Valid: true},
			Amount:    ledger.Amount,
		}

		_, err = s.Store.CreateLedgerAllocationForFeeReduction(ctx, queryArgs)
		if err != nil {
			return shared.ValidationError{}, err
		}
	}

	err = tx.Commit(ctx)
	if err != nil {
		return shared.ValidationError{}, err
	}

	return shared.ValidationError{}, nil
}

func calculateEndDate(startYear string, lengthOfAward string) pgtype.Date {
	startYearTransformed, _ := strconv.Atoi(startYear)
	lengthOfAwardTransformed, _ := strconv.Atoi(lengthOfAward)

	endDate := startYearTransformed + lengthOfAwardTransformed
	financeYearEnd := "-03-31"
	feeReductionEndDateString := strconv.Itoa(endDate) + financeYearEnd
	feeReductionEndDate, _ := time.Parse("2006-01-02", feeReductionEndDateString)
	return pgtype.Date{Time: feeReductionEndDate, Valid: true}
}

func calculateStartDate(startYear string) pgtype.Date {
	financeYearStart := "-04-01"
	feeReductionStartDateString := startYear + financeYearStart
	feeReductionStartDate, _ := time.Parse("2006-01-02", feeReductionStartDateString)
	return pgtype.Date{Time: feeReductionStartDate, Valid: true}
}
