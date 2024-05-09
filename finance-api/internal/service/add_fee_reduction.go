package service

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/opg-sirius-finance-hub/finance-api/internal/store"
	"github.com/opg-sirius-finance-hub/shared"
	"log"
	"strconv"
	"strings"
	"time"
)

func (s *Service) AddFeeReduction(id int, data shared.AddFeeReduction) error {
	ctx := context.Background()

	countOverlappingFeeReductionParams := store.CountOverlappingFeeReductionParams{
		ClientID:  int32(id),
		Startdate: calculateFeeReductionStartDate(data.StartYear),
		Enddate:   calculateFeeReductionEndDate(data.StartYear, data.LengthOfAward),
	}

	hasFeeReduction, _ := s.Store.CountOverlappingFeeReduction(ctx, countOverlappingFeeReductionParams)
	if hasFeeReduction == 1 {
		return errors.New("overlap")
	}

	addFeeReductionQueryArgs := store.AddFeeReductionParams{
		ClientID:     int32(id),
		Type:         strings.ToUpper(data.FeeType),
		Startdate:    calculateFeeReductionStartDate(data.StartYear),
		Enddate:      calculateFeeReductionEndDate(data.StartYear, data.LengthOfAward),
		Notes:        data.Notes,
		Deleted:      false,
		Datereceived: pgtype.Date{Time: data.DateReceived.Time, Valid: true},
	}

	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return err
	}

	defer func() {
		if err := tx.Rollback(ctx); err != nil {
			log.Println("Error rolling back transaction:", err)
		}
	}()

	transaction := s.Store.WithTx(tx)

	var feeReduction store.FeeReduction
	feeReduction, err = transaction.AddFeeReduction(ctx, addFeeReductionQueryArgs)
	if err != nil {
		return err
	}

	var invoices []store.Invoice
	invoices, err = transaction.AddFeeReductionToInvoices(ctx, feeReduction.ID)
	if err != nil {
		return err
	}

	for _, invoice := range invoices {
		var amount int32 = 0
		switch data.FeeType {
		case "exemption", "hardship":
			amount = invoice.Amount
		case "remission":
			invoiceFeeRangeParams := store.GetInvoiceFeeRangerAmountParams{
				InvoiceID:        pgtype.Int4{Int32: invoice.ID, Valid: true},
				Supervisionlevel: "GENERAL",
			}
			amount, _ = transaction.GetInvoiceFeeRangerAmount(ctx, invoiceFeeRangeParams)
		}
		ledgerQueryArgs := store.CreateLedgerForFeeReductionParams{
			Method:          strings.ToUpper(string(data.FeeType[0])) + data.FeeType[1:] + " credit for invoice " + invoice.Reference,
			Amount:          amount,
			Notes:           pgtype.Text{String: "Credit due to approved " + strings.ToUpper(string(data.FeeType[0])) + data.FeeType[1:]},
			Type:            "CREDIT " + strings.ToUpper(data.FeeType),
			FinanceClientID: pgtype.Int4{Int32: invoice.FinanceClientID.Int32, Valid: true},
			FeeReductionID:  pgtype.Int4{Int32: invoice.FeeReductionID.Int32, Valid: true},
			//TODO make sure we have correct createdby ID
			CreatedbyID: pgtype.Int4{Int32: 1},
		}
		var ledger store.Ledger
		ledger, err = transaction.CreateLedgerForFeeReduction(ctx, ledgerQueryArgs)
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

	err = tx.Commit(ctx)
	if err != nil {
		return err
	}

	return nil
}

func calculateFeeReductionEndDate(startYear string, lengthOfAward int) pgtype.Date {
	startYearTransformed, _ := strconv.Atoi(startYear)

	endDate := startYearTransformed + lengthOfAward
	financeYearEnd := "-03-31"
	feeReductionEndDateString := strconv.Itoa(endDate) + financeYearEnd
	feeReductionEndDate, _ := time.Parse("2006-01-02", feeReductionEndDateString)
	return pgtype.Date{Time: feeReductionEndDate, Valid: true}
}

func calculateFeeReductionStartDate(startYear string) pgtype.Date {
	financeYearStart := "-04-01"
	feeReductionStartDateString := startYear + financeYearStart
	feeReductionStartDate, _ := time.Parse("2006-01-02", feeReductionStartDateString)
	return pgtype.Date{Time: feeReductionStartDate, Valid: true}
}
