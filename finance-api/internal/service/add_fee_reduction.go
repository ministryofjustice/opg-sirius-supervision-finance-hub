package service

import (
	"context"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/opg-sirius-finance-hub/finance-api/internal/store"
	"github.com/opg-sirius-finance-hub/shared"
	"strconv"
	"time"
)

func (s *Service) AddFeeReduction(ctx context.Context, clientId int, data shared.AddFeeReduction) error {
	countOverlappingFeeReductionParams := store.CountOverlappingFeeReductionParams{
		ClientID:   int32(clientId),
		Overlaps:   calculateFeeReductionStartDate(data.StartYear),
		Overlaps_2: calculateFeeReductionEndDate(data.StartYear, data.LengthOfAward),
	}

	hasFeeReduction, _ := s.store.CountOverlappingFeeReduction(ctx, countOverlappingFeeReductionParams)
	if hasFeeReduction != 0 {
		return BadRequest{Reason: "overlap"}
	}

	addFeeReductionQueryArgs := store.AddFeeReductionParams{
		ClientID:     int32(clientId),
		Type:         data.FeeType.Key(),
		Startdate:    calculateFeeReductionStartDate(data.StartYear),
		Enddate:      calculateFeeReductionEndDate(data.StartYear, data.LengthOfAward),
		Notes:        data.Notes,
		Datereceived: pgtype.Date{Time: data.DateReceived.Time, Valid: true},
		//TODO make sure we have correct createdby ID in ticket PFS-136
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

	invoices, err := transaction.GetInvoiceBalancesForFeeReductionRange(ctx, feeReduction.ID)
	if err != nil {
		return err
	}

	for _, invoice := range invoices {
		amount := calculateFeeReduction(ctx, transaction, shared.ParseFeeReductionType(feeReduction.Type), invoice.ID, invoice.Amount, invoice.Feetype)
		if amount > 0 {
			err = addLedger(ctx, transaction, amount, data.FeeType, feeReduction.ID, int32(clientId), invoice.ID, invoice.Outstanding)
			if err != nil {
				return err
			}
		}
	}

	err = tx.Commit(ctx)
	if err != nil {
		return err
	}

	return nil
}

func calculateFeeReduction(ctx context.Context, tStore *store.Queries, feeReductionType shared.FeeReductionType, invoiceId int32, invoiceTotal int32, invoiceFeeType string) int32 {
	var amount int32
	switch feeReductionType {
	case shared.FeeReductionTypeExemption, shared.FeeReductionTypeHardship:
		amount = invoiceTotal
	case shared.FeeReductionTypeRemission:
		if invoiceFeeType == "AD" {
			amount = invoiceTotal / 2
		} else {
			invoiceFeeRangeParams := store.GetInvoiceFeeRangeAmountParams{
				InvoiceID:        pgtype.Int4{Int32: invoiceId, Valid: true},
				Supervisionlevel: "GENERAL",
			}
			amount, _ = tStore.GetInvoiceFeeRangeAmount(ctx, invoiceFeeRangeParams)
		}
	default:
		amount = 0
	}
	return amount
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
