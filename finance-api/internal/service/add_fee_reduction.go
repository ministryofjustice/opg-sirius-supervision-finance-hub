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

func (s *Service) AddFeeReduction(id int, data shared.AddFeeReduction) error {
	ctx := context.Background()

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
		Type:         strings.ToUpper(data.FeeType),
		Startdate:    calculateFeeReductionStartDate(data.StartYear),
		Enddate:      calculateFeeReductionEndDate(data.StartYear, data.LengthOfAward),
		Notes:        data.Notes,
		Deleted:      false,
		Datereceived: pgtype.Date{Time: data.DateReceived.Time, Valid: true},
	}

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
		err2 := s.AddLedgerAndAllocations(data.FeeType, feeReduction.ID, feeReduction.FinanceClientID.Int32, invoice, transaction, ctx)
		if err2 != nil {
			return err2
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
