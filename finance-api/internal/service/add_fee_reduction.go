package service

import (
	"context"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/apierror"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/auth"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/store"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
	"slices"
	"strconv"
	"time"
)

func (s *Service) AddFeeReduction(ctx context.Context, clientId int, data shared.AddFeeReduction) error {
	overlapParams := store.CountOverlappingFeeReductionParams{
		ClientID:   int32(clientId),
		Overlaps:   calculateFeeReductionStartDate(data.StartYear),
		Overlaps_2: calculateFeeReductionEndDate(data.StartYear, data.LengthOfAward),
	}

	hasFeeReduction, _ := s.store.CountOverlappingFeeReduction(ctx, overlapParams)
	if hasFeeReduction != 0 {
		return apierror.BadRequest{Reason: "overlap"}
	}

	feeReductionParams := store.AddFeeReductionParams{
		ClientID:     int32(clientId),
		Type:         data.FeeType.Key(),
		Startdate:    calculateFeeReductionStartDate(data.StartYear),
		Enddate:      calculateFeeReductionEndDate(data.StartYear, data.LengthOfAward),
		Notes:        data.Notes,
		Datereceived: pgtype.Date{Time: data.DateReceived.Time, Valid: true},
		CreatedBy:    pgtype.Int4{Int32: int32(ctx.(auth.Context).User.ID), Valid: true},
	}

	ctx, cancelTx := s.WithCancel(ctx)
	defer cancelTx()

	tx, err := s.BeginStoreTx(ctx)
	if err != nil {
		return err
	}

	feeReduction, err := tx.AddFeeReduction(ctx, feeReductionParams)
	if err != nil {
		return err
	}

	invoices, err := tx.GetInvoiceBalancesForFeeReductionRange(ctx, feeReduction.ID)
	if err != nil {
		return err
	}

	for _, invoice := range invoices {
		amount := calculateFeeReduction(shared.ParseFeeReductionType(feeReduction.Type), invoice.Amount, invoice.Feetype, int32(invoice.GeneralSupervisionFee))
		ledger, allocations := generateLedgerEntries(ctx, addLedgerVars{
			amount:             amount,
			transactionType:    data.FeeType,
			feeReductionId:     feeReduction.ID,
			clientId:           int32(clientId),
			invoiceId:          invoice.ID,
			outstandingBalance: invoice.Outstanding,
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

	err = tx.Commit(ctx)
	if err != nil {
		return err
	}

	return s.ReapplyCredit(ctx, int32(clientId))
}

func calculateFeeReduction(feeReductionType shared.FeeReductionType, invoiceTotal int32, invoiceFeeType string, generalSupervisionFee int32) int32 {
	if feeReductionType != shared.FeeReductionTypeRemission {
		return invoiceTotal
	}

	nonGeneralRemissionFeeTypes := []string{"AD", "GA", "GT", "GS"}
	if slices.Contains(nonGeneralRemissionFeeTypes, invoiceFeeType) {
		return invoiceTotal / 2
	}

	return generalSupervisionFee / 2
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
