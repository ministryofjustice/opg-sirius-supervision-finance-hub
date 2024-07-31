package service

import (
	"context"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/opg-sirius-finance-hub/finance-api/internal/store"
	"github.com/opg-sirius-finance-hub/shared"
	"strconv"
	"time"
)

func (s *Service) AddFeeReduction(ctx context.Context, id int, data shared.AddFeeReduction) error {
	countOverlappingFeeReductionParams := store.CountOverlappingFeeReductionParams{
		ClientID:   int32(id),
		Overlaps:   calculateFeeReductionStartDate(data.StartYear),
		Overlaps_2: calculateFeeReductionEndDate(data.StartYear, data.LengthOfAward),
	}

	hasFeeReduction, _ := s.store.CountOverlappingFeeReduction(ctx, countOverlappingFeeReductionParams)
	if hasFeeReduction != 0 {
		return BadRequest{Reason: "overlap"}
	}

	addFeeReductionQueryArgs := store.AddFeeReductionParams{
		ClientID:     int32(id),
		Type:         data.FeeType.Key(),
		Startdate:    calculateFeeReductionStartDate(data.StartYear),
		Enddate:      calculateFeeReductionEndDate(data.StartYear, data.LengthOfAward),
		Notes:        data.Notes,
		Datereceived: pgtype.Date{Time: data.DateReceived.Time, Valid: true},
		//TODO make sure we have correct createdby ID in ticket PFS-88
		CreatedBy: pgtype.Int4{Int32: 1, Valid: true},
	}

	ctx, cancelTx := context.WithCancel(ctx)
	defer cancelTx()

	tx, err := s.tx.Begin(ctx)
	if err != nil {
		return err
	}

	transaction := s.store.WithTx(tx)

	var feeReduction store.FeeReduction
	feeReduction, err = transaction.AddFeeReduction(ctx, addFeeReductionQueryArgs)
	if err != nil {
		return err
	}

	var invoices []store.Invoice
	invoices, err = transaction.GetInvoicesValidForFeeReduction(ctx, feeReduction.ID)
	if err != nil {
		return err
	}

	for _, invoice := range invoices {
		err = s.AddLedgerAndAllocations(data.FeeType, feeReduction.ID, feeReduction.FinanceClientID.Int32, invoice, transaction, ctx)
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
