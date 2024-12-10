package service

import (
	"context"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/apierror"
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

	feeReduction, err := transaction.AddFeeReduction(ctx, feeReductionParams)
	if err != nil {
		return err
	}

	invoices, err := transaction.GetInvoiceBalancesForFeeReductionRange(ctx, feeReduction.ID)
	if err != nil {
		return err
	}

	for _, invoice := range invoices {
		amount := calculateFeeReduction(shared.ParseFeeReductionType(feeReduction.Type), invoice.Amount, invoice.Feetype, invoice.GeneralSupervisionFee.Int32)
		ledger, allocations := generateLedgerEntries(addLedgerVars{
			amount:             amount,
			transactionType:    data.FeeType,
			feeReductionId:     feeReduction.ID,
			clientId:           int32(clientId),
			invoiceId:          invoice.ID,
			outstandingBalance: invoice.Outstanding,
		})
		ledgerId, err := transaction.CreateLedger(ctx, ledger)
		if err != nil {
			return err
		}

		for _, allocation := range allocations {
			allocation.LedgerID = pgtype.Int4{Int32: ledgerId, Valid: true}
			err = transaction.CreateLedgerAllocation(ctx, allocation)
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
	var reduction int32
	if feeReductionType == shared.FeeReductionTypeRemission {
		nonGeneralRemissionFeeTypes := []string{"AD", "GA", "GT", "GS"}
		if slices.Contains(nonGeneralRemissionFeeTypes, invoiceFeeType) {
			reduction = invoiceTotal / 2
		} else {
			if generalSupervisionFee > 0 {
				reduction = generalSupervisionFee / 2
			}
		}
	} else {
		reduction = invoiceTotal
	}
	return reduction
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
